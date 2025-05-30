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
