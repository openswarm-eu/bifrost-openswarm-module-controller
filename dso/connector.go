package dso

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
	config       Config
	ddaConnector *dda.Connector
	state        *state

	registerCallbacks                  map[string]api.ActionCallback
	deregisterCallbacks                map[string]api.ActionCallback
	registerEnergyCommunityCallbacks   map[string]api.ActionCallback
	deregisterEnergyCommunityCallbacks map[string]api.ActionCallback

	ctx context.Context
}

func newConnector(config Config, ddaConnector *dda.Connector, state *state) *connector {
	return &connector{
		config:                             config,
		ddaConnector:                       ddaConnector,
		state:                              state,
		registerCallbacks:                  make(map[string]api.ActionCallback),
		deregisterCallbacks:                make(map[string]api.ActionCallback),
		registerEnergyCommunityCallbacks:   make(map[string]api.ActionCallback),
		deregisterEnergyCommunityCallbacks: make(map[string]api.ActionCallback),
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

				var msg common.RegisterSensorMessage
				if err := json.Unmarshal(registerSensor.Params, &msg); err != nil {
					log.Printf("Could not unmarshal incoming register sensor message, %s", err)
					continue
				}

				if msg.Timestamp.Before(time.Now().Add(-c.config.MaximumMessageAge)) {
					continue
				}

				log.Printf("dso - got register sensor %s", msg.SensorId)

				c.registerCallbacks[msg.SensorId] = registerSensor.Callback
				c.writeSensorToLog(msg)
			case deregisterSensor := <-deregisterSensorChannel:
				if !c.state.leader {
					continue
				}

				var msg common.RegisterSensorMessage
				if err := json.Unmarshal(deregisterSensor.Params, &msg); err != nil {
					log.Printf("Could not unmarshal incoming deregister sensor message, %s", err)
					continue
				}

				if msg.Timestamp.Before(time.Now().Add(-c.config.MaximumMessageAge)) {
					continue
				}

				log.Printf("dso - got deregister sensor %s", msg.SensorId)

				c.deregisterCallbacks[msg.SensorId] = deregisterSensor.Callback
				c.removeSensorFromLog(msg)
			case registerEnergyCommunity := <-registerEnergyCommunityChannel:
				if !c.state.leader {
					continue
				}

				var msg common.RegisterEnergyCommunityMessage
				if err := json.Unmarshal(registerEnergyCommunity.Params, &msg); err != nil {
					log.Printf("Could not unmarshal incoming register energy community message, %s", err)
					continue
				}

				if msg.Timestamp.Before(time.Now().Add(-c.config.MaximumMessageAge)) {
					continue
				}

				log.Printf("dso - got register energy community %s", msg.EnergyCommunityId)

				c.registerEnergyCommunityCallbacks[msg.EnergyCommunityId] = registerEnergyCommunity.Callback
				c.writeEnergyCommunityToLog(msg.EnergyCommunityId, 0)
			case derigsterEnergyCommunity := <-deregisterEnergyCommunityChannel:
				if !c.state.leader {
					continue
				}

				var msg common.RegisterEnergyCommunityMessage
				if err := json.Unmarshal(derigsterEnergyCommunity.Params, &msg); err != nil {
					log.Printf("Could not unmarshal incoming deregister energy community message, %s", err)
					continue
				}

				if msg.Timestamp.Before(time.Now().Add(-c.config.MaximumMessageAge)) {
					continue
				}

				log.Printf("dso - got deregister energy community %s", msg.EnergyCommunityId)

				c.deregisterEnergyCommunityCallbacks[msg.EnergyCommunityId] = derigsterEnergyCommunity.Callback
				c.removeEnergyCommunityFromLog(msg.EnergyCommunityId)
			case stateChange := <-sc:
				if strings.HasPrefix(stateChange.Key, sensor_prefix) {
					sensorId := strings.TrimPrefix(stateChange.Key, sensor_prefix)
					c.state.newTopology.Version++

					if stateChange.Op == stateAPI.InputOpSet {
						var sensorLogEntry sensorLogEntry
						if err := json.Unmarshal(stateChange.Value, &sensorLogEntry); err != nil {
							log.Printf("Could not unmarshal incoming sensor log entry message, %s", err)
							continue
						}

						c.state.addNodeToTopology(sensorId, sensorLogEntry.ParentSensorId, sensorLogEntry.Limit)

						if !c.state.leader {
							continue
						}

						for sensorId, callback := range c.registerCallbacks {
							if _, ok := c.state.newTopology.Sensors[sensorId]; !ok {
								continue
							}
							callback(api.ActionResult{Data: []byte(sensorId)})
							delete(c.registerCallbacks, sensorId)
						}
					}

					for sensorId, callback := range c.deregisterCallbacks {
						c.state.removeNodeFromTopology(sensorId)

						if !c.state.leader {
							continue
						}

						if _, ok := c.state.newTopology.Sensors[sensorId]; ok {
							continue
						}
						callback(api.ActionResult{Data: []byte(sensorId)})
						delete(c.deregisterCallbacks, sensorId)
					}
				} else if strings.HasPrefix(stateChange.Key, energy_community_prefix) {
					energyCommunityId := strings.TrimPrefix(stateChange.Key, energy_community_prefix)

					if stateChange.Op == stateAPI.InputOpSet {
						var energyCommunityLogEntry energyCommunityLogEntry
						if err := json.Unmarshal(stateChange.Value, &energyCommunityLogEntry); err != nil {
							log.Printf("Could not unmarshal incoming energy community log entry message, %s", err)
							continue
						}

						c.state.energyCommunities[energyCommunityId] = energyCommunityLogEntry.Version

						if !c.state.leader {
							continue
						}

						if callback, ok := c.registerEnergyCommunityCallbacks[energyCommunityId]; ok {
							callback(api.ActionResult{Data: []byte(energyCommunityId)})
							delete(c.registerEnergyCommunityCallbacks, energyCommunityId)
						}
					} else {
						delete(c.state.energyCommunities, energyCommunityId)

						if callback, ok := c.deregisterEnergyCommunityCallbacks[energyCommunityId]; ok {
							callback(api.ActionResult{Data: []byte(energyCommunityId)})
							delete(c.registerEnergyCommunityCallbacks, energyCommunityId)
						}
					}
				}
			}
		}
	}()
	return nil
}

func (c *connector) writeSensorToLog(registerMessage common.RegisterSensorMessage) error {
	data, _ := json.Marshal(sensorLogEntry{Limit: registerMessage.Limit, ParentSensorId: registerMessage.ParentSensorId})

	input := stateAPI.Input{
		Op:    stateAPI.InputOpSet,
		Key:   sensor_prefix + registerMessage.SensorId,
		Value: data,
	}

	return c.ddaConnector.ProposeInput(c.ctx, &input)
}

func (c *connector) removeSensorFromLog(registerMessage common.RegisterSensorMessage) error {
	input := stateAPI.Input{
		Op:  stateAPI.InputOpDelete,
		Key: sensor_prefix + registerMessage.SensorId,
	}

	return c.ddaConnector.ProposeInput(c.ctx, &input)
}

func (c *connector) writeEnergyCommunityToLog(id string, version int) {
	data, _ := json.Marshal(energyCommunityLogEntry{Version: version})

	input := stateAPI.Input{
		Op:    stateAPI.InputOpSet,
		Key:   energy_community_prefix + id,
		Value: data,
	}

	c.ddaConnector.ProposeInput(c.ctx, &input)
}

func (c *connector) removeEnergyCommunityFromLog(id string) error {
	input := stateAPI.Input{
		Op:  stateAPI.InputOpDelete,
		Key: energy_community_prefix + id,
	}

	return c.ddaConnector.ProposeInput(c.ctx, &input)
}

func (c *connector) leaderCh(ctx context.Context) <-chan bool {
	return c.ddaConnector.LeaderCh(ctx)
}

func (c *connector) getFlowProposals() {
	if len(c.state.energyCommunities) == 0 {
		addEvent("flowProposalsReceived")
		return
	}

	ctx, cancel := context.WithTimeout(
		context.Background(),
		time.Duration(c.config.FlowProposalsTimeout))

	numOutstandingProposals := len(c.state.energyCommunities)
	flowProposals := make(chan common.FlowProposalsMessage, numOutstandingProposals)

	// to get an "AfterEqual()", subtract some time from the current time
	startTime := time.Now().Add(-1 * time.Millisecond)
	for energyCommunityId, topologyVersion := range c.state.energyCommunities {
		if c.state.topology.Version != topologyVersion {
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
					if flowProposal.Timestamp.After(startTime) {
						flowProposals <- flowProposal
					}
				}
			}
		}(energyCommunityId)
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
					log.Printf("dso - received flow proposal for %s from %s: %+v", sensorId, energyCommunityProposal.EnergyCommunityId, flowProposal)
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

func (c *connector) getSensorMeasurements() {
	numOutstandingSensorResponses := len(c.state.localSenorInformations)

	ctx, cancel := context.WithTimeout(
		context.Background(),
		time.Duration(c.config.SensorMeasurementsTimeout))

	sensorResponses, err := c.ddaConnector.PublishAction(ctx, api.Action{Type: common.GET_SENSOR_MEASUREMENT_ACTION, Id: uuid.NewString(), Source: "dso"})
	if err != nil {
		log.Printf("dso - could not get sensor response - %s", err)
		cancel()
		return
	}

	// to get an "AfterEqual()", subtract some time from the current time
	startTime := time.Now().Add(-1 * time.Millisecond)
	for {
		select {
		case <-ctx.Done():
			cancel()
			addEvent("sensorMeasurementsReceived")
			return
		case sensorResponse := <-sensorResponses:
			var value common.Value
			if err := json.Unmarshal(sensorResponse.Data, &value); err != nil {
				log.Printf("Could not unmarshal incoming sensor message, %s", err)
				continue
			}

			if value.Timestamp.After(startTime) {
				if sensor, ok := c.state.localSenorInformations[value.Id]; ok {
					log.Println("dso - got sensor measurement", value.Id, value.Value)
					sensor.measurement = value.Value
				}

				numOutstandingSensorResponses--
				if numOutstandingSensorResponses == 0 {
					cancel()
					addEvent("sensorMeasurementsReceived")
					return
				}
			}
		}
	}
}

func (c *connector) sendSensorLimits() {
	for energyCommunityId, sensorLimitsMessage := range c.state.energyCommunitySensorLimits {
		log.Println("dso - sending sensor limits for energy community", energyCommunityId, sensorLimitsMessage)
		sensorLimitsMessage.Timestamp = time.Now()
		data, _ := json.Marshal(sensorLimitsMessage)

		if err := c.ddaConnector.PublishEvent(api.Event{Type: common.AppendId(common.SET_SENSOR_LIMITS_EVENT, energyCommunityId), Id: uuid.NewString(), Source: "dso", Data: data}); err != nil {
			log.Printf("dso - could not send sensor limits event - %s", err)
		}
	}
}

const sensor_prefix = "sensor_"
const energy_community_prefix = "energy_community_"

type sensorLogEntry struct {
	Limit          float64
	ParentSensorId string
}

type energyCommunityLogEntry struct {
	Version int
}
