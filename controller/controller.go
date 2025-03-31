package controller

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"time"

	"code.siemens.com/energy-community-controller/common"
	"code.siemens.com/energy-community-controller/dda"
	"code.siemens.com/energy-community-controller/sct"
)

type Controller struct {
	config    common.ControllerConfig
	connector *dda.Connector
	ctx       context.Context

	sct                *sct.SCT
	pvProductionValues []dda.Value
	chargerIds         []dda.Message
	setPoints          []dda.Value
}

func NewController(config common.ControllerConfig, connector *dda.Connector) (*Controller, error) {
	c := Controller{config: config, connector: connector}
	switch config.Algorithm {
	case "equal":
		s1, err := os.Open("resources/simpleController1.xml")
		if err != nil {
			return nil, fmt.Errorf("failed to open file: %v", err)
		}
		defer s1.Close()

		s2, err := os.Open("resources/simpleController2.xml")
		if err != nil {
			return nil, fmt.Errorf("failed to open file: %v", err)
		}
		defer s2.Close()

		callbacks := make(map[string]func())
		callbacks["calculateEqualAllocationSetPoints"] = c.calculateChargerPower
		callbacks["getData"] = c.getData
		callbacks["sendSetPoints"] = c.sendChargingSetPoints
		if sct, err := sct.NewSCT([]io.Reader{s1, s2}, callbacks); err != nil {
			return nil, err
		} else {
			c.sct = sct
		}
	default:
		return nil, errors.New("unknown allocation algorithm")
	}
	return &c, nil
}

func (c *Controller) Start(ctx context.Context) {
	c.ctx = ctx
	var ticker common.Ticker

	c.sct.Start(ctx)

	go func() {
		for {
			select {
			case v := <-c.connector.LeaderCh():
				if v {
					log.Println("controller - I'm leader, starting logic")
					ticker.Start(c.config.Periode, c.newRound)
				} else {
					log.Println("controller - lost leadership, stop logic")
					ticker.Stop()
				}
			case <-ctx.Done():
				log.Printf("controller - shutdown leader channel observer")
				ticker.Stop()
				return
			}
		}
	}()
}

func (c *Controller) newRound() {
	c.sct.AddEvent("newRound")
}

func (c *Controller) getData() {
	go func() {
		ctx, cancel := context.WithCancel(c.ctx)
		productionResponses, err := c.connector.GetProduction(ctx)
		if err != nil {
			log.Printf("controller - could not get PV production - %s", err)
			cancel()
			return
		}

		chargerResponses, err := c.connector.GetChargers(ctx)
		if err != nil {
			log.Printf("controller - could not get available chargers - %s", err)
			cancel()
			return
		}

		c.pvProductionValues = make([]dda.Value, 0)
		c.chargerIds = make([]dda.Message, 0)

		// to get an "AfterEqual()", subtract the minimal timeresolution of message timestamps (unix time - which are in seconds)
		startTime := time.Now().Add(-1 * time.Second)
		go func() {
			for pvResponse := range productionResponses {
				if pvResponse.Timestamp.After(startTime) {
					c.pvProductionValues = append(c.pvProductionValues, pvResponse)
				}
			}
		}()

		go func() {
			for chargerId := range chargerResponses {
				if chargerId.Timestamp.After(startTime) {
					c.chargerIds = append(c.chargerIds, chargerId)
				}
			}
		}()

		<-time.After(c.config.WaitTimeForInputs)
		cancel()

		c.sct.AddEvent("dataReceived")
	}()
}

func (c *Controller) calculateChargerPower() {
	log.Println("controller -", c.pvProductionValues)
	log.Println("controller -", c.chargerIds)

	var sumPvProduction float64
	for _, productionValue := range c.pvProductionValues {
		sumPvProduction += productionValue.Value
	}

	var chargingSetPoint float64
	numChargers := len(c.chargerIds)
	if numChargers > 0 {
		chargingSetPoint = sumPvProduction / float64(len(c.chargerIds))
	} else {
		chargingSetPoint = 0
	}

	c.setPoints = make([]dda.Value, len(c.chargerIds))

	for i, charger := range c.chargerIds {
		c.setPoints[i] = dda.Value{Message: dda.Message{Id: charger.Id, Timestamp: time.Now()}, Value: chargingSetPoint}
	}
}

func (c *Controller) sendChargingSetPoints() {
	c.connector.SendChargingSetPoints(c.setPoints)
}
