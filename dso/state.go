package dso

import (
	"code.siemens.com/energy-community-controller/common"
	"github.com/coatyio/dda/services/com/api"
)

type state struct {
	leader            bool
	energyCommunities []energyCommunity
	//flowProposals []common.FlowProposalsMessage

	topology                           topology
	registerCallbacks                  map[string]api.ActionCallback
	deregisterCallbacks                map[string]api.ActionCallback
	registerEnergyCommunityCallbacks   map[string]api.ActionCallback
	deregisterEnergyCommunityCallbacks map[string]api.ActionCallback
}

type energyCommunity struct {
	Id              string
	TopologyVersion int
}

type topology struct {
	Version int
	Sensors map[string][]string //sensorId -> childSensorIds
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

func (s *state) cloneEnergyCommunities() []energyCommunity {
	result := make([]energyCommunity, len(s.energyCommunities))

	copy(result, s.energyCommunities)

	return result
}

func addNodeToTopology(registerSensorMessage common.DdaRegisterSensorMessage, topology *topology) {
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

func removeNodeFromTopology(registerSensorMessage common.DdaRegisterSensorMessage, topology *topology) {
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

func removeEnergyCommunity(newEnergyCommunities []energyCommunity, energyCommunityId string) []energyCommunity {
	for i, energyCommunity := range newEnergyCommunities {
		if energyCommunity.Id == energyCommunityId {
			return append(newEnergyCommunities[:i], newEnergyCommunities[i+1:]...)
		}
	}

	return newEnergyCommunities
}
