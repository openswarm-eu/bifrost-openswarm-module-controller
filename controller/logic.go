package controller

import (
	"context"
	"fmt"
	"io"
	"log"
	"math"
	"os"
	"time"

	"code.siemens.com/energy-community-controller/common"
	"code.siemens.com/energy-community-controller/sct"
)

var eventChannel chan string

func addEvent(event string) {
	eventChannel <- event
}

type logic struct {
	config    common.ControllerConfig
	connector *connector
	state     *state
	sct       *sct.SCT
}

func newLogic(config common.ControllerConfig, connector *connector, state *state) (*logic, error) {
	l := logic{config: config, connector: connector, state: state}

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
	callbacks["calculateEqualAllocationSetPoints"] = l.calculateChargerPower
	callbacks["getData"] = connector.getData
	callbacks["sendSetPoints"] = connector.sendChargingSetPoints
	if sct, err := sct.NewSCT([]io.Reader{s1, s2}, callbacks); err != nil {
		return nil, err
	} else {
		l.sct = sct
	}

	return &l, nil
}

func (l *logic) start(ctx context.Context) error {
	eventChannel = make(chan string, 100)
	var ticker common.Ticker

	l.sct.Start(ctx)

	go func() {
		for {
			select {
			case v := <-l.connector.leaderCh(ctx):
				if v {
					log.Println("controller - I'm leader, starting logic")
					ticker.Start(l.config.Periode, l.newRound)
				} else {
					log.Println("controller - lost leadership, stop logic")
					ticker.Stop()
				}
			case event := <-eventChannel:
				l.sct.AddEvent(event)
			case <-ctx.Done():
				log.Printf("controller - shutdown leader channel observer")
				ticker.Stop()
				return
			}
		}
	}()

	return nil
}

func (l *logic) newRound() {
	addEvent("newRound")
}

func (l *logic) calculateChargerPower() {
	log.Println("controller -", l.state.pvProductionValues)
	log.Println("controller -", l.state.chargerRequests)

	var sumPvProduction float64
	for _, productionValue := range l.state.pvProductionValues {
		sumPvProduction += productionValue.Value
	}

	var chargingSetPoint float64
	numChargers := len(l.state.chargerRequests)
	if numChargers > 0 {
		chargingSetPoint = sumPvProduction / float64(len(l.state.chargerRequests))
	} else {
		chargingSetPoint = 0
	}

	l.state.setPoints = make([]common.Value, len(l.state.chargerRequests))

	for i, charger := range l.state.chargerRequests {
		l.state.setPoints[i] = common.Value{Message: common.Message{Id: charger.Id, Timestamp: time.Now()}, Value: chargingSetPoint}
	}
}

func (l *logic) calculateFlowProposal() {
	for _, sensor := range l.state.sensors {
		sensor.sensorLimit = math.MaxFloat64
	}

	l.calculateSetPoints()
}

func (l *logic) calculateSetPoints() {
	l.state.rootSensor.reset()
	l.state.rootSensor.setSetPoints()
}
