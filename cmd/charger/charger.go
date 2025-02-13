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
	log.Println("starting charger")

	bootstrap := flag.Bool("b", false, "bootstrap raft")
	leadershipElectionEnabled := flag.Bool("l", false, "participate in leader election")
	flag.Parse()

	cfg := ddaConnector.NewConfig()
	cfg.Url = "tcp://localhost:1883"
	cfg.Name = "charger"
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

	getChargerChannel, err := ddaClient.SubscribeGetChargers(ctx)
	if err != nil {
		log.Fatalln(err)
	}

	chargingSetPointChannel, err := ddaClient.SubscribeChargingSetPoint(ctx)
	if err != nil {
		log.Fatalln(err)
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)

	for {
		select {
		case getChargerRequest := <-getChargerChannel:
			getChargerRequest.Callback(ddaClient.CreateGetChargerResponse())
		case chargingSetPoint := <-chargingSetPointChannel:
			if chargingSetPoint.Timestamp.After(time.Now().Add(-500 * time.Millisecond)) {
				log.Printf("Got new charging set point: %d", chargingSetPoint.Value)
			} else {
				log.Println("Got too old chargingPoint, ignoring it")
			}
		case <-sigChan:
			return
		}
	}
}
