package dso

import (
	"github.com/coatyio/dda/services/com/api"
)

type state struct {
	leader            bool
	energyCommunities map[string]energyCommunity // energyCommunityId -> energyCommunity
	//flowProposals []common.FlowProposalsMessage

	topology                           topology
	registerCallbacks                  map[string]api.ActionCallback
	deregisterCallbacks                map[string]api.ActionCallback
	registerEnergyCommunityCallbacks   map[string]api.ActionCallback
	deregisterEnergyCommunityCallbacks map[string]api.ActionCallback
}

type energyCommunity struct {
	id              string
	topologyVersion int
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
