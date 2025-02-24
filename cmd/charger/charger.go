package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"time"

	"code.siemens.com/energy-community-controller/common"
	"code.siemens.com/energy-community-controller/controller"
	"code.siemens.com/energy-community-controller/dda"
	"code.siemens.com/energy-community-controller/mqtt"
	"github.com/google/uuid"
)

func main() {
	log.Println("starting charger")

	var id string
	var url string
	var energyCommunityId string
	bootstrap := flag.Bool("b", false, "bootstrap raft")
	leadershipElectionEnabled := flag.Bool("l", false, "participate in leader election")
	flag.StringVar(&id, "id", uuid.NewString(), "id")
	flag.StringVar(&url, "url", "tcp://localhost:1883", "mqtt url")
	flag.StringVar(&energyCommunityId, "energyCommunityId", "energyCommunity", "energy community id")
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
	var err error

	ctx, cancel := context.WithCancel(context.Background())

	defer func() {
		log.Println("shutting down")
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
		if controller, err := controller.NewController(cfg.Controller, ddaConnector); err != nil {
			log.Fatalln(err)
		} else {
			controller.Start(ctx)
		}
	}

	if mqttConnector, err = mqtt.NewConnector(cfg); err != nil {
		log.Fatalln(err)
	}

	if err = mqttConnector.Open(ctx); err != nil {
		log.Fatalln(err)
	}

	getChargerChannel, err := ddaConnector.SubscribeGetChargers(ctx)
	if err != nil {
		log.Fatalln(err)
	}

	chargingSetPointChannel, err := ddaConnector.SubscribeChargingSetPoint(ctx)
	if err != nil {
		log.Fatalln(err)
	}

	chargingSetPointMonitorDuration := cfg.Controller.Periode + cfg.Charger.MaximumAcceptableSetPointOffset
	var chargingSetPointMonitor common.Timer
	chargingSetPointMonitor.Start(chargingSetPointMonitorDuration, func() {
		log.Println("charging set point timeout")
	})

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)

	for {
		select {
		case getChargerRequest := <-getChargerChannel:
			getChargerRequest.Callback(ddaConnector.CreateGetChargerResponse())
		case chargingSetPoint := <-chargingSetPointChannel:
			if chargingSetPoint.Timestamp.After(time.Now().Add(-cfg.Charger.MaximumAcceptableSetPointOffset)) {
				log.Printf("Got new charging set point: %f", chargingSetPoint.Value)
				chargingSetPointMonitor.Reset(chargingSetPointMonitorDuration)
				mqttConnector.PublishChargingSetPoint(ctx, chargingSetPoint.Value)
			} else {
				log.Println("Got too old charging set point, ignoring it")
				log.Printf("now: %s, got: %s", time.Now(), chargingSetPoint.Timestamp)
			}
		case <-sigChan:
			return
		}
	}
}
