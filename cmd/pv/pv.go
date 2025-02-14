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

	bootstrap := flag.Bool("b", false, "bootstrap raft")
	leadershipElectionEnabled := flag.Bool("l", false, "participate in leader election")
	flag.Parse()

	cfg := common.NewConfig()
	cfg.Url = "tcp://localhost:1883"
	cfg.Name = "pv"
	cfg.Id = uuid.NewString()
	cfg.Leader.Protocol = "raft"
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
			pvProduction = newProduction
		case getProductionRequest := <-getProductionChannel:
			getProductionRequest.Callback(ddaConnector.CreateGetProductionResponse(pvProduction))
		case <-sigChan:
			return
		}
	}
}
