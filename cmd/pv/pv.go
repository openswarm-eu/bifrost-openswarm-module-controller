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
	log.Println("starting pv")

	var nodeId string
	var url string
	var energyCommunityId string
	var sensorId string
	bootstrap := flag.Bool("b", false, "bootstrap raft")
	leadershipElectionEnabled := flag.Bool("l", false, "participate in leader election")
	flag.StringVar(&nodeId, "id", uuid.NewString(), "node id")
	flag.StringVar(&url, "url", "tcp://localhost:1883", "mqtt url")
	flag.StringVar(&energyCommunityId, "energyCommunityId", "energyCommunity", "energy community id")
	flag.StringVar(&sensorId, "sensorId", "sensor", "sensor id")
	flag.Parse()

	cfg := common.NewConfig()
	cfg.Name = "pv"
	cfg.Url = url
	cfg.Id = nodeId
	cfg.EnergyCommunityId = energyCommunityId
	cfg.Leader.Enabled = *leadershipElectionEnabled
	cfg.Leader.Bootstrap = *bootstrap

	var ddaConnector *dda.Connector
	var mqttConnector *mqtt.Connector
	var demand float64
	var err error

	ctx, cancel := context.WithCancel(context.Background())

	defer func() {
		log.Println("shutting down")

		deregister(ctx, ddaConnector, nodeId, sensorId)
		cancel()

		if ddaConnector != nil {
			ddaConnector.Close()
		}

		if mqttConnector != nil {
			mqttConnector.Close()
		}
	}()

	if ddaConnector, err = dda.NewConnector(cfg); err != nil {
		log.Fatalln(err)
	}

	if err = ddaConnector.Open(); err != nil {
		log.Fatalln(err)
	}

	if cfg.Leader.Enabled {
		if controller, err := controller.NewController(cfg.Controller, cfg.Id, ddaConnector); err != nil {
			log.Fatalln(err)
		} else {
			if err := controller.Start(ctx); err != nil {
				log.Fatalln(err)
			}
		}
	}

	if mqttConnector, err = mqtt.NewConnector(cfg); err != nil {
		log.Fatalln(err)
	}

	if err = mqttConnector.Open(ctx); err != nil {
		log.Fatalln(err)
	}

	register(ctx, ddaConnector, nodeId, sensorId)

	demandChannel, err := mqttConnector.SubscribeToDemands(ctx)
	if err != nil {
		log.Fatalln(err)
	}

	pvDemandRequests, err := ddaConnector.SubscribeAction(ctx, api.SubscriptionFilter{Type: common.GET_PV_DEMAND_ACTION})
	if err != nil {
		log.Fatalln(err)
	}

	setPointChannel, err := ddaConnector.SubscribeEvent(ctx, api.SubscriptionFilter{Type: common.AppendId(common.SET_POINT, cfg.Id)})
	if err != nil {
		log.Fatalln(err)
	}

	setPointMonitorDuration := cfg.Controller.Periode + cfg.Charger.MaximumAcceptableSetPointOffset
	var setPointMonitor common.Timer
	setPointMonitor.Start(setPointMonitorDuration, func() {
		log.Println("pv - set point timeout")
		mqttConnector.PublishSetPoint(ctx, 0)
	})

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)

	for {
		select {
		case newDemand := <-demandChannel:
			log.Printf("pv - got new production value: %f", newDemand)
			demand = newDemand
		case pvDemandRequest := <-pvDemandRequests:
			msg := common.Value{Message: common.Message{Id: nodeId, Timestamp: time.Now()}, Value: demand}
			data, _ := json.Marshal(msg)
			pvDemandRequest.Callback(api.ActionResult{Data: data})
		case setPoint := <-setPointChannel:
			var value common.Value
			if err := json.Unmarshal(setPoint.Data, &value); err != nil {
				log.Printf("pv - could not unmarshal incoming set point, %s", err)
				continue
			}

			if value.Timestamp.After(time.Now().Add(-cfg.Charger.MaximumAcceptableSetPointOffset)) {
				log.Printf("pv - got new  set point: %f", value.Value)
				setPointMonitor.Reset(setPointMonitorDuration)
				mqttConnector.PublishSetPoint(ctx, value.Value)
			} else {
				log.Println("pv - got too old set point, ignoring it")
				log.Printf("pv - now: %s, got: %s", time.Now(), value.Timestamp)
			}
		case <-sigChan:
			return
		}
	}
}

func register(ctx context.Context, ddaConnector *dda.Connector, nodeId string, sensorId string) {
	registerMessage := common.DdaRegisterNodeMessage{NodeId: nodeId, SensorId: sensorId, NodeType: common.PV_NODE_TPYE, Timestamp: time.Now().Unix()}
	data, _ := json.Marshal(registerMessage)

	for {
		log.Println("pv - trying to register node")

		registerContext, registerCancel := context.WithCancel(ctx)

		result, err := ddaConnector.PublishAction(registerContext, api.Action{Type: common.REGISTER_ACTION, Id: uuid.NewString(), Source: nodeId, Params: data})
		if err != nil {
			log.Fatalln(err)
		}

		select {
		case <-result:
			log.Println("pv - node registered")
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
	registerMessage := common.DdaRegisterNodeMessage{NodeId: nodeId, SensorId: sensorId, NodeType: common.PV_NODE_TPYE, Timestamp: time.Now().Unix()}
	data, _ := json.Marshal(registerMessage)

	for {
		log.Println("pv - trying to deregister node")

		deregisterContext, deregisterCancel := context.WithCancel(ctx)

		result, err := ddaConnector.PublishAction(deregisterContext, api.Action{Type: common.DEREGISTER_ACTION, Id: uuid.NewString(), Source: sensorId, Params: data})
		if err != nil {
			log.Fatalln(err)
		}

		select {
		case <-result:
			log.Println("pv - node deregistered")
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
