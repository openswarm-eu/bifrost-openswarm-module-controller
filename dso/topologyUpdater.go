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
	ctx          context.Context
	cancel       context.CancelFunc

	writeEnergyCommunityToLog         bool
	writeEnergyCommunityToLogCallback func(topology topology)
}

func newEnergyCommunityTopologyUpdater(ddaConnector *dda.Connector, state *state, writeEnergyCommunityToLogCallback func(topology topology)) *energyCommunityTopologyUpdater {
	return &energyCommunityTopologyUpdater{
		ddaConnector:                      ddaConnector,
		state:                             state,
		writeEnergyCommunityToLogCallback: writeEnergyCommunityToLogCallback,
	}
}

func (tu *energyCommunityTopologyUpdater) sendUpdatesToEnergyCommunities() {
	tu.writeEnergyCommunityToLog = false

	go func() {
		for {
			energyCommunityWithOldTopology := make([]*energyCommunity, 0)
			for _, energyCommunity := range tu.state.energyCommunities {
				if energyCommunity.TopologyVersion != tu.state.topology.Version {
					energyCommunityWithOldTopology = append(energyCommunityWithOldTopology, energyCommunity)
				}
			}

			if len(energyCommunityWithOldTopology) == 0 {
				return
			}

			if tu.cancel != nil {
				tu.cancel()
			}

			tu.ctx, tu.cancel = context.WithTimeout(
				context.Background(),
				time.Duration(1*time.Second))

			topology := make(map[string][]string)
			for sensorId, sensor := range tu.state.topology.Sensors {
				topology[sensorId] = sensor.childrenSensorId
			}

			message := common.TopologyMessage{Topology: topology, Timestamp: time.Now().Unix()}
			data, _ := json.Marshal(message)

			for _, energyCommunity := range energyCommunityWithOldTopology {
				go tu.sendUpdateToEnergyCommunity(energyCommunity, data, tu.ctx)
			}

			<-tu.ctx.Done()

			if tu.ctx.Err() == context.Canceled {
				return
			}

			if tu.writeEnergyCommunityToLog {
				tu.writeEnergyCommunityToLogCallback(tu.state.topology)
			}

			time.Sleep(4 * time.Second)
		}
	}()
}

func (tu *energyCommunityTopologyUpdater) sendUpdateToEnergyCommunity(energyCommunity *energyCommunity, data []byte, ctx context.Context) {
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
	tu.writeEnergyCommunityToLog = true
}

func (tu *energyCommunityTopologyUpdater) stop() {
	if tu.cancel != nil {
		tu.cancel()
	}
}
