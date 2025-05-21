package dso

import (
	"github.com/coatyio/dda/services/com/api"
)

type state struct {
	leader bool
	//energyCommunities      []energyCommunity
	//flowProposals []common.FlowProposalsMessage

	topology            topology
	registerCallbacks   map[string]api.ActionCallback
	deregisterCallbacks map[string]api.ActionCallback
}

/*type energyCommunity struct {
	id                        string
	acknowledgeToplogyVersion int
}*/

type topology struct {
	Version int
	Sensors map[string][]string //sensorId -> childSensorIds
}
