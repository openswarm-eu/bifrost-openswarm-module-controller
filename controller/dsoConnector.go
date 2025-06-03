package controller

import (
	"context"
	"encoding/json"
	"log"
	"log/slog"
	"time"

	"code.siemens.com/energy-community-controller/common"
	"code.siemens.com/energy-community-controller/dda"
	"github.com/coatyio/dda/services/com/api"
	"github.com/google/uuid"
)

type dsoConnector struct {
	config                   Config
	energyCommunityId        string
	ddaConnector             *dda.Connector
	energyCommunityConnector *energyCommunityConnector
	state                    *state
	flowProposalCallback     api.ActionCallback
}

func newDsoConnector(config Config, energyCommunityId string, ddaConnector *dda.Connector, energyCommunityConnector *energyCommunityConnector, state *state) *dsoConnector {
	return &dsoConnector{
		config:                   config,
		energyCommunityId:        energyCommunityId,
		ddaConnector:             ddaConnector,
		energyCommunityConnector: energyCommunityConnector,
		state:                    state,
	}
}

func (c *dsoConnector) start(ctx context.Context) error {

	topologyUpdateChannel, err := c.ddaConnector.SubscribeAction(ctx, api.SubscriptionFilter{Type: common.AppendId(common.TOPOLOGY_UPDATE_ACTION, c.energyCommunityId)})
	if err != nil {
		return err
	}

	requestFlowProposalChannel, err := c.ddaConnector.SubscribeAction(ctx, api.SubscriptionFilter{Type: common.AppendId(common.GET_FLOW_PROPOSAL_ACTION, c.energyCommunityId)})
	if err != nil {
		return err
	}

	setSensorLimitsChannel, err := c.ddaConnector.SubscribeEvent(ctx, api.SubscriptionFilter{Type: common.AppendId(common.SET_SENSOR_LIMITS_EVENT, c.energyCommunityId)})
	if err != nil {
		return err
	}

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case topologyUpdate := <-topologyUpdateChannel:
				if !c.state.leader {
					continue
				}

				var topologyMessage common.TopologyMessage
				err := json.Unmarshal([]byte(topologyUpdate.Params), &topologyMessage)
				if err != nil {
					slog.Error("controller - error unmarshalling topology update message", "error", err)
					continue
				}

				if topologyMessage.Timestamp.Before(time.Now().Add(-c.config.MaximumMessageAge)) {
					continue
				}

				slog.Info("controller - received topology update message", "topology", topologyMessage.Topology)

				c.energyCommunityConnector.writeTopologyToLog(topologyMessage, func(data []byte) {
					topologyUpdate.Callback(api.ActionResult{Data: data})
				})
			case requestFlowProposal := <-requestFlowProposalChannel:
				if !c.state.leader {
					continue
				}

				c.flowProposalCallback = requestFlowProposal.Callback
				addEvent("flowProposalRequest")
			case sensorLimit := <-setSensorLimitsChannel:
				if !c.state.leader {
					continue
				}

				var sensorLimitsMessage common.EnergyCommunitySensorLimitMessage
				err := json.Unmarshal([]byte(sensorLimit.Data), &sensorLimitsMessage)
				if err != nil {
					slog.Error("controller - error unmarshalling sensor limits message", "error", err)
					continue
				}

				if sensorLimitsMessage.Timestamp.Before(time.Now().Add(-c.config.MaximumMessageAge)) {
					continue
				}

				slog.Info("controller - received sensor limits message", "sensorLimits", sensorLimitsMessage.SensorLimits)

				c.state.toplogy.setAllSensorLimits(0)
				for sensorId, limit := range sensorLimitsMessage.SensorLimits {
					c.state.toplogy.setSensorLimit(sensorId, limit)
				}
				addEvent("sensorLimitsReceived")
			}
		}
	}()

	return nil
}

func (c *dsoConnector) stop() {
	if c.state.registeredAtDso && c.state.clusterMembers == 1 {
		for {
			slog.Info("controller - trying to unregister energy community at DSO")

			deregisterMessage := common.RegisterEnergyCommunityMessage{EnergyCommunityId: c.energyCommunityId, Timestamp: time.Now()}
			data, _ := json.Marshal(deregisterMessage)

			deregisterContext, deregisterCancel := context.WithTimeout(
				context.Background(),
				time.Duration(c.config.RegistrationTimeout))

			result, err := c.ddaConnector.PublishAction(deregisterContext, api.Action{Type: common.DEREGISTER_ENERGY_COMMUNITY_ACTION, Id: uuid.NewString(), Source: c.energyCommunityId, Params: data})
			if err != nil {
				log.Fatalln(err)
			}

			select {
			case <-result:
				slog.Info("controller - energy community deregistered")
				deregisterCancel()
				return
			case <-deregisterContext.Done():
				if deregisterContext.Err() == context.Canceled {
					deregisterCancel()
					return
				}
			}
		}
	}
}

func (c *dsoConnector) registerAtDso(ctx context.Context) {
	go func() {
		for {
			slog.Info("controller - trying to register energy community at DSO")

			registerMessage := common.RegisterEnergyCommunityMessage{EnergyCommunityId: c.energyCommunityId, Timestamp: time.Now()}
			data, _ := json.Marshal(registerMessage)

			registerContext, registerCancel := context.WithTimeout(
				ctx,
				time.Duration(c.config.RegistrationTimeout))

			result, err := c.ddaConnector.PublishAction(registerContext, api.Action{Type: common.REGISTER_ENERGY_COMMUNITY_ACTION, Id: uuid.NewString(), Source: c.energyCommunityId, Params: data})
			if err != nil {
				log.Fatalln(err)
			}

			select {
			case <-result:
				slog.Info("controller - energy community registered")
				registerCancel()
				c.energyCommunityConnector.writeSuccessfullDsoRegistrationToLog()
				return
			case <-registerContext.Done():
				if registerContext.Err() == context.Canceled {
					registerCancel()
					return
				}
			}
		}
	}()
}

func (c *dsoConnector) sendFlowProposal() {
	c.state.toplogy.rootSensor.updateNumberOfGlobalChargerssForFlowProposal()
	c.state.toplogy.rootSensor.updateNumberOfGlobalPVsForFlowProposal()

	flowProposals := make(map[string]common.FlowProposal)
	for _, sensor := range c.state.toplogy.sensors {
		numberOfNodes := 0
		if sensor.flow < 0 {
			numberOfNodes = sensor.numGlobalPVs
		} else if sensor.flow > 0 {
			numberOfNodes = sensor.numGlobalChargers
		}

		flowProposals[sensor.id] = common.FlowProposal{
			Flow:          sensor.flow,
			NumberOfNodes: numberOfNodes,
		}
	}

	data, _ := json.Marshal(common.FlowProposalsMessage{EnergyCommunityId: c.energyCommunityId, Proposals: flowProposals, Timestamp: time.Now()})
	c.flowProposalCallback(api.ActionResult{Data: data})
}
