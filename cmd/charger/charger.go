package main

import (
	"context"
	"encoding/json"
	"flag"
	"log"
	"os"
	"os/signal"
	"time"

	"code.siemens.com/energy-community-controller/common"
	"code.siemens.com/energy-community-controller/controller"
	"code.siemens.com/energy-community-controller/dda"
	"code.siemens.com/energy-community-controller/mqtt"
	"github.com/coatyio/dda/services/com/api"
	"github.com/google/uuid"
)

func main() {
	log.Println("starting charger")

	var nodeId string
	var url string
	var energyCommunityId string
	var sensorId string
	bootstrap := flag.Bool("b", false, "bootstrap raft")
	leadershipElectionEnabled := flag.Bool("l", false, "participate in leader election")
	flag.StringVar(&nodeId, "id", uuid.NewString(), "id")
	flag.StringVar(&url, "url", "tcp://localhost:1883", "mqtt url")
	flag.StringVar(&energyCommunityId, "energyCommunityId", "energyCommunity", "energy community id")
	flag.StringVar(&sensorId, "sensorId", "sensor", "sensor id")
	flag.Parse()

	var ddaConnectorEnergyCommunity *dda.Connector
	var mqttConnector *mqtt.Connector
	var ctrl *controller.Controller
	var demand float64
	var err error

	maximumAcceptableSetPointOffset := 1000 * time.Millisecond
	controllerPeriode := 1000 * time.Millisecond

	ctx, cancel := context.WithCancel(context.Background())

	defer func() {
		log.Println("shutting down")

		cancel()

		deregister(ctx, ddaConnectorEnergyCommunity, nodeId, sensorId)
		if ctrl != nil {
			ctrl.Stop()
		}

		if ddaConnectorEnergyCommunity != nil {
			ddaConnectorEnergyCommunity.Close()
		}

		if mqttConnector != nil {
			mqttConnector.Close()
		}
	}()

	ddaEnergyCommunityConfig := dda.NewConfig()
	ddaEnergyCommunityConfig.Name = "pv"
	ddaEnergyCommunityConfig.Url = url
	ddaEnergyCommunityConfig.Id = nodeId
	ddaEnergyCommunityConfig.Cluster = energyCommunityId
	ddaEnergyCommunityConfig.Leader.Enabled = *leadershipElectionEnabled
	ddaEnergyCommunityConfig.Leader.Bootstrap = *bootstrap

	if ddaConnectorEnergyCommunity, err = dda.NewConnector(ddaEnergyCommunityConfig); err != nil {
		log.Fatalln(err)
	}

	if err = ddaConnectorEnergyCommunity.Open(); err != nil {
		log.Fatalln(err)
	}

	if *leadershipElectionEnabled {
		ddaDsoConfig := dda.NewConfig()
		ddaDsoConfig.Name = energyCommunityId
		ddaDsoConfig.Url = url
		ddaDsoConfig.Id = nodeId
		ddaDsoConfig.Cluster = "dso"
		ddaDsoConfig.Leader.Enabled = false
		ddaDsoConfig.Leader.Bootstrap = false

		var ddaConnectorDso *dda.Connector
		if ddaConnectorDso, err = dda.NewConnector(ddaDsoConfig); err != nil {
			log.Fatalln(err)
		}

		if err = ddaConnectorDso.Open(); err != nil {
			log.Fatalln(err)
		}

		controllerConfig := controller.NewConfig()
		controllerConfig.Periode = controllerPeriode

		if ctrl, err = controller.NewController(controllerConfig, energyCommunityId, ddaConnectorEnergyCommunity, ddaConnectorDso); err != nil {
			log.Fatalln(err)
		} else {
			if err := ctrl.Start(); err != nil {
				log.Fatalln(err)
			}
		}
	}

	if mqttConnector, err = mqtt.NewConnector(mqtt.Config{Url: url, Id: nodeId}); err != nil {
		log.Fatalln(err)
	}

	if err = mqttConnector.Open(ctx); err != nil {
		log.Fatalln(err)
	}

	register(ctx, ddaConnectorEnergyCommunity, nodeId, sensorId)

	demandChannel, err := mqttConnector.SubscribeToDemands(ctx)
	if err != nil {
		log.Fatalln(err)
	}

	chargerDemandRequests, err := ddaConnectorEnergyCommunity.SubscribeAction(ctx, api.SubscriptionFilter{Type: common.GET_CHARGER_DEMAND_ACTION})
	if err != nil {
		log.Fatalln(err)
	}

	setPointChannel, err := ddaConnectorEnergyCommunity.SubscribeEvent(ctx, api.SubscriptionFilter{Type: common.AppendId(common.SET_POINT, energyCommunityId)})
	if err != nil {
		log.Fatalln(err)
	}

	setPointMonitorDuration := controllerPeriode + maximumAcceptableSetPointOffset
	var setPointMonitor common.Timer
	setPointMonitor.Start(setPointMonitorDuration, func() {
		log.Println("charger - set point timeout")
		mqttConnector.PublishSetPoint(ctx, 0)
	})

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)

	for {
		select {
		case newDemand := <-demandChannel:
			log.Printf("pv - got new production value: %f", newDemand)
			demand = newDemand
		case chargerDemandRequest := <-chargerDemandRequests:
			msg := common.Value{Message: common.Message{Id: nodeId, Timestamp: time.Now()}, Value: demand}
			data, _ := json.Marshal(msg)
			chargerDemandRequest.Callback(api.ActionResult{Data: data})
		case setPoint := <-setPointChannel:
			var value common.Value
			if err := json.Unmarshal(setPoint.Data, &value); err != nil {
				log.Printf("charger - could not unmarshal incoming set point, %s", err)
				continue
			}

			if value.Id != nodeId {
				continue
			}

			if value.Timestamp.After(time.Now().Add(-maximumAcceptableSetPointOffset)) {
				log.Printf("charger - got new set point: %f", value.Value)
				setPointMonitor.Reset(setPointMonitorDuration)
				mqttConnector.PublishSetPoint(ctx, value.Value)
			} else {
				log.Println("charger - got too old set point, ignoring it")
				log.Printf("charger - now: %s, got: %s", time.Now(), value.Timestamp)
			}
		case <-sigChan:
			return
		}
	}
}

func register(ctx context.Context, ddaConnector *dda.Connector, nodeId string, sensorId string) {
	registerMessage := common.RegisterNodeMessage{NodeId: nodeId, SensorId: sensorId, NodeType: common.CHARGER_NODE_TYPE, Timestamp: time.Now().Unix()}
	data, _ := json.Marshal(registerMessage)

	for {
		log.Println("charger - trying to register node")

		registerContext, registerCancel := context.WithCancel(ctx)

		result, err := ddaConnector.PublishAction(registerContext, api.Action{Type: common.REGISTER_ACTION, Id: uuid.NewString(), Source: nodeId, Params: data})
		if err != nil {
			log.Fatalln(err)
		}

		select {
		case <-result:
			log.Println("charger - node registered")
			registerCancel()
			return
		case <-time.After(5 * time.Second):
			registerCancel()
		case <-registerContext.Done():
			registerCancel()
			return
		}
	}
}

func deregister(ctx context.Context, ddaConnector *dda.Connector, nodeId string, sensorId string) {
	registerMessage := common.RegisterNodeMessage{NodeId: nodeId, SensorId: sensorId, NodeType: common.CHARGER_NODE_TYPE, Timestamp: time.Now().Unix()}
	data, _ := json.Marshal(registerMessage)

	for {
		log.Println("charger - trying to deregister node")

		deregisterContext, deregisterCancel := context.WithCancel(ctx)

		result, err := ddaConnector.PublishAction(deregisterContext, api.Action{Type: common.DEREGISTER_ACTION, Id: uuid.NewString(), Source: nodeId, Params: data})
		if err != nil {
			log.Fatalln(err)
		}

		select {
		case <-result:
			log.Println("pchargerv - node deregistered")
			deregisterCancel()
			return
		case <-time.After(5 * time.Second):
			deregisterCancel()
		case <-deregisterContext.Done():
			deregisterCancel()
			return
		}
	}
}
