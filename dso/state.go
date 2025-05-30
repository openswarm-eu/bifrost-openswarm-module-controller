package dso

import (
	"code.siemens.com/energy-community-controller/common"
)

type state struct {
	// exchanged via raft
	energyCommunities []*energyCommunity
	topology          topology

	// local only
	leader                      bool
	newTopology                 topology
	localSenorInformations      map[string]*localSenorInformation                   // sensorId -> localSenorInformation
	energyCommunitySensorLimits map[string]common.EnergyCommunitySensorLimitMessage // energyCommunityId -> FlowSetPointsMessage
}

type energyCommunity struct {
	Id              string
	TopologyVersion int
}

type topology struct {
	Version int
	Sensors map[string]*sensor
}

type sensor struct {
	Limit            float64
	ChildrenSensorId []string
}

type localSenorInformation struct {
	measurement    float64
	sumECLimits    float64
	ecFlowProposal map[string]common.FlowProposal // energyCommunityId --> FlowProposal
}

func (s *state) removeEnergyCommunity(energyCommunityId string) {
	for i, energyCommunity := range s.energyCommunities {
		if energyCommunity.Id == energyCommunityId {
			s.energyCommunities = append(s.energyCommunities[:i], s.energyCommunities[i+1:]...)
		}
	}
}

func (s *state) cloneTopology() topology {
	result := topology{
		Version: s.topology.Version,
		Sensors: make(map[string]*sensor),
	}

	for sensorId, s := range s.topology.Sensors {
		result.Sensors[sensorId] = &sensor{Limit: s.Limit, ChildrenSensorId: make([]string, len(s.ChildrenSensorId))}
		copy(result.Sensors[sensorId].ChildrenSensorId, s.ChildrenSensorId)
	}

	return result
}

func (s *state) updateLocalSensorInformation() {
	s.localSenorInformations = make(map[string]*localSenorInformation)

	for sensorId := range s.topology.Sensors {
		if _, ok := s.localSenorInformations[sensorId]; !ok {
			s.localSenorInformations[sensorId] = &localSenorInformation{
				measurement:    0,
				ecFlowProposal: make(map[string]common.FlowProposal),
			}
		}
	}
}

func addNodeToTopology(registerSensorMessage common.RegisterSensorMessage, topology *topology) {
	if _, ok := topology.Sensors[registerSensorMessage.SensorId]; !ok {
		topology.Sensors[registerSensorMessage.SensorId] = &sensor{ChildrenSensorId: make([]string, 0)}
	}

	s := topology.Sensors[registerSensorMessage.SensorId]
	s.Limit = registerSensorMessage.Limit

	if registerSensorMessage.ParentSensorId == "" {
		return
	}

	if _, ok := topology.Sensors[registerSensorMessage.ParentSensorId]; !ok {
		topology.Sensors[registerSensorMessage.ParentSensorId] = &sensor{ChildrenSensorId: make([]string, 0)}
	}

	parentSensor := topology.Sensors[registerSensorMessage.ParentSensorId]
	parentSensor.ChildrenSensorId = append(parentSensor.ChildrenSensorId, registerSensorMessage.SensorId)
	topology.Sensors[registerSensorMessage.ParentSensorId] = parentSensor
}

func removeNodeFromTopology(registerSensorMessage common.RegisterSensorMessage, topology *topology) {
	if _, ok := topology.Sensors[registerSensorMessage.SensorId]; !ok {
		return
	}

	delete(topology.Sensors, registerSensorMessage.SensorId)

	if registerSensorMessage.ParentSensorId == "" {
		return
	}

	if _, ok := topology.Sensors[registerSensorMessage.ParentSensorId]; !ok {
		return
	}

	parentSensor := topology.Sensors[registerSensorMessage.ParentSensorId]
	childrenSensorId := parentSensor.ChildrenSensorId
	for i, childId := range childrenSensorId {
		if childId == registerSensorMessage.SensorId {
			parentSensor.ChildrenSensorId = append(childrenSensorId[:i], childrenSensorId[i+1:]...)
			break
		}
	}
}
