package dso

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"code.siemens.com/energy-community-controller/common"
	"code.siemens.com/energy-community-controller/dda"
	"github.com/coatyio/dda/services/com/api"
	stateAPI "github.com/coatyio/dda/services/state/api"
	"github.com/google/uuid"
)

type connector struct {
	config       Config
	ddaConnector *dda.Connector
	state        *state

	energyCommunityTopologyUpdater     *energyCommunityTopologyUpdater
	registerCallbacks                  map[string]api.ActionCallback
	deregisterCallbacks                map[string]api.ActionCallback
	registerEnergyCommunityCallbacks   map[string]api.ActionCallback
	deregisterEnergyCommunityCallbacks map[string]api.ActionCallback

	ctx context.Context
}

func newConnector(config Config, ddaConnector *dda.Connector, state *state) *connector {
	c := connector{
		config:                             config,
		ddaConnector:                       ddaConnector,
		state:                              state,
		registerCallbacks:                  make(map[string]api.ActionCallback),
		deregisterCallbacks:                make(map[string]api.ActionCallback),
		registerEnergyCommunityCallbacks:   make(map[string]api.ActionCallback),
		deregisterEnergyCommunityCallbacks: make(map[string]api.ActionCallback),
	}

	c.energyCommunityTopologyUpdater = newEnergyCommunityTopologyUpdater(ddaConnector, state, c.writeToplogyToLog)

	return &c

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

	leaderCh := c.leaderCh(ctx)

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case v := <-leaderCh:
				if v {
					c.energyCommunityTopologyUpdater.sendUpdatesToEnergyCommunities()
				} else {
					c.energyCommunityTopologyUpdater.stop()
				}
			case registerSensor := <-registerSensorChannel:
				if !c.state.leader {
					continue
				}

				log.Println("dso - got register sensor")
				var msg common.RegisterSensorMessage
				if err := json.Unmarshal(registerSensor.Params, &msg); err != nil {
					log.Printf("Could not unmarshal incoming register sensor message, %s", err)
					continue
				}

				c.registerCallbacks[msg.SensorId] = registerSensor.Callback
				newTopology := c.state.cloneTopology()
				newTopology.Version++
				removeNodeFromTopology(msg, &newTopology)
				addNodeToTopology(msg, &newTopology)
				c.writeToplogyToLog(newTopology)
			case deregisterSensor := <-deregisterSensorChannel:
				if !c.state.leader {
					continue
				}

				log.Println("dso - got deregister sensor")
				var msg common.RegisterSensorMessage
				if err := json.Unmarshal(deregisterSensor.Params, &msg); err != nil {
					log.Printf("Could not unmarshal incoming deregister sensor message, %s", err)
					continue
				}

				c.deregisterCallbacks[msg.SensorId] = deregisterSensor.Callback
				newTopology := c.state.cloneTopology()
				newTopology.Version++
				removeNodeFromTopology(msg, &newTopology)
				c.writeToplogyToLog(newTopology)
			case registerEnergyCommunity := <-registerEnergyCommunityChannel:
				if !c.state.leader {
					continue
				}

				log.Println("dso - got register energy community")
				var msg common.RegisterEnergyCommunityMessage
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

				c.registerEnergyCommunityCallbacks[msg.EnergyCommunityId] = registerEnergyCommunity.Callback
				c.state.energyCommunities = append(c.state.energyCommunities, &energyCommunity{Id: msg.EnergyCommunityId, TopologyVersion: 0})
				c.writeEnergyCommunityToLog()
			case derigsterEnergyCommunity := <-deregisterEnergyCommunityChannel:
				if !c.state.leader {
					continue
				}

				log.Println("dso - got deregister energy community")
				var msg common.RegisterEnergyCommunityMessage
				if err := json.Unmarshal(derigsterEnergyCommunity.Params, &msg); err != nil {
					log.Printf("Could not unmarshal incoming deregister energy community message, %s", err)
					continue
				}

				c.deregisterEnergyCommunityCallbacks[msg.EnergyCommunityId] = derigsterEnergyCommunity.Callback
				c.state.removeEnergyCommunity(msg.EnergyCommunityId)
				c.writeEnergyCommunityToLog()
			case stateChange := <-sc:
				switch stateChange.Key {
				case topology_key:
					var topology topology
					if err := json.Unmarshal(stateChange.Value, &topology); err != nil {
						log.Printf("Could not unmarshal incoming topology state change message, %s", err)
						continue
					}

					c.state.newTopology = topology

					if !c.state.leader {
						continue
					}

					for sensorId, callback := range c.registerCallbacks {
						if _, ok := c.state.topology.Sensors[sensorId]; !ok {
							continue
						}
						callback(api.ActionResult{Data: []byte(sensorId)})
						delete(c.registerCallbacks, sensorId)
					}

					for sensorId, callback := range c.deregisterCallbacks {
						if _, ok := c.state.topology.Sensors[sensorId]; ok {
							continue
						}
						callback(api.ActionResult{Data: []byte(sensorId)})
						delete(c.deregisterCallbacks, sensorId)
					}

					c.energyCommunityTopologyUpdater.sendUpdatesToEnergyCommunities()
				case energy_community_key:
					var energyCommunities []energyCommunity
					if err := json.Unmarshal(stateChange.Value, &energyCommunities); err != nil {
						log.Printf("Could not unmarshal incoming energy community state change message, %s", err)
						continue
					}
					ecs := make([]*energyCommunity, len(energyCommunities))
					for i := range energyCommunities {
						ecs[i] = &energyCommunities[i]
					}
					c.state.energyCommunities = ecs
					if !c.state.leader {
						continue
					}
					newEnergyCommunityJoined := false
					for energyCommunityId, callback := range c.registerEnergyCommunityCallbacks {
						for _, energyCommunity := range c.state.energyCommunities {
							if energyCommunity.Id == energyCommunityId {
								newEnergyCommunityJoined = true
								callback(api.ActionResult{Data: []byte(energyCommunityId)})
								delete(c.registerEnergyCommunityCallbacks, energyCommunityId)
								break
							}
						}
					}
					for energyCommunityId, callback := range c.deregisterEnergyCommunityCallbacks {
						for _, energyCommunity := range c.state.energyCommunities {
							if energyCommunity.Id == energyCommunityId {
								break
							}
						}
						callback(api.ActionResult{Data: []byte(energyCommunityId)})
						delete(c.deregisterEnergyCommunityCallbacks, energyCommunityId)
					}
					if newEnergyCommunityJoined {
						c.energyCommunityTopologyUpdater.sendUpdatesToEnergyCommunities()
					}
				}
			}
		}
	}()
	return nil
}

func (c *connector) stop() {
	c.energyCommunityTopologyUpdater.stop()
}

func (c *connector) writeToplogyToLog(topology topology) {
	data, _ := json.Marshal(topology)

	input := stateAPI.Input{
		Op:    stateAPI.InputOpSet,
		Key:   topology_key,
		Value: data,
	}

	c.ddaConnector.ProposeInput(c.ctx, &input)
}

func (c *connector) writeEnergyCommunityToLog() {
	data, _ := json.Marshal(c.state.energyCommunities)

	input := stateAPI.Input{
		Op:    stateAPI.InputOpSet,
		Key:   energy_community_key,
		Value: data,
	}

	c.ddaConnector.ProposeInput(c.ctx, &input)
}

func (c *connector) leaderCh(ctx context.Context) <-chan bool {
	return c.ddaConnector.LeaderCh(ctx)
}

func (c *connector) getFlowProposals() {
	for _, localSenorInformation := range c.state.localSenorInformations {
		localSenorInformation.ecFlowProposal = make(map[string]common.FlowProposal)
	}

	ctx, cancel := context.WithTimeout(
		context.Background(),
		time.Duration(500*time.Millisecond))

	numOutstandingProposals := len(c.state.energyCommunities)
	flowProposals := make(chan common.FlowProposalsMessage, numOutstandingProposals)

	for _, energyCommunity := range c.state.energyCommunities {
		if c.state.topology.Version != energyCommunity.TopologyVersion {
			continue
		}
		go func(energyCommunityId string) {
			if result, err := c.ddaConnector.PublishAction(ctx, api.Action{Type: common.AppendId(common.GET_FLOW_PROPOSAL_ACTION, energyCommunityId), Id: uuid.NewString(), Source: "dso"}); err != nil {
				log.Printf("dso - could not send get flow proposal action - %s", err)
			} else {
				msg := <-result
				if len(msg.Data) == 0 {
					return
				}

				var flowProposal common.FlowProposalsMessage
				if err := json.Unmarshal(msg.Data, &flowProposal); err != nil {
					log.Printf("could not unmarshal flow proposal - %s", err)
				} else {
					flowProposals <- flowProposal
				}
			}
		}(energyCommunity.Id)
	}

	for {
		select {
		case <-ctx.Done():
			cancel()
			addEvent("flowProposalsReceived")
			return
		case energyCommunityProposal := <-flowProposals:
			for sensorId, flowProposal := range energyCommunityProposal.Proposals {
				if sensor, ok := c.state.localSenorInformations[sensorId]; ok {
					sensor.ecFlowProposal[energyCommunityProposal.EnergyCommunityId] = flowProposal
				}
			}
			numOutstandingProposals--
			if numOutstandingProposals == 0 {
				cancel()
				addEvent("flowProposalsReceived")
				return
			}
		}
	}
}

func (c *connector) getSensorData() {
	for _, sensor := range c.state.localSenorInformations {
		sensor.measurement = 0
	}

	go func() {
		ctx, cancel := context.WithCancel(c.ctx)

		sensorResponses, err := c.ddaConnector.PublishAction(ctx, api.Action{Type: common.GET_SENSOR_MEASUREMENT_ACTION, Id: uuid.NewString(), Source: "dso"})
		if err != nil {
			log.Printf("dso - could not get sensor response - %s", err)
			cancel()
			return
		}

		// to get an "AfterEqual()", subtract the minimal timeresolution of message timestamps (unix time - which are in seconds)
		startTime := time.Now().Add(-1 * time.Second)
		go func() {
			for sensorResponse := range sensorResponses {
				var value common.Value
				if err := json.Unmarshal(sensorResponse.Data, &value); err != nil {
					log.Printf("Could not unmarshal incoming sensor message, %s", err)
					continue
				}

				if value.Timestamp.After(startTime) {
					if sensor, ok := c.state.localSenorInformations[value.Id]; ok {
						sensor.measurement = value.Value
					}
				}
			}
		}()

		<-time.After(c.config.WaitTimeForInputs)
		cancel()

		addEvent("dataReceived")
	}()
}

func (c *connector) sendSensorLimits() {
	for energyCommunityId, sensorLimitsMessage := range c.state.energyCommunitySensorLimits {
		data, _ := json.Marshal(sensorLimitsMessage)

		if err := c.ddaConnector.PublishEvent(api.Event{Type: common.AppendId(common.SET_SENSOR_LIMITS_EVENT, energyCommunityId), Source: "dso", Data: data}); err != nil {
			log.Printf("dso - could not send sensor limits event - %s", err)
		}
	}
}

const topology_key = "topology"
const energy_community_key = "energycommunity"
