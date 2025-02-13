package controller

import (
	"context"
	"errors"
	"log"
	"time"

	"code.siemens.com/energy-community-controller/common"
	"code.siemens.com/energy-community-controller/ddaConnector"
)

type Controller struct {
	connector           *ddaConnector.DdaClient
	allocationAlgorithm allocationLogic
	ctx                 context.Context
}

func NewController(connector *ddaConnector.DdaClient, allocationAlgorithmType string) (*Controller, error) {
	c := Controller{connector: connector}
	switch allocationAlgorithmType {
	case "equal":
		c.allocationAlgorithm = equalAllocationAlgorithm{}
	default:
		return nil, errors.New("unknown allocation algorithm")
	}
	return &c, nil
}

func (c *Controller) Start(ctx context.Context, periode time.Duration) {
	c.ctx = ctx
	var ticker common.Ticker

	go func() {
		for {
			select {
			case v := <-c.connector.LeaderCh():
				if v {
					log.Println("I'm leader, starting logic")
					ticker.Start(periode, c.logic)
				} else {
					log.Println("Lost leadership, stop logic")
					ticker.Stop()
				}
			case <-ctx.Done():
				log.Printf("shutdown leader channel observer")
				ticker.Stop()
				return
			}
		}
	}()
}

func (c *Controller) logic() {
	log.Println("execute logic loop")

	ctx, cancel := context.WithCancel(c.ctx)
	productionResponses, err := c.connector.GetProduction(ctx)
	if err != nil {
		log.Printf("Could not get PV production - %s", err)
		cancel()
		return
	}

	chargerResponses, err := c.connector.GetChargers(ctx)
	if err != nil {
		log.Printf("Could not get available chargers - %s", err)
		cancel()
		return
	}

	productions := make([]ddaConnector.Value, 0)
	chargerIds := make([]ddaConnector.Message, 0)

	startTime := time.Now().Add(-1 * time.Second)
	go func() {
		for pvResponse := range productionResponses {
			if pvResponse.Timestamp.After(startTime) {
				productions = append(productions, pvResponse)
			}
		}
	}()

	go func() {
		for chargerId := range chargerResponses {
			if chargerId.Timestamp.After(startTime) {
				chargerIds = append(chargerIds, chargerId)
			}
		}
	}()

	<-time.After(100 * time.Millisecond)
	cancel()

	setPoints := c.allocationAlgorithm.calculateChargerPower(productions, chargerIds)
	c.connector.SendChargingSetPoints(setPoints)
}

type allocationLogic interface {
	calculateChargerPower(pvProductionValues []ddaConnector.Value, chargers []ddaConnector.Message) []ddaConnector.Value
}
