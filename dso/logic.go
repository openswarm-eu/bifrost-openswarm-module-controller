package dso

import (
	"context"
	"fmt"
	"io"
	"log"
	"math"
	"os"
	"time"

	"code.siemens.com/energy-community-controller/common"
	"code.siemens.com/energy-community-controller/sct"
)

var eventChannel chan string

func addEvent(event string) {
	eventChannel <- event
}

type logic struct {
	config                         Config
	connector                      *connector
	energyCommunityTopologyUpdater *energyCommunityTopologyUpdater
	state                          *state
	sct                            *sct.SCT
}

func newLogic(config Config, connector *connector, energyCommunityTopologyUpdater *energyCommunityTopologyUpdater, state *state) (*logic, error) {
	l := logic{
		config:                         config,
		connector:                      connector,
		energyCommunityTopologyUpdater: energyCommunityTopologyUpdater,
		state:                          state,
	}

	s1, err := os.Open("resources/sensor1.xml")
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %v", err)
	}
	defer s1.Close()

	s2, err := os.Open("resources/sensor2.xml")
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %v", err)
	}
	defer s2.Close()

	s3, err := os.Open("resources/sensor3.xml")
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %v", err)
	}
	defer s2.Close()

	callbacks := make(map[string]func())
	callbacks["getFlowProposals"] = connector.getFlowProposals
	callbacks["getSensorMeasurements"] = connector.getSensorMeasurements
	callbacks["calculateSensorLimits"] = l.calculateSensorLimits
	callbacks["sendSensorLimits"] = connector.sendSensorLimits
	if sct, err := sct.NewSCT([]io.Reader{s1, s2, s3}, callbacks); err != nil {
		return nil, err
	} else {
		l.sct = sct
	}

	return &l, nil
}

func (l *logic) start(ctx context.Context) error {
	eventChannel = make(chan string, 100)
	var ticker common.Ticker

	l.sct.Start(ctx)

	leaderCh := l.connector.leaderCh(ctx)

	go func() {
		for {
			select {
			case v := <-leaderCh:
				if v {
					log.Println("dso - I'm leader, starting logic")
					l.state.leader = true
					l.energyCommunityTopologyUpdater.sendUpdatesToEnergyCommunities()
					//ticker.Start(l.config.Periode, l.newRound)
					ticker.Start(10*time.Second, l.newRound)
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
	l.state.updateLocalSensorInformation()
	l.energyCommunityTopologyUpdater.sendUpdatesToEnergyCommunities()

	addEvent("newRound")
}

func (l *logic) calculateSensorLimits() {
	l.state.energyCommunitySensorLimits = make(map[string]common.EnergyCommunitySensorLimitMessage)

	for _, energyCommunity := range l.state.energyCommunities {
		l.state.energyCommunitySensorLimits[energyCommunity.Id] = common.EnergyCommunitySensorLimitMessage{SensorLimits: make(map[string]float64)}
	}

	for sensorId, localSensorInformation := range l.state.localSenorInformations {
		sumFlowProposals := 0.0

		for _, flowProposal := range localSensorInformation.ecFlowProposal {
			sumFlowProposals += flowProposal.Flow
		}

		sensorLimit := l.state.topology.Sensors[sensorId].Limit
		availableDemand := sensorLimit - (math.Abs(localSensorInformation.measurement) - math.Abs(localSensorInformation.sumECLimits))
		if availableDemand >= math.Abs(sumFlowProposals) {
			for energyCommunityId, flowProposal := range localSensorInformation.ecFlowProposal {
				l.state.energyCommunitySensorLimits[energyCommunityId].SensorLimits[sensorId] = math.Abs(flowProposal.Flow)
			}
		} else {
			numComponents := 0
			for _, flowProposal := range localSensorInformation.ecFlowProposal {
				if (sumFlowProposals > 0 && flowProposal.Flow > 0) || (sumFlowProposals < 0 && flowProposal.Flow < 0) {
					numComponents += flowProposal.NumberOfNodes
				} else {
					availableDemand += math.Abs(flowProposal.Flow)
				}
			}

			for availableDemand > 0 {
				ecLimit := availableDemand / float64(numComponents)
				for energyCommunityId, flowProposal := range localSensorInformation.ecFlowProposal {
					if (sumFlowProposals > 0 && flowProposal.Flow > 0) || (sumFlowProposals < 0 && flowProposal.Flow < 0) {
						openDemand := math.Abs(flowProposal.Flow) - l.state.energyCommunitySensorLimits[energyCommunityId].SensorLimits[sensorId]*float64(flowProposal.NumberOfNodes)

						if openDemand == 0 {
							continue
						}

						if ecLimit >= openDemand {
							l.state.energyCommunitySensorLimits[energyCommunityId].SensorLimits[sensorId] += openDemand
							availableDemand -= openDemand
							numComponents -= flowProposal.NumberOfNodes
						} else {
							metDemand := ecLimit * float64(flowProposal.NumberOfNodes)
							l.state.energyCommunitySensorLimits[energyCommunityId].SensorLimits[sensorId] += metDemand
							availableDemand -= metDemand
						}
					} else {
						l.state.energyCommunitySensorLimits[energyCommunityId].SensorLimits[sensorId] = math.Abs(flowProposal.Flow)
					}
				}
			}
		}
	}

	for _, localSensorInformation := range l.state.localSenorInformations {
		localSensorInformation.sumECLimits = 0.0
	}

	for energyCommunityId, flowSetPointMessage := range l.state.energyCommunitySensorLimits {
		for sensorId, limit := range flowSetPointMessage.SensorLimits {
			if l.state.localSenorInformations[sensorId].ecFlowProposal[energyCommunityId].Flow < 0 {
				l.state.localSenorInformations[sensorId].sumECLimits -= limit
			} else {
				l.state.localSenorInformations[sensorId].sumECLimits += limit
			}
		}
	}
}
