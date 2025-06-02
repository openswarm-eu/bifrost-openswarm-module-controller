package dso

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"code.siemens.com/energy-community-controller/common"
	"code.siemens.com/energy-community-controller/dda"
	"github.com/coatyio/dda/services/com/api"
	"github.com/google/uuid"
)

type energyCommunityTopologyUpdater struct {
	ddaConnector *dda.Connector
	state        *state

	writeEnergyCommunityToLogCallback func()

	ctx context.Context
}

func newEnergyCommunityTopologyUpdater(ddaConnector *dda.Connector, state *state, writeEnergyCommunityToLogCallback func()) *energyCommunityTopologyUpdater {
	return &energyCommunityTopologyUpdater{
		ddaConnector:                      ddaConnector,
		state:                             state,
		writeEnergyCommunityToLogCallback: writeEnergyCommunityToLogCallback,
	}
}

func (tu *energyCommunityTopologyUpdater) setContext(ctx context.Context) {
	tu.ctx = ctx
}

func (tu *energyCommunityTopologyUpdater) sendUpdatesToEnergyCommunities() {
	energyCommunityWithOldTopology := make([]*energyCommunity, 0)
	for _, energyCommunity := range tu.state.energyCommunities {
		if energyCommunity.TopologyVersion != tu.state.topology.Version {
			energyCommunityWithOldTopology = append(energyCommunityWithOldTopology, energyCommunity)
		}
	}

	outstandingTopologyUpdates := len(energyCommunityWithOldTopology)
	if outstandingTopologyUpdates == 0 {
		return
	}

	topology := make(map[string][]string)
	for sensorId, sensor := range tu.state.topology.Sensors {
		topology[sensorId] = sensor.ChildrenSensorId
	}

	message := common.TopologyMessage{Topology: topology, Timestamp: time.Now().Unix()}
	data, _ := json.Marshal(message)

	updateSuccessChannel := make(chan struct{}, outstandingTopologyUpdates)
	writeEnergyCommunityToLog := false

	ctx, cancel := context.WithTimeout(
		tu.ctx,
		time.Duration(500*time.Millisecond))
	for _, energyCommunity := range energyCommunityWithOldTopology {
		go tu.sendUpdateToEnergyCommunity(energyCommunity, data, ctx, updateSuccessChannel)
	}

outer:
	for {
		select {
		case <-ctx.Done():
			cancel()
			break outer
		case <-updateSuccessChannel:
			writeEnergyCommunityToLog = true
			outstandingTopologyUpdates--
			if outstandingTopologyUpdates == 0 {
				cancel()
				break outer
			}
		}
	}

	if writeEnergyCommunityToLog {
		tu.writeEnergyCommunityToLogCallback()
	}
}

func (tu *energyCommunityTopologyUpdater) sendUpdateToEnergyCommunity(energyCommunity *energyCommunity, data []byte, ctx context.Context, sucessChannel chan struct{}) {
	log.Printf("dso - sending topology update to: %s", energyCommunity.Id)
	result, err := tu.ddaConnector.PublishAction(ctx, api.Action{Type: common.AppendId(common.TOPOLOGY_UPDATE_ACTION, energyCommunity.Id), Id: uuid.NewString(), Source: "dso", Params: data})
	if err != nil {
		log.Printf("Could not send topology update to %s - %s", energyCommunity.Id, err)
	}

	response := <-result
	if len(response.Data) == 0 {
		return
	}

	log.Printf("dso - %s: accepted new topo", energyCommunity.Id)
	energyCommunity.TopologyVersion = tu.state.topology.Version
	sucessChannel <- struct{}{}
}
