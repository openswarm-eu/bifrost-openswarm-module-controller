package dso

import (
	"math"

	"code.siemens.com/energy-community-controller/common"
)

type sensorLimitsCalculator struct {
	state *state
}

func newSensorLimitsCalculator(state *state) sensorLimitsCalculator {
	return sensorLimitsCalculator{
		state: state,
	}
}

func (l sensorLimitsCalculator) calculateSensorLimits() {
	l.state.energyCommunitySensorLimits = make(map[string]common.EnergyCommunitySensorLimitMessage)

	for energyCommunityId := range l.state.energyCommunities {
		l.state.energyCommunitySensorLimits[energyCommunityId] = common.EnergyCommunitySensorLimitMessage{SensorLimits: make(map[string]float64)}
		for sensorId := range l.state.topology.Sensors {
			l.state.energyCommunitySensorLimits[energyCommunityId].SensorLimits[sensorId] = 0.0
		}
	}

	for sensorId, sensor := range l.state.topology.Sensors {
		sumFlowProposals := 0.0

		for _, flowProposal := range sensor.ecFlowProposal {
			sumFlowProposals += flowProposal.Flow
		}

		sensorLimit := l.state.topology.Sensors[sensorId].limit
		availableDemand := sensorLimit - (math.Abs(sensor.measurement) - math.Abs(sensor.sumECLimits))
		if availableDemand >= math.Abs(sumFlowProposals) {
			for energyCommunityId, flowProposal := range sensor.ecFlowProposal {
				l.state.energyCommunitySensorLimits[energyCommunityId].SensorLimits[sensorId] = math.Abs(flowProposal.Flow)
			}
		} else {
			numComponents := 0
			for _, flowProposal := range sensor.ecFlowProposal {
				if (sumFlowProposals > 0 && flowProposal.Flow > 0) || (sumFlowProposals < 0 && flowProposal.Flow < 0) {
					numComponents += flowProposal.NumberOfNodes
				} else {
					availableDemand += math.Abs(flowProposal.Flow)
				}
			}

			for availableDemand > 0 {
				ecLimit := availableDemand / float64(numComponents)
				for energyCommunityId, flowProposal := range sensor.ecFlowProposal {
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

	for _, sensor := range l.state.topology.Sensors {
		sensor.sumECLimits = 0.0
	}

	for energyCommunityId, flowSetPointMessage := range l.state.energyCommunitySensorLimits {
		for sensorId, limit := range flowSetPointMessage.SensorLimits {
			if l.state.topology.Sensors[sensorId].ecFlowProposal[energyCommunityId].Flow < 0 {
				l.state.topology.Sensors[sensorId].sumECLimits -= limit
			} else {
				l.state.topology.Sensors[sensorId].sumECLimits += limit
			}
		}
	}
}
