package dso

import (
	"context"
	"encoding/json"
	"log"

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

	registerEnergyCommunityChannel, err := c.ddaConnector.SubscribeAction(ctx, api.SubscriptionFilter{Type: common.REGISTER_ENERGY_COMMUNITY_ACTION})
	if err != nil {
		return err
	}

	deregisterEnergyCommunityChannel, err := c.ddaConnector.SubscribeAction(ctx, api.SubscriptionFilter{Type: common.DEREGISTER_ENERGY_COMMUNITY_ACTION})
	if err != nil {
		return err
	}

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
			case registerEnergyCommunity := <-registerEnergyCommunityChannel:
				if !c.state.leader {
					continue
				}

				log.Println("dso - got register energy community")
				var msg common.DdaRegisterEnergyCommunityMessage
				if err := json.Unmarshal(registerEnergyCommunity.Params, &msg); err != nil {
					log.Printf("Could not unmarshal incoming register energy community message, %s", err)
					continue
				}

				for _, energyCommunity := range c.state.energyCommunities {
					if energyCommunity.Id == msg.EnergyCommunityId {
						log.Printf("Energy community %s already registered", msg.EnergyCommunityId)
						registerEnergyCommunity.Callback(api.ActionResult{Data: []byte(msg.EnergyCommunityId)})
						continue
					}
				}
				c.state.registerEnergyCommunityCallbacks[msg.EnergyCommunityId] = registerEnergyCommunity.Callback

				newEnergyCommunities := c.state.cloneEnergyCommunities()
				newEnergyCommunities = append(newEnergyCommunities, energyCommunity{Id: msg.EnergyCommunityId, TopologyVersion: 0})
				c.writeEnergyCommunityToLog(newEnergyCommunities)
			case derigsterEnergyCommunity := <-deregisterEnergyCommunityChannel:
				if !c.state.leader {
					continue
				}

				log.Println("dso - got deregister energy community")
				var msg common.DdaRegisterEnergyCommunityMessage
				if err := json.Unmarshal(derigsterEnergyCommunity.Params, &msg); err != nil {
					log.Printf("Could not unmarshal incoming deregister energy community message, %s", err)
					continue
				}

				c.state.deregisterEnergyCommunityCallbacks[msg.EnergyCommunityId] = derigsterEnergyCommunity.Callback

				newEnergyCommunities := c.state.cloneEnergyCommunities()
				newEnergyCommunities = removeEnergyCommunity(newEnergyCommunities, msg.EnergyCommunityId)
				c.writeEnergyCommunityToLog(newEnergyCommunities)
			case stateChange := <-sc:
				switch stateChange.Key {
				case TOPOLOGY_KEY:
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
				case ENERGY_COMMUNITY_KEY:
					var energyCommunities []energyCommunity
					if err := json.Unmarshal(stateChange.Value, &energyCommunities); err != nil {
						log.Printf("Could not unmarshal incoming energy community state change message, %s", err)
						continue
					}

					c.state.energyCommunities = energyCommunities

					if !c.state.leader {
						continue
					}

					for energyCommunityId, callback := range c.state.registerEnergyCommunityCallbacks {
						for _, energyCommunity := range c.state.energyCommunities {
							if energyCommunity.Id == energyCommunityId {
								callback(api.ActionResult{Data: []byte(energyCommunityId)})
								delete(c.state.registerEnergyCommunityCallbacks, energyCommunityId)
								break
							}
						}
					}
					for energyCommunityId, callback := range c.state.deregisterEnergyCommunityCallbacks {
						for _, energyCommunity := range c.state.energyCommunities {
							if energyCommunity.Id == energyCommunityId {
								break
							}
						}
						callback(api.ActionResult{Data: []byte(energyCommunityId)})
						delete(c.state.deregisterEnergyCommunityCallbacks, energyCommunityId)
					}
				}
			}
		}
	}()
	return nil
}

func (c *connector) writeToplogyToLog(topology topology) error {
	topology.Version++
	data, _ := json.Marshal(topology)

	input := stateAPI.Input{
		Op:    stateAPI.InputOpSet,
		Key:   TOPOLOGY_KEY,
		Value: data,
	}

	return c.ddaConnector.ProposeInput(c.ctx, &input)
}

func (c *connector) writeEnergyCommunityToLog(energyCommunities []energyCommunity) error {
	data, _ := json.Marshal(energyCommunities)

	input := stateAPI.Input{
		Op:    stateAPI.InputOpSet,
		Key:   ENERGY_COMMUNITY_KEY,
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

const TOPOLOGY_KEY = "topology"
const ENERGY_COMMUNITY_KEY = "energycommunity"
