package dso

import (
	"context"
	"log"
	"math"

	"code.siemens.com/energy-community-controller/common"
	"code.siemens.com/energy-community-controller/sct"
)

var eventChannel chan string

func addEvent(event string) {
	eventChannel <- event
}

type logic struct {
	config    Config
	connector *connector
	state     *state
	sct       *sct.SCT
}

func newLogic(config Config, connector *connector, state *state) (*logic, error) {
	l := logic{config: config, connector: connector, state: state}

	/*s1, err := os.Open("resources/simpleController1.xml")
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %v", err)
	}
	defer s1.Close()

	s2, err := os.Open("resources/simpleController2.xml")
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %v", err)
	}
	defer s2.Close()*/

	callbacks := make(map[string]func())
	callbacks["triggerNewRound"] = connector.triggerNewRound
	callbacks["getData"] = connector.getSensorData
	callbacks["calculateLimits"] = l.calculateLimits
	callbacks["sendLimits"] = connector.sendLimits
	/*if sct, err := sct.NewSCT([]io.Reader{s1, s2}, callbacks); err != nil {
		return nil, err
	} else {
		l.sct = sct
	}*/

	return &l, nil
}

func (l *logic) start(ctx context.Context) error {
	eventChannel = make(chan string, 100)
	var ticker common.Ticker

	//l.sct.Start(ctx)

	leaderCh := l.connector.leaderCh(ctx)

	go func() {
		for {
			select {
			case v := <-leaderCh:
				if v {
					log.Println("dso - I'm leader, starting logic")
					l.state.leader = true
					ticker.Start(l.config.Periode, l.newRound)
				} else {
					log.Println("dso - lost leadership, stop logic")
					l.state.leader = false
					ticker.Stop()
				}
			case event := <-eventChannel:
				l.sct.AddEvent(event)
			case <-ctx.Done():
				log.Printf("dso - shutdown leader channel observer")
				ticker.Stop()
				return
			}
		}
	}()

	return nil
}

func (l *logic) newRound() {
	l.state.topology = l.state.newTopology
	//addEvent("newRound")
}

func (l *logic) calculateLimits() {
	l.state.ecLimits = make(map[string]common.EnergyCommunitySensorLimitMessage)

	for _, energyCommunity := range l.state.energyCommunities {
		l.state.ecLimits[energyCommunity.Id] = common.EnergyCommunitySensorLimitMessage{SensorLimits: make(map[string]float64)} //remove empty energy communities in the end
	}

	for sensorId, localSensorInformation := range l.state.localSenorInformations {
		sumFlowProposals := 0.0

		for _, flowProposal := range localSensorInformation.ecFlowProposal {
			sumFlowProposals += flowProposal.Flow
		}

		sensorLimit := l.state.topology.Sensors[sensorId].limit
		availableDemand := sensorLimit - (math.Abs(localSensorInformation.measurement) - math.Abs(localSensorInformation.sumECLimits))
		if availableDemand >= math.Abs(sumFlowProposals) {
			for energyCommunityId, flowProposal := range localSensorInformation.ecFlowProposal {
				l.state.ecLimits[energyCommunityId].SensorLimits[sensorId] = flowProposal.Flow
			}
		} else {
			numComponents := 0
			for _, flowProposal := range localSensorInformation.ecFlowProposal {
				if (sumFlowProposals > 0 && flowProposal.Flow > 0) || (sumFlowProposals < 0 && flowProposal.Flow < 0) {
					numComponents += flowProposal.NumberOfNodes
				}
			}

			for availableDemand > 0 {
				ecLimit := availableDemand / float64(numComponents)
				for energyCommunityId, flowProposal := range localSensorInformation.ecFlowProposal {
					if (sumFlowProposals > 0 && flowProposal.Flow > 0) || (sumFlowProposals < 0 && flowProposal.Flow < 0) {
						openDemand := flowProposal.Flow - l.state.ecLimits[energyCommunityId].SensorLimits[sensorId]*float64(flowProposal.NumberOfNodes)

						if openDemand == 0 {
							continue
						}

						if ecLimit >= openDemand {
							l.state.ecLimits[energyCommunityId].SensorLimits[sensorId] += openDemand
							availableDemand -= openDemand
							numComponents -= flowProposal.NumberOfNodes
						} else {
							metDemand := ecLimit * float64(flowProposal.NumberOfNodes)
							l.state.ecLimits[energyCommunityId].SensorLimits[sensorId] += metDemand
							availableDemand -= metDemand
						}
					} else {
						l.state.ecLimits[energyCommunityId].SensorLimits[sensorId] = flowProposal.Flow
					}
				}
			}
		}
	}

	for energyCommunityId, flowSetPointMessage := range l.state.ecLimits {
		if len(flowSetPointMessage.SensorLimits) == 0 {
			delete(l.state.ecLimits, energyCommunityId)
		}
	}
}
