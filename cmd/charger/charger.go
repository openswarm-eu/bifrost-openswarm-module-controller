package main

import (
	"flag"
	"log"
	"os"
	"os/signal"

	"code.siemens.com/energy-community-controller/ddaConnector"
)

func main() {
	log.Println("starting server")

	bootstrap := flag.Bool("b", false, "bootstrap raft")
	flag.Parse()

	cfg := ddaConnector.NewConfig()
	cfg.Url = "tcp://localhost:1883"
	cfg.Name = "charger"
	cfg.Leader.Protocol = "dda"
	cfg.Leader.Disabled = false
	cfg.Leader.Bootstrap = *bootstrap

	var ddaClient *ddaConnector.DdaClient
	var err error

	defer func() {
		if ddaClient != nil {
			log.Println("shutting down")
			ddaClient.Close()
		}
	}()

	if ddaClient, err = ddaConnector.NewConnector(cfg); err != nil {
		log.Fatalln(err)
	}

	if err := ddaClient.Open(); err != nil {
		log.Fatalln(err)
	}

	lc := ddaClient.LeaderCh()

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

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)

	for {
		select {
		case v := <-lc:
			log.Printf("%v", v)
		case <-sigChan:
			return
		}
	}
}
