package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"time"

	"code.siemens.com/energy-community-controller/controller"
	"code.siemens.com/energy-community-controller/ddaConnector"
	"github.com/google/uuid"
)

func main() {
	log.Println("starting pv")

	bootstrap := flag.Bool("b", false, "bootstrap raft")
	leadershipElectionEnabled := flag.Bool("l", false, "participate in leader election")
	flag.Parse()

	cfg := ddaConnector.NewConfig()
	cfg.Url = "tcp://localhost:1883"
	cfg.Name = "pv"
	cfg.Id = uuid.NewString()
	cfg.Leader.Protocol = "dda"
	cfg.Leader.Enabled = *leadershipElectionEnabled
	cfg.Leader.Bootstrap = *bootstrap

	var ddaClient *ddaConnector.DdaClient
	var err error

	ctx, cancel := context.WithCancel(context.Background())

	defer func() {
		if ddaClient != nil {
			log.Println("shutting down")
			cancel()
			ddaClient.Close()
		}
	}()

	if ddaClient, err = ddaConnector.NewConnector(cfg); err != nil {
		log.Fatalln(err)
	}

	if err := ddaClient.Open(); err != nil {
		log.Fatalln(err)
	}

	if cfg.Leader.Enabled {
		if controller, err := controller.NewController(ddaClient, "equal"); err != nil {
			log.Fatalln(err)
		} else {
			controller.Start(ctx, 10*time.Second)
		}
	}

	/*var gridConnector *internal.MqttConnector

	defer func() {
		if gridConnector != nil {
			gridConnector.Disconnect(context.Background())
		}
	}()

	gridConnector = internal.NewMqttConnector("tcp://localhost:1883")
	if err = gridConnector.Connect(context.Background()); err != nil {
		log.Fatalln(err)
	}*/

	getProductionChannel, err := ddaClient.SubscribeGetProduction(ctx)
	if err != nil {
		log.Fatalln(err)
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)

	for {
		select {
		case getProductionRequest := <-getProductionChannel:
			getProductionRequest.Callback(ddaClient.CreateGetProductionResponse(1000))
		case <-sigChan:
			return
		}
	}
}
