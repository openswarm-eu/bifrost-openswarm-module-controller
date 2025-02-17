package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"

	"code.siemens.com/energy-community-controller/common"
	"code.siemens.com/energy-community-controller/controller"
	"code.siemens.com/energy-community-controller/dda"
	"code.siemens.com/energy-community-controller/mqtt"
	"github.com/google/uuid"
)

func main() {
	log.Println("starting pv")

	var id string
	var url string
	var energyCommunityId string
	bootstrap := flag.Bool("b", false, "bootstrap raft")
	leadershipElectionEnabled := flag.Bool("l", false, "participate in leader election")
	flag.StringVar(&id, "id", uuid.NewString(), "node id")
	flag.StringVar(&url, "url", "tcp://localhost:1883", "mqtt url")
	flag.StringVar(&energyCommunityId, "energyCommunityId", "energyCommunity", "energy community id")
	flag.Parse()

	cfg := common.NewConfig()
	cfg.Name = "pv"
	cfg.Url = url
	cfg.Id = id
	cfg.EnergyCommunityId = energyCommunityId
	cfg.Leader.Enabled = *leadershipElectionEnabled
	cfg.Leader.Bootstrap = *bootstrap

	var ddaConnector *dda.Connector
	var mqttConnector *mqtt.Connector
	var pvProduction int
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

	getProductionChannel, err := ddaConnector.SubscribeGetProduction(ctx)
	if err != nil {
		log.Fatalln(err)
	}

	productionChannel, err := mqttConnector.SubscribeToPvProduction(ctx)
	if err != nil {
		log.Fatalln(err)
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)

	for {
		select {
		case newProduction := <-productionChannel:
			log.Printf("Got new production value: %d", newProduction)
			pvProduction = newProduction
		case getProductionRequest := <-getProductionChannel:
			getProductionRequest.Callback(ddaConnector.CreateGetProductionResponse(pvProduction))
		case <-sigChan:
			return
		}
	}
}
