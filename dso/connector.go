package dso

import (
	"context"
	"encoding/json"
	"log"
	"strings"

	"code.siemens.com/energy-community-controller/common"
	"code.siemens.com/energy-community-controller/dda"
	"github.com/coatyio/dda/services/com/api"
	stateAPI "github.com/coatyio/dda/services/state/api"
)

type connector struct {
	ddaConnector *dda.Connector
	state        *state

	ctx context.Context
}

func newConnector(ddaConnector *dda.Connector, state *state) *connector {
	return &connector{
		ddaConnector: ddaConnector,
		state:        state,
	}
}

func (c *connector) start(ctx context.Context) error {
	c.ctx = ctx

	registerSensorChannel, err := c.ddaConnector.SubscribeAction(ctx, api.SubscriptionFilter{Type: common.REGISTER_ACTION})
	if err != nil {
		return err
	}

	deregisterSensorChannel, err := c.ddaConnector.SubscribeAction(ctx, api.SubscriptionFilter{Type: common.DEREGISTER_ACTION})
	if err != nil {
		return err
	}

	/*registerEnergyCommunityChannel, err := c.ddaConnector.SubscribeAction(ctx, api.SubscriptionFilter{Type: common.REGISTER_ENERGY_COMMUNITY_ACTION})
	if err != nil {
		return err
	}*/

	/*deregisterEnergyCommunityChannel, err := c.ddaConnector.SubscribeAction(ctx, api.SubscriptionFilter{Type: common.DERIGSTER_ENERGY_COMMUNITY_ACTION})
	if err != nil {
		return err
	}*/

	/*flowProposalChannel, err := c.ddaConnector.SubscribeEvent(ctx, api.SubscriptionFilter{Type: common.FLOW_PROPOSAL_EVENT})
	if err != nil {
		return err
	}*/

	sc, err := c.ddaConnector.ObserveStateChange(ctx)
	if err != nil {
		return err
	}

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case registerSensor := <-registerSensorChannel:
				if !c.state.leader {
					continue
				}

				log.Println("dso - got register sensor")
				var msg common.DdaRegisterSensorMessage
				if err := json.Unmarshal(registerSensor.Params, &msg); err != nil {
					log.Printf("Could not unmarshal incoming register sensor message, %s", err)
					continue
				}

				c.state.registerCallbacks[msg.SensorId] = registerSensor.Callback
				newTopology := c.state.cloneTopology()
				removeNodeFromTopology(msg, &newTopology)
				addNodeToTopology(msg, &newTopology)
				c.writeToplogyToLog(newTopology)
			case deregisterSensor := <-deregisterSensorChannel:
				if !c.state.leader {
					continue
				}

				log.Println("dso - got deregister sensor")
				var msg common.DdaRegisterSensorMessage
				if err := json.Unmarshal(deregisterSensor.Params, &msg); err != nil {
					log.Printf("Could not unmarshal incoming deregister sensor message, %s", err)
					continue
				}

				c.state.deregisterCallbacks[msg.SensorId] = deregisterSensor.Callback
				newTopology := c.state.cloneTopology()
				removeNodeFromTopology(msg, &newTopology)
				c.writeToplogyToLog(newTopology)
			/*case registerEnergyCommunity := <-registerEnergyCommunityChannel:
			if !c.state.leader {
				continue
			}

			log.Println("dsp - got register energy community")
			var msg common.DdaRegisterEnergyCommunityMessage
			if err := json.Unmarshal(registerEnergyCommunity.Params, &msg); err != nil {
				log.Printf("Could not unmarshal incoming register energy community message, %s", err)
				continue
			}

			c.state.registerEnergyCommunityCallbacks[msg.EnergyCommunityId] = registerEnergyCommunity.Callback
			c.writeEnergyCommunityToLog(energyCommunity{id: msg.EnergyCommunityId, topologyVersion: 0})*/
			/*case flowProposal := <-flowProposalChannel:
			if c.state.leader {
				var msg common.FlowProposalsMessage
				if err := json.Unmarshal(flowProposal.Data, &msg); err != nil {
					log.Printf("Could not unmarshal incoming flow proposal message, %s", err)
					continue
				}
				c.state.flowProposals = append(c.state.flowProposals, msg)
			}*/
			case stateChange := <-sc:
				if !strings.HasPrefix(stateChange.Key, TOPOLOGY_PREFIX) {
					continue
				}

				var topology topology
				if err := json.Unmarshal(stateChange.Value, &topology); err != nil {
					log.Printf("Could not unmarshal incoming topology state change message, %s", err)
					continue
				}

				c.state.topology = topology

				if !c.state.leader {
					continue
				}

				for sensorId, callback := range c.state.registerCallbacks {
					if _, ok := c.state.topology.Sensors[sensorId]; !ok {
						continue
					}
					callback(api.ActionResult{Data: []byte(sensorId)})
					delete(c.state.registerCallbacks, sensorId)
				}
				for sensorId, callback := range c.state.deregisterCallbacks {
					if _, ok := c.state.topology.Sensors[sensorId]; ok {
						continue
					}
					callback(api.ActionResult{Data: []byte(sensorId)})
					delete(c.state.deregisterCallbacks, sensorId)
				}

				// trigger toplogy update at energy communities
			}
		}
	}()
	return nil
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

	return
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

func (c *connector) writeToplogyToLog(topology topology) error {
	topology.Version++
	data, _ := json.Marshal(topology)

	input := stateAPI.Input{
		Op:    stateAPI.InputOpSet,
		Key:   TOPOLOGY_PREFIX,
		Value: data,
	}

	return c.ddaConnector.ProposeInput(c.ctx, &input)
}

func (c *connector) writeEnergyCommunityToLog(energyCommunities energyCommunity) error {
	data, _ := json.Marshal(energyCommunities)

	input := stateAPI.Input{
		Op:    stateAPI.InputOpSet,
		Key:   ENERGY_COMMUNITY_PREFIX,
		Value: data,
	}

	return c.ddaConnector.ProposeInput(c.ctx, &input)
}

func (c *connector) leaderCh(ctx context.Context) <-chan bool {
	return c.ddaConnector.LeaderCh(ctx)
}

func (d *connector) triggerNewRound() {
	/*for _, energyCommunity := range d.state.energyCommunities {
		if d.state.currentTopologyVersion != energyCommunity.acknowledgeToplogyVersion {
			continue
		}
		if err := d.ddaConnector.PublishEvent(api.Event{Type: common.AppendId(common.NEW_ROUND_EVENT, energyCommunity.id), Source: d.id, Id: uuid.NewString(), Data: []byte("")}); err != nil {
			log.Printf("could not send new round event - %s", err)
		}
	}*/
}

func (c *connector) getSensorData() {

}

func (c *connector) sendLimits() {
}

const TOPOLOGY_PREFIX = "topology_"
const ENERGY_COMMUNITY_PREFIX = "energycommunity_"
