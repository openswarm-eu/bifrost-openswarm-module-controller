package controller

import (
	"context"
	"encoding/json"
	"log"
	"strings"
	"time"

	"code.siemens.com/energy-community-controller/common"
	"code.siemens.com/energy-community-controller/dda"
	"github.com/coatyio/dda/services/com/api"
	stateAPI "github.com/coatyio/dda/services/state/api"
	"github.com/google/uuid"
)

type energyCommunityConnector struct {
	config            Config
	energyCommunityId string
	ddaConnector      *dda.Connector
	state             *state
	callbackManager   *callbackManager

	ctx context.Context
}

func newEnergyCommunityConnector(config Config, energyCommunityId string, ddaConnector *dda.Connector, state *state) *energyCommunityConnector {
	return &energyCommunityConnector{
		config:            config,
		energyCommunityId: energyCommunityId,
		ddaConnector:      ddaConnector,
		state:             state,
		callbackManager:   newCallbackManager(),
	}
}

func (c *energyCommunityConnector) start(ctx context.Context) error {
	c.ctx = context.Background()

	registerNodeChannel, err := c.ddaConnector.SubscribeAction(ctx, api.SubscriptionFilter{Type: common.REGISTER_ACTION})
	if err != nil {
		return err
	}

	deregisterNodeChannel, err := c.ddaConnector.SubscribeAction(ctx, api.SubscriptionFilter{Type: common.DEREGISTER_ACTION})
	if err != nil {
		return err
	}

	sc, err := c.ddaConnector.ObserveStateChange(ctx)
	if err != nil {
		return err
	}

	mc, err := c.ddaConnector.ObserveMembershipChange(ctx)
	if err != nil {
		return err
	}

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case registerNode := <-registerNodeChannel:
				if !c.state.leader {
					continue
				}

				log.Println("controller - got register node")
				var msg common.RegisterNodeMessage
				if err := json.Unmarshal(registerNode.Params, &msg); err != nil {
					log.Printf("Could not unmarshal incoming register message, %s", err)
					continue
				}

				c.callbackManager.addCallback(msg.NodeId, func(data []byte) {
					registerNode.Callback(api.ActionResult{Data: data})
				})
				c.writeNodeToLog(msg)
			case deregisterNode := <-deregisterNodeChannel:
				if !c.state.leader {
					continue
				}

				log.Println("controller - got deregister node")
				var msg common.RegisterNodeMessage
				if err := json.Unmarshal(deregisterNode.Params, &msg); err != nil {
					log.Printf("Could not unmarshal incoming deregister message, %s", err)
					continue
				}

				c.callbackManager.addCallback(msg.NodeId, func(data []byte) {
					deregisterNode.Callback(api.ActionResult{Data: data})
				})
				c.removeNodeFromLog(msg)
			case stateChange := <-sc:
				if strings.HasPrefix(stateChange.Key, NODE_PREFIX) {
					nodeId := strings.TrimPrefix(stateChange.Key, NODE_PREFIX)

					if stateChange.Op == stateAPI.InputOpSet {
						var msg node
						if err := json.Unmarshal(stateChange.Value, &msg); err != nil {
							log.Printf("Could not unmarshal incoming node state change message, %s", err)
							continue
						}

						if msg.NodeType == common.PV_NODE_TPYE {
							c.state.toplogy.addPV(msg.Id, msg.SensorId)
						} else if msg.NodeType == common.CHARGER_NODE_TYPE {
							c.state.toplogy.addCharger(msg.Id, msg.SensorId)
						}

						if callback, ok := c.callbackManager.getCallback(nodeId); ok {
							callback([]byte(nodeId))
						}
					} else {
						c.state.toplogy.removeNode(nodeId)
						if callback, ok := c.callbackManager.getCallback(nodeId); ok {
							callback([]byte(nodeId))
						}
					}
				} else if stateChange.Key == REGISTER_DSO_KEY {
					c.state.registeredAtDso = true
				} else if stateChange.Key == TOPOLOGY_KEY {
					var topologyEntry topologyLogEntry
					if err := json.Unmarshal(stateChange.Value, &topologyEntry); err != nil {
						log.Printf("Could not unmarshal incoming topology message, %s", err)
						continue
					}

					c.state.toplogy.buildTopology(topologyEntry.Children)

					if callback, ok := c.callbackManager.getCallback(topologyEntry.Id); ok {
						callback([]byte(topologyEntry.Id))
					}
				}
			case membershipChange := <-mc:
				if membershipChange.Joined {
					c.state.clusterMembers++
				} else {
					c.state.clusterMembers--
				}
			}
		}
	}()

	return nil
}

func (c *energyCommunityConnector) leaderCh(ctx context.Context) <-chan bool {
	return c.ddaConnector.LeaderCh(ctx)
}

func (c *energyCommunityConnector) getData() {
	for _, pv := range c.state.toplogy.pvs {
		pv.demand = 0
	}
	for _, charger := range c.state.toplogy.chargers {
		charger.demand = 0
	}

	ctx, cancel := context.WithTimeout(
		context.Background(),
		time.Duration(500*time.Millisecond))

	outstandingPVResponses := len(c.state.toplogy.pvs)
	outstandingChargerResponses := len(c.state.toplogy.chargers)

	pvResponses, err := c.ddaConnector.PublishAction(ctx, api.Action{Type: common.GET_PV_DEMAND_ACTION, Id: uuid.NewString(), Source: c.energyCommunityId})
	if err != nil {
		log.Printf("controller - could not get PV response - %s", err)
		cancel()
		return
	}

	chargerResponses, err := c.ddaConnector.PublishAction(ctx, api.Action{Type: common.GET_CHARGER_DEMAND_ACTION, Id: uuid.NewString(), Source: c.energyCommunityId})
	if err != nil {
		log.Printf("controller - could not get charger response - %s", err)
		cancel()
		return
	}

	// to get an "AfterEqual()", subtract the minimal timeresolution of message timestamps (unix time - which are in seconds)
	startTime := time.Now().Add(-1 * time.Second)
	go func() {
		for {
			select {
			case <-ctx.Done():
				cancel()
				addEvent("dataReceived")
				return
			case pvResponse := <-pvResponses:
				var value common.Value
				if err := json.Unmarshal(pvResponse.Data, &value); err != nil {
					log.Printf("Could not unmarshal incoming PV message, %s", err)
					continue
				}

				if value.Timestamp.After(startTime) {
					if pv, ok := c.state.toplogy.pvs[value.Id]; ok {
						log.Println("controller - got pv response", value.Id, value.Value)
						pv.demand = value.Value
					}

					outstandingPVResponses--
					if outstandingPVResponses == 0 && outstandingChargerResponses == 0 {
						cancel()
						addEvent("dataReceived")
						return
					}
				}
			case chargerResponse := <-chargerResponses:
				var value common.Value
				if err := json.Unmarshal(chargerResponse.Data, &value); err != nil {
					log.Printf("Could not unmarshal incoming charger message, %s", err)
					continue
				}

				if value.Timestamp.After(startTime) {
					if charger, ok := c.state.toplogy.pvs[value.Id]; ok {
						charger.demand = value.Value
					}

					outstandingChargerResponses--
					if outstandingPVResponses == 0 && outstandingChargerResponses == 0 {
						cancel()
						addEvent("dataReceived")
						return
					}
				}
			}
		}
	}()
}

func (c *energyCommunityConnector) writeNodeToLog(registerMessage common.RegisterNodeMessage) error {
	data, _ := json.Marshal(node{Id: registerMessage.NodeId, SensorId: registerMessage.SensorId, NodeType: registerMessage.NodeType})

	input := stateAPI.Input{
		Op:    stateAPI.InputOpSet,
		Key:   NODE_PREFIX + registerMessage.NodeId,
		Value: data,
	}

	return c.ddaConnector.ProposeInput(c.ctx, &input)
}

func (c *energyCommunityConnector) removeNodeFromLog(registerMessage common.RegisterNodeMessage) error {
	input := stateAPI.Input{
		Op:  stateAPI.InputOpDelete,
		Key: NODE_PREFIX + registerMessage.NodeId,
	}

	return c.ddaConnector.ProposeInput(c.ctx, &input)
}

func (c *energyCommunityConnector) writeSuccessfullDsoRegistrationToLog() error {
	input := stateAPI.Input{
		Op:    stateAPI.InputOpSet,
		Key:   REGISTER_DSO_KEY,
		Value: []byte("true"),
	}

	return c.ddaConnector.ProposeInput(c.ctx, &input)
}

func (c *energyCommunityConnector) writeTopologyToLog(toplogy common.TopologyMessage, callback func(data []byte)) error {
	id := uuid.NewString()
	c.callbackManager.addCallback(id, callback)

	data, _ := json.Marshal(topologyLogEntry{Id: id, Children: toplogy.Topology})
	input := stateAPI.Input{
		Op:    stateAPI.InputOpSet,
		Key:   TOPOLOGY_KEY,
		Value: data,
	}

	return c.ddaConnector.ProposeInput(c.ctx, &input)
}

func (c *energyCommunityConnector) sendSetPoints() {
	for _, charger := range c.state.toplogy.chargers {
		msg := common.Value{Message: common.Message{Id: charger.id, Timestamp: time.Now()}, Value: charger.setPoint}
		data, _ := json.Marshal(msg)
		if err := c.ddaConnector.PublishEvent(api.Event{Type: common.AppendId(common.SET_POINT, charger.id), Source: c.energyCommunityId, Id: uuid.NewString(), Data: data}); err != nil {
			log.Printf("could not send charging set point - %s", err)
		}
	}

	for _, pv := range c.state.toplogy.pvs {
		msg := common.Value{Message: common.Message{Id: pv.id, Timestamp: time.Now()}, Value: pv.setPoint}
		data, _ := json.Marshal(msg)
		if err := c.ddaConnector.PublishEvent(api.Event{Type: common.AppendId(common.SET_POINT, pv.id), Source: c.energyCommunityId, Id: uuid.NewString(), Data: data}); err != nil {
			log.Printf("could not send pv set point - %s", err)
		}
	}
}

const NODE_PREFIX = "node_"

const REGISTER_DSO_KEY = "registered"
const TOPOLOGY_KEY = "topology"

type node struct {
	Id       string
	SensorId string
	NodeType string
}

type topologyLogEntry struct {
	Id       string
	Children map[string][]string
}
