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

type connector struct {
	config       common.ControllerConfig
	id           string
	ddaConnector *dda.Connector
	state        *state

	ctx    context.Context
	leader bool
}

func newConnector(config common.ControllerConfig, id string, ddaConnector *dda.Connector, state *state) *connector {
	return &connector{
		config:       config,
		id:           id,
		ddaConnector: ddaConnector,
		state:        state,
		leader:       false,
	}
}

func (c *connector) start(ctx context.Context) error {
	c.ctx = ctx

	leaderChannel := c.leaderCh(ctx)

	registerNodeChannel, err := c.ddaConnector.SubscribeEvent(ctx, api.SubscriptionFilter{Type: common.REGISTER_EVENT})
	if err != nil {
		return err
	}

	deregisterNodeChannel, err := c.ddaConnector.SubscribeEvent(ctx, api.SubscriptionFilter{Type: common.DEREGISTER_EVENT})
	if err != nil {
		return err
	}

	requestFlowProposalChannel, err := c.ddaConnector.SubscribeEvent(ctx, api.SubscriptionFilter{Type: common.NEW_ROUND_EVENT})
	if err != nil {
		return err
	}

	sensorLimitChannel, err := c.ddaConnector.SubscribeEvent(ctx, api.SubscriptionFilter{Type: common.AppendId(common.SENSOR_LIMITS_EVENT, c.id)})
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
			case registerNode := <-registerNodeChannel:
				if c.leader {
					continue
				}

				log.Println("controller - got register node")
				var msg common.DdaRegisterNodeMessage
				if err := json.Unmarshal(registerNode.Data, &msg); err != nil {
					log.Printf("Could not unmarshal incoming register message, %s", err)
					continue
				}
				c.writeNodeToLog(msg)
			case deregisterNode := <-deregisterNodeChannel:
				if c.leader {
					continue
				}

				log.Println("controller - got deregister node")
				var msg common.DdaRegisterNodeMessage
				if err := json.Unmarshal(deregisterNode.Data, &msg); err != nil {
					log.Printf("Could not unmarshal incoming deregister message, %s", err)
					continue
				}

				c.removeNodeFromLog(msg)
			case v := <-leaderChannel:
				if v {
					c.leader = true
				} else {
					c.leader = false
				}
			case stateChange := <-sc:
				if !strings.HasPrefix(stateChange.Key, NODE_PREFIX) {
					continue
				}

				nodeId := strings.TrimPrefix(stateChange.Key, NODE_PREFIX)

				if stateChange.Op == stateAPI.InputOpSet {
					var msg node
					if err := json.Unmarshal(stateChange.Value, &msg); err != nil {
						log.Printf("Could not unmarshal incoming node state change message, %s", err)
						continue
					}

					if _, ok := c.state.sensors[msg.SensorId]; !ok {
						c.state.sensors[msg.SensorId] = &sensor{id: msg.SensorId, childSensors: make([]*sensor, 0), pvs: make([]*component, 0), chargers: make([]*component, 0)}
					}

					if msg.NodeType == common.PV_NODE_TPYE {
						c.state.sensors[msg.SensorId].pvs = append(c.state.sensors[msg.SensorId].pvs, &component{id: msg.Id, demand: 0, setPoint: 0})
					} else if msg.NodeType == common.CHARGER_NODE_TYPE {
						c.state.sensors[msg.SensorId].chargers = append(c.state.sensors[msg.SensorId].chargers, &component{id: msg.Id, demand: 0, setPoint: 0})
					}
				} else {
					for _, sensor := range c.state.sensors {
						found := false
						for i, pv := range sensor.pvs {
							if pv.id == nodeId {
								sensor.pvs = append(sensor.pvs[:i], sensor.pvs[i+1:]...)
								found = true
								break
							}
						}

						if !found {
							for i, charger := range sensor.chargers {
								if charger.id == nodeId {
									sensor.chargers = append(sensor.chargers[:i], sensor.chargers[i+1:]...)
									break
								}
							}
						}

						if found && len(sensor.pvs) != 0 && len(sensor.chargers) != 0 && len(sensor.childSensors) != 0 {
							delete(c.state.sensors, sensor.id)
						}
					}
				}

				if c.leader {
					c.ddaConnector.PublishEvent(api.Event{Type: common.AppendId(common.REGISTER_RESPONSE_EVENT, nodeId), Source: c.id, Id: uuid.NewString(), Data: []byte(nodeId)})
				}
			case <-requestFlowProposalChannel:
				if c.leader {
					addEvent("newRound")
				}
			case sensorLimits := <-sensorLimitChannel:
				if !c.leader {
					continue
				}

				var msg common.SensorLimitsMessage
				if err := json.Unmarshal(sensorLimits.Data, &msg); err != nil {
					log.Printf("Could not unmarshal incoming sensor limits message, %s", err)
					continue
				}

				for _, limit := range msg.Limits {
					if sensor, ok := c.state.sensors[limit.SensorId]; ok {
						sensor.limit = limit.Limit
					}
				}

				addEvent("sensorLimitsReceived")
			}
		}
	}()

	return nil
}

func (c *connector) leaderCh(ctx context.Context) <-chan bool {
	return c.ddaConnector.LeaderCh(ctx)
}

func (c *connector) getData() {
	for _, pv := range c.state.pvs {
		pv.demand = 0
	}
	for _, charger := range c.state.chargers {
		charger.demand = 0
	}

	go func() {
		ctx, cancel := context.WithCancel(c.ctx)

		pvResponses, err := c.ddaConnector.PublishAction(ctx, api.Action{Type: common.GET_PV_DEMAND_ACTION, Id: uuid.NewString(), Source: c.id})
		if err != nil {
			log.Printf("controller - could not get PV response - %s", err)
			cancel()
			return
		}

		chargerResponses, err := c.ddaConnector.PublishAction(ctx, api.Action{Type: common.GET_CHARGER_DEMAND_ACTION, Id: uuid.NewString(), Source: c.id})
		if err != nil {
			log.Printf("controller - could not get charger response - %s", err)
			cancel()
			return
		}

		// to get an "AfterEqual()", subtract the minimal timeresolution of message timestamps (unix time - which are in seconds)
		startTime := time.Now().Add(-1 * time.Second)
		go func() {
			for pvResponse := range pvResponses {
				var value common.Value
				if err := json.Unmarshal(pvResponse.Data, &value); err != nil {
					log.Printf("Could not unmarshal incoming PV message, %s", err)
					continue
				}

				if value.Timestamp.After(startTime) {
					if pv, ok := c.state.pvs[value.Id]; ok {
						pv.demand = value.Value
					}
				}
			}
		}()

		go func() {
			for chargerResponse := range chargerResponses {
				var value common.Value
				if err := json.Unmarshal(chargerResponse.Data, &value); err != nil {
					log.Printf("Could not unmarshal incoming charger message, %s", err)
					continue
				}

				if value.Timestamp.After(startTime) {
					if charger, ok := c.state.pvs[value.Id]; ok {
						charger.demand = value.Value
					}
				}
			}
		}()

		<-time.After(c.config.WaitTimeForInputs)
		cancel()

		addEvent("dataReceived")
	}()
}

func (c *connector) writeNodeToLog(registerMessage common.DdaRegisterNodeMessage) error {
	data, _ := json.Marshal(node{Id: registerMessage.NodeId, SensorId: registerMessage.SensorId, NodeType: registerMessage.NodeType})

	input := stateAPI.Input{
		Op:    stateAPI.InputOpSet,
		Key:   NODE_PREFIX + registerMessage.NodeId,
		Value: data,
	}

	return c.ddaConnector.ProposeInput(c.ctx, &input)
}

func (c *connector) removeNodeFromLog(registerMessage common.DdaRegisterNodeMessage) error {
	input := stateAPI.Input{
		Op:  stateAPI.InputOpDelete,
		Key: NODE_PREFIX + registerMessage.NodeId,
	}

	return c.ddaConnector.ProposeInput(c.ctx, &input)
}

func (c *connector) sendFlows() {
	/*flowProposals := make([]common.FlowProposal, len(c.state.sensors))
	for _, sensor := range c.state.sensors {
		flowProposals = append(flowProposals, common.FlowProposal{
			SensorId: sensor.id,
			Flow:     sensor.flow,
		})
	}

	data, _ := json.Marshal(common.FlowProposalsMessage{Proposals: flowProposals})*/
}

func (c *connector) sendSetPoints() {
	for _, charger := range c.state.chargers {
		msg := common.Value{Message: common.Message{Id: charger.id, Timestamp: time.Now()}, Value: charger.setPoint}
		data, _ := json.Marshal(msg)
		if err := c.ddaConnector.PublishEvent(api.Event{Type: common.AppendId(common.SET_POINT, charger.id), Source: c.id, Id: uuid.NewString(), Data: data}); err != nil {
			log.Printf("could not send charging set point - %s", err)
		}
	}

	for _, pv := range c.state.pvs {
		msg := common.Value{Message: common.Message{Id: pv.id, Timestamp: time.Now()}, Value: pv.setPoint}
		data, _ := json.Marshal(msg)
		if err := c.ddaConnector.PublishEvent(api.Event{Type: common.AppendId(common.SET_POINT, pv.id), Source: c.id, Id: uuid.NewString(), Data: data}); err != nil {
			log.Printf("could not send pv set point - %s", err)
		}
	}
}

const NODE_PREFIX = "node_"

type node struct {
	Id       string
	SensorId string
	NodeType string
}
