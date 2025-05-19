package controller

import (
	"context"
	"fmt"
	"io"
	"log"
	"math"
	"os"

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
	callbacks["getData"] = connector.getData
	callbacks["calculateSetPointsWithoutLimits"] = l.calculateSetPointsWithoutLimits
	callbacks["calculateSetPointsWithLimits"] = l.calculateSetPointsWithLimits
	callbacks["sendFlows"] = connector.sendFlows
	callbacks["sendSetPoints"] = connector.sendSetPoints
	if sct, err := sct.NewSCT([]io.Reader{s1, s2}, callbacks); err != nil {
		return nil, err
	} else {
		l.sct = sct
	}

	return &l, nil
}

func (l *logic) start(ctx context.Context) error {
	eventChannel = make(chan string, 100)
	//var ticker common.Ticker

	l.sct.Start(ctx)

	go func() {
		for {
			select {
			case v := <-l.connector.leaderCh(ctx):
				if v {
					log.Println("controller - I'm leader, starting logic")
					///ticker.Start(l.config.Periode, l.newRound)
				} else {
					log.Println("controller - lost leadership, stop logic")
					//ticker.Stop()
				}
			case event := <-eventChannel:
				l.sct.AddEvent(event)
			case <-ctx.Done():
				log.Printf("controller - shutdown leader channel observer")
				//ticker.Stop()
				return
			}
		}
	}()

	return nil
}

/*func (l *logic) newRound() {
	addEvent("newRound")
}*/

func (l *logic) calculateSetPointsWithoutLimits() {
	for _, sensor := range l.state.sensors {
		sensor.limit = math.MaxFloat64
	}

	l.calculateSetPointsWithLimits()
}

func (l *logic) calculateSetPointsWithLimits() {
	l.state.rootSensor.reset()
	l.state.rootSensor.setSetPoints()
}
