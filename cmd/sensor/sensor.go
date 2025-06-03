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

	var sensorId string
	var parentSensorId string
	var limit float64
	var url string
	bootstrap := flag.Bool("b", false, "bootstrap raft")
	leadershipElectionEnabled := flag.Bool("l", false, "participate in leader election")
	flag.StringVar(&sensorId, "id", uuid.NewString(), "sensor id")
	flag.StringVar(&parentSensorId, "parentId", "", "parent sensor id")
	flag.Float64Var(&limit, "limit", 0.0, "sensor limit")
	flag.StringVar(&url, "url", "tcp://localhost:1883", "mqtt url")
	flag.Parse()

	var ddaConnector *dda.Connector
	var mqttConnector *mqtt.Connector
	var measurement float64
	var err error

	registrationTimeout := 2000 * time.Millisecond

	ctx, cancel := context.WithCancel(context.Background())

	defer func() {
		log.Println("shutting down")

		deregister(ctx, ddaConnector, sensorId, parentSensorId, registrationTimeout)
		cancel()

		if ddaConnector != nil {
			ddaConnector.Close()
		}

		if mqttConnector != nil {
			mqttConnector.Close()
		}
	}()

	ddaConfig := dda.NewConfig()
	ddaConfig.Name = "sensor"
	ddaConfig.Url = url
	ddaConfig.Id = sensorId
	ddaConfig.Cluster = "dso"
	ddaConfig.Leader.Enabled = *leadershipElectionEnabled
	ddaConfig.Leader.Bootstrap = *bootstrap

	if ddaConnector, err = dda.NewConnector(ddaConfig); err != nil {
		log.Fatalln(err)
	}

	if err = ddaConnector.Open(); err != nil {
		log.Fatalln(err)
	}

	if *leadershipElectionEnabled {
		if dso, err := dso.NewDso(dso.NewConfig(), ddaConnector); err != nil {
			log.Fatalln(err)
		} else {
			if err := dso.Start(ctx); err != nil {
				log.Fatalln(err)
			}
		}
	}

	if mqttConnector, err = mqtt.NewConnector(mqtt.Config{Url: url, Id: sensorId}); err != nil {
		log.Fatalln(err)
	}

	if err = mqttConnector.Open(ctx); err != nil {
		log.Fatalln(err)
	}

	register(ctx, ddaConnector, sensorId, parentSensorId, limit, registrationTimeout)

	measurementChannel, err := mqttConnector.SubscribeToSensorMeasurements(ctx)
	if err != nil {
		log.Fatalln(err)
	}

	measurementRequestChannel, err := ddaConnector.SubscribeAction(ctx, api.SubscriptionFilter{Type: common.GET_SENSOR_MEASUREMENT_ACTION})
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
			msg := common.Value{Message: common.Message{Id: sensorId, Timestamp: time.Now()}, Value: measurement}
			data, _ := json.Marshal(msg)
			measurementRequest.Callback(api.ActionResult{Data: data})
		case <-sigChan:
			return
		}
	}
}

func register(ctx context.Context, ddaConnector *dda.Connector, sensorId string, parentSensorId string, limit float64, registrationTimeout time.Duration) {
	registerMessage := common.RegisterSensorMessage{SensorId: sensorId, ParentSensorId: parentSensorId, Limit: limit, Timestamp: time.Now()}
	data, _ := json.Marshal(registerMessage)

	for {
		log.Println("sensor - trying to register sensor")

		registerContext, registerCancel := context.WithTimeout(
			ctx,
			registrationTimeout)

		result, err := ddaConnector.PublishAction(registerContext, api.Action{Type: common.REGISTER_ACTION, Id: uuid.NewString(), Source: sensorId, Params: data})
		if err != nil {
			log.Fatalln(err)
		}

		select {
		case <-result:
			log.Println("sensor - registered")
			registerCancel()
			return
		case <-registerContext.Done():
			if registerContext.Err() == context.Canceled {
				registerCancel()
				return
			}
		}
	}
}

func deregister(ctx context.Context, ddaConnector *dda.Connector, sensorId string, parentSensorId string, registrationTimeout time.Duration) {
	registerMessage := common.RegisterSensorMessage{SensorId: sensorId, ParentSensorId: parentSensorId, Timestamp: time.Now()}
	data, _ := json.Marshal(registerMessage)

	for {
		log.Println("sensor - trying to deregister sensor")

		deregisterContext, deregisterCancel := context.WithTimeout(
			ctx,
			registrationTimeout)

		result, err := ddaConnector.PublishAction(deregisterContext, api.Action{Type: common.DEREGISTER_ACTION, Id: uuid.NewString(), Source: sensorId, Params: data})
		if err != nil {
			log.Fatalln(err)
		}

		select {
		case <-result:
			log.Println("sensor - deregistered")
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
