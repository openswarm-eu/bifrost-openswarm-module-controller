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
	"code.siemens.com/energy-community-controller/dda"
	"code.siemens.com/energy-community-controller/dso"
	"code.siemens.com/energy-community-controller/mqtt"
	"github.com/coatyio/dda/services/com/api"
	"github.com/google/uuid"
)

func main() {
	log.Println("starting sensor")

	var id string
	var parentSensorId string
	var url string
	bootstrap := flag.Bool("b", false, "bootstrap raft")
	leadershipElectionEnabled := flag.Bool("l", false, "participate in leader election")
	flag.StringVar(&id, "id", uuid.NewString(), "sensor id")
	flag.StringVar(&parentSensorId, "parentId", "", "parent sensor id")
	flag.StringVar(&url, "url", "tcp://localhost:1883", "mqtt url")
	flag.Parse()

	cfg := common.NewConfig()
	cfg.Name = "sensor"
	cfg.Url = url
	cfg.Id = id
	cfg.Leader.Enabled = *leadershipElectionEnabled
	cfg.Leader.Bootstrap = *bootstrap

	var ddaConnector *dda.Connector
	var mqttConnector *mqtt.Connector
	var measurement float64
	var err error

	ctx, cancel := context.WithCancel(context.Background())

	defer func() {
		log.Println("shutting down")

		deregister(ctx, ddaConnector, id, parentSensorId)
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
		if dso, err := dso.NewDso(cfg.Controller, ddaConnector); err != nil {
			log.Fatalln(err)
		} else {
			if err := dso.Start(ctx); err != nil {
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

	register(ctx, ddaConnector, id, parentSensorId)

	measurementChannel, err := mqttConnector.SubscribeToSensorMeasurements(ctx)
	if err != nil {
		log.Fatalln(err)
	}

	measurementRequestChannel, err := ddaConnector.SubscribeAction(ctx, api.SubscriptionFilter{Type: common.GET_MEASUREMENT_ACTION})
	if err != nil {
		log.Fatalln(err)
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)

	for {
		select {
		case msrmt := <-measurementChannel:
			log.Printf("sensor - got new measurement: %f", msrmt)
			measurement = msrmt
		case measurementRequest := <-measurementRequestChannel:
			msg := common.Value{Message: common.Message{Id: id, Timestamp: time.Now()}, Value: measurement}
			data, _ := json.Marshal(msg)
			measurementRequest.Callback(api.ActionResult{Data: data})
		case <-sigChan:
			return
		}
	}
}

func register(ctx context.Context, ddaConnector *dda.Connector, sensorId string, parentSensorId string) {
	registerMessage := common.DdaRegisterSensorMessage{SensorId: sensorId, ParentSensorId: parentSensorId, Timestamp: time.Now().Unix()}
	data, _ := json.Marshal(registerMessage)

	for {
		log.Println("sensor - trying to register sensor")

		registerContext, registerCancel := context.WithCancel(ctx)

		result, err := ddaConnector.PublishAction(registerContext, api.Action{Type: common.REGISTER_ACTION, Id: uuid.NewString(), Source: sensorId, Params: data})
		if err != nil {
			log.Fatalln(err)
		}

		select {
		case <-result:
			log.Println("sensor - sensor registered")
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

func deregister(ctx context.Context, ddaConnector *dda.Connector, sensorId string, parentSensorId string) {
	registerMessage := common.DdaRegisterSensorMessage{SensorId: sensorId, ParentSensorId: parentSensorId, Timestamp: time.Now().Unix()}
	data, _ := json.Marshal(registerMessage)

	for {
		log.Println("sensor - trying to deregister sensor")

		deregisterContext, deregisterCancel := context.WithCancel(ctx)

		result, err := ddaConnector.PublishAction(deregisterContext, api.Action{Type: common.DEREGISTER_ACTION, Id: uuid.NewString(), Source: sensorId, Params: data})
		if err != nil {
			log.Fatalln(err)
		}

		select {
		case <-result:
			log.Println("sensor - sensor deregistered")
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
