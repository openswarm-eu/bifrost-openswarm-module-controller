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
	config       Config
	ddaConnector *dda.Connector
	state        *state

	writeEnergyCommunityToLogCallback func(id string, version int)

	ctx context.Context
}

func newEnergyCommunityTopologyUpdater(config Config, ddaConnector *dda.Connector, state *state, writeEnergyCommunityToLogCallback func(id string, version int)) *energyCommunityTopologyUpdater {
	return &energyCommunityTopologyUpdater{
		config:                            config,
		ddaConnector:                      ddaConnector,
		state:                             state,
		writeEnergyCommunityToLogCallback: writeEnergyCommunityToLogCallback,
	}
}

func (tu *energyCommunityTopologyUpdater) setContext(ctx context.Context) {
	tu.ctx = ctx
}

func (tu *energyCommunityTopologyUpdater) sendUpdatesToEnergyCommunities() {
	energyCommunityWithOldTopology := make([]string, 0)
	for energyCommunityId, topologyVersion := range tu.state.energyCommunities {
		if topologyVersion != tu.state.topology.Version {
			energyCommunityWithOldTopology = append(energyCommunityWithOldTopology, energyCommunityId)
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

	message := common.TopologyMessage{Topology: topology, Timestamp: time.Now()}
	data, _ := json.Marshal(message)

	updateSuccessChannel := make(chan struct{}, outstandingTopologyUpdates)
	ctx, cancel := context.WithTimeout(
		tu.ctx,
		time.Duration(tu.config.TopologyUpdateAcknowledgementTimeout))
	for _, energyCommunity := range energyCommunityWithOldTopology {
		go tu.sendUpdateToEnergyCommunity(energyCommunity, data, ctx, updateSuccessChannel)
	}

	for {
		select {
		case <-ctx.Done():
			cancel()
			return
		case <-updateSuccessChannel:
			outstandingTopologyUpdates--
			if outstandingTopologyUpdates == 0 {
				cancel()
				return
			}
		}
	}
}

func (tu *energyCommunityTopologyUpdater) sendUpdateToEnergyCommunity(energyCommunityId string, data []byte, ctx context.Context, sucessChannel chan struct{}) {
	log.Printf("dso - sending topology update to: %s", energyCommunityId)
	result, err := tu.ddaConnector.PublishAction(ctx, api.Action{Type: common.AppendId(common.TOPOLOGY_UPDATE_ACTION, energyCommunityId), Id: uuid.NewString(), Source: "dso", Params: data})
	if err != nil {
		log.Printf("Could not send topology update to %s - %s", energyCommunityId, err)
	}

	response := <-result
	if len(response.Data) == 0 {
		return
	}

	log.Printf("dso - %s: accepted new topo", energyCommunityId)
	tu.state.energyCommunities[energyCommunityId] = tu.state.topology.Version
	tu.writeEnergyCommunityToLogCallback(energyCommunityId, tu.state.topology.Version)
	sucessChannel <- struct{}{}
}
