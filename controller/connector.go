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
	ddaConnector *dda.Connector
	state        *state

	ctx    context.Context
	leader bool
}

func newConnector(config common.ControllerConfig, ddaConnector *dda.Connector, state *state) *connector {
	return &connector{
		config:       config,
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

	sc, err := c.ddaConnector.ObserveStateChange(ctx)
	if err != nil {
		return err
	}

	go func() {
		for {
			select {
			case registerNode := <-registerNodeChannel:
				if c.leader {
					log.Println("controller - got register node")
					var msg common.DdaRegisterMessage
					if err := json.Unmarshal(registerNode.Data, &msg); err != nil {
						log.Printf("Could not unmarshal incoming register message, %s", err)
						continue
					}
					c.writeNodeToLog(msg.NodeId, msg.SensorId)
				}
			case deregisterNode := <-deregisterNodeChannel:
				if c.leader {
					log.Println("controller - got deregister node")
					var msg common.DdaRegisterMessage
					if err := json.Unmarshal(deregisterNode.Data, &msg); err != nil {
						log.Printf("Could not unmarshal incoming deregister message, %s", err)
						continue
					}
					c.removeNodeFromLog(msg.NodeId, msg.SensorId)
				}
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

				sensorId := string(stateChange.Value)
				nodeId := strings.TrimPrefix(stateChange.Key, NODE_PREFIX)

				if stateChange.Op == stateAPI.InputOpSet {
					if _, ok := c.state.topology[sensorId]; !ok {
						c.state.topology[sensorId] = make([]string, 1)
					}
					c.state.topology[sensorId] = append(c.state.topology[sensorId], nodeId)
				} else {
					if _, ok := c.state.topology[sensorId]; !ok {
						continue
					}
					for i, id := range c.state.topology[sensorId] {
						if id == nodeId {
							c.state.topology[sensorId] = append(c.state.topology[sensorId][:i], c.state.topology[sensorId][i+1:]...)
							break
						}
					}

					if len(c.state.topology[sensorId]) == 0 {
						delete(c.state.topology, sensorId)
					}
				}

				if c.leader {
					c.ddaConnector.PublishEvent(api.Event{Type: common.REGISTER_RESPONSE_EVENT, Source: "controller", Id: uuid.NewString(), Data: []byte(nodeId)})
				}
			}
		}
	}()

	return nil
}

func (c *connector) leaderCh(ctx context.Context) <-chan bool {
	return c.ddaConnector.LeaderCh(ctx)
}

func (c *connector) getData() {
	go func() {
		ctx, cancel := context.WithCancel(c.ctx)

		productionResponses, err := c.ddaConnector.PublishAction(ctx, api.Action{Type: common.PRODUCTION_ACTION, Id: uuid.NewString(), Source: "controller"})
		if err != nil {
			log.Printf("controller - could not get PV production - %s", err)
			cancel()
			return
		}

		chargerResponses, err := c.ddaConnector.PublishAction(ctx, api.Action{Type: common.CHARGER_ACTION, Id: uuid.NewString(), Source: "controller"})
		if err != nil {
			log.Printf("controller - could not get available chargers - %s", err)
			cancel()
			return
		}

		c.state.pvProductionValues = make([]common.Value, 0)
		c.state.chargerIds = make([]common.Message, 0)

		// to get an "AfterEqual()", subtract the minimal timeresolution of message timestamps (unix time - which are in seconds)
		startTime := time.Now().Add(-1 * time.Second)
		go func() {
			for productionResponse := range productionResponses {
				var value common.Value
				if err := json.Unmarshal(productionResponse.Data, &value); err != nil {
					log.Printf("Could not unmarshal incoming charger message, %s", err)
					continue
				}

				if value.Timestamp.After(startTime) {
					c.state.pvProductionValues = append(c.state.pvProductionValues, value)
				}
			}
		}()

		go func() {
			for chargerResponse := range chargerResponses {
				var msg common.Message
				if err := json.Unmarshal(chargerResponse.Data, &msg); err != nil {
					log.Printf("Could not unmarshal incoming charger message, %s", err)
					continue
				}

				if msg.Timestamp.After(startTime) {
					c.state.chargerIds = append(c.state.chargerIds, msg)
				}
			}
		}()

		<-time.After(c.config.WaitTimeForInputs)
		cancel()

		addEvent("dataReceived")
	}()
}

func (c *connector) writeNodeToLog(nodeId string, sensorId string) error {
	input := stateAPI.Input{
		Op:    stateAPI.InputOpSet,
		Key:   NODE_PREFIX + nodeId,
		Value: []byte(sensorId),
	}

	return c.ddaConnector.ProposeInput(c.ctx, &input)
}

func (c *connector) removeNodeFromLog(nodeId string, sensorId string) error {
	input := stateAPI.Input{
		Op:    stateAPI.InputOpDelete,
		Key:   NODE_PREFIX + nodeId,
		Value: []byte(sensorId),
	}

	return c.ddaConnector.ProposeInput(c.ctx, &input)
}

func (c *connector) sendChargingSetPoints() {
	for _, setPoint := range c.state.setPoints {
		data, _ := json.Marshal(setPoint)
		if err := c.ddaConnector.PublishEvent(api.Event{Type: common.CHARGING_SET_POINT, Source: "ddaConsistencyProvider", Id: uuid.NewString(), Data: data}); err != nil {
			log.Printf("could not send charging set point - %s", err)
		}
	}
}

const NODE_PREFIX = "node_"
