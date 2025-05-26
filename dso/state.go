package dso

import (
	"code.siemens.com/energy-community-controller/common"
)

type state struct {
	leader            bool
	energyCommunities []*energyCommunity
	//flowProposals []common.FlowProposalsMessage

	topology topology
	sensors  map[string]sensor
}

type energyCommunity struct {
	Id              string
	TopologyVersion int
}

type topology struct {
	Version int
	Sensors map[string][]string //sensorId -> childSensorIds
}

type sensor struct {
	id          string
	measurement float64
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
		Sensors: make(map[string][]string),
	}

	for sensorId, childSensorIds := range s.topology.Sensors {
		result.Sensors[sensorId] = make([]string, len(childSensorIds))
		copy(result.Sensors[sensorId], childSensorIds)
	}

	return result
}

func addNodeToTopology(registerSensorMessage common.RegisterSensorMessage, topology *topology) {
	if _, ok := topology.Sensors[registerSensorMessage.SensorId]; !ok {
		topology.Sensors[registerSensorMessage.SensorId] = make([]string, 0)
	}

	if registerSensorMessage.ParentSensorId == "" {
		return
	}

	if _, ok := topology.Sensors[registerSensorMessage.ParentSensorId]; !ok {
		topology.Sensors[registerSensorMessage.ParentSensorId] = make([]string, 0)
	}

	topology.Sensors[registerSensorMessage.ParentSensorId] = append(topology.Sensors[registerSensorMessage.ParentSensorId], registerSensorMessage.SensorId)
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

	childIds := topology.Sensors[registerSensorMessage.ParentSensorId]
	for i, childId := range childIds {
		if childId == registerSensorMessage.SensorId {
			topology.Sensors[registerSensorMessage.ParentSensorId] = append(childIds[:i], childIds[i+1:]...)
			break
		}
	}
}
