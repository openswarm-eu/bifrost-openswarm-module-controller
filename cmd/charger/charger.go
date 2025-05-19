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

	var id string
	var url string
	var energyCommunityId string
	var sensorId string
	bootstrap := flag.Bool("b", false, "bootstrap raft")
	leadershipElectionEnabled := flag.Bool("l", false, "participate in leader election")
	flag.StringVar(&id, "id", uuid.NewString(), "id")
	flag.StringVar(&url, "url", "tcp://localhost:1883", "mqtt url")
	flag.StringVar(&energyCommunityId, "energyCommunityId", "energyCommunity", "energy community id")
	flag.StringVar(&sensorId, "sensorId", "sensor", "sensor id")
	flag.Parse()

	cfg := common.NewConfig()
	cfg.Name = "charger"
	cfg.Url = url
	cfg.Id = id
	cfg.EnergyCommunityId = energyCommunityId
	cfg.Leader.Enabled = *leadershipElectionEnabled
	cfg.Leader.Bootstrap = *bootstrap

	var ddaConnector *dda.Connector
	var mqttConnector *mqtt.Connector
	var demand float64
	var err error

	ctx, cancel := context.WithCancel(context.Background())

	defer func() {
		log.Println("charger - shutting down")

		deregister(ctx, ddaConnector, cfg)
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

	register(ctx, ddaConnector, cfg)

	demandChannel, err := mqttConnector.SubscribeToDemands(ctx)
	if err != nil {
		log.Fatalln(err)
	}

	chargerDemandRequests, err := ddaConnector.SubscribeAction(ctx, api.SubscriptionFilter{Type: common.GET_CHARGER_DEMAND_ACTION})
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
			msg := common.Value{Message: common.Message{Id: id, Timestamp: time.Now()}, Value: demand}
			data, _ := json.Marshal(msg)
			chargerDemandRequest.Callback(api.ActionResult{Data: data})
		case setPoint := <-setPointChannel:
			var value common.Value
			if err := json.Unmarshal(setPoint.Data, &value); err != nil {
				log.Printf("charger - could not unmarshal incoming set point, %s", err)
				continue
			}

			if value.Id != id {
				continue
			}

			if value.Timestamp.After(time.Now().Add(-cfg.Charger.MaximumAcceptableSetPointOffset)) {
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

func register(ctx context.Context, ddaConnector *dda.Connector, cfg *common.Config) {
	registerContext, registerCancel := context.WithCancel(ctx)
	defer registerCancel()
	registerResponseChannel, err := ddaConnector.SubscribeEvent(registerContext, api.SubscriptionFilter{Type: common.AppendId(common.REGISTER_RESPONSE_EVENT, cfg.Id)})
	if err != nil {
		log.Fatalln(err)
	}

	for {
		log.Println("charger - trying to register node")

		err = ddaConnector.RegisterNode(cfg.Id, cfg.SensorId, common.CHARGER_NODE_TYPE)
		if err != nil {
			log.Fatalln(err)
		}

		select {
		case <-registerResponseChannel:
			log.Println("charger - node registered")
			return
		case <-time.After(5 * time.Second):
			continue
		case <-registerContext.Done():
			return
		}
	}
}

func deregister(ctx context.Context, ddaConnector *dda.Connector, cfg *common.Config) {
	deregisterContext, deregisterCancel := context.WithCancel(ctx)
	defer deregisterCancel()

	deregisterResponseChannel, err := ddaConnector.SubscribeEvent(deregisterContext, api.SubscriptionFilter{Type: common.AppendId(common.REGISTER_RESPONSE_EVENT, cfg.Id)})
	if err != nil {
		log.Fatalln(err)
	}
	for {
		log.Println("charger - trying to deregister node")

		err = ddaConnector.DeregisterNode(cfg.Id, cfg.SensorId, common.CHARGER_NODE_TYPE)
		if err != nil {
			log.Fatalln(err)
		}

		select {
		case <-deregisterResponseChannel:
			log.Println("charger - node deregistered")
			return
		case <-time.After(5 * time.Second):
			continue
		}
	}
}
