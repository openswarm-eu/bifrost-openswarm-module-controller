package controller

import (
	"context"
	"fmt"
	"io"
	"log/slog"
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
	config                   Config
	energyCommunityConnector *energyCommunityConnector
	dsoConnector             *dsoConnector
	state                    *state
	sct                      *sct.SCT

	timeoutTimer common.Timer
}

func newLogic(config Config, energyCommunityConnector *energyCommunityConnector, dsoConnector *dsoConnector, state *state) (*logic, error) {
	l := logic{
		config:                   config,
		energyCommunityConnector: energyCommunityConnector,
		dsoConnector:             dsoConnector,
		state:                    state,
	}

	s1, err := os.Open("resources/controller1.xml")
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %v", err)
	}
	defer s1.Close()

	s2, err := os.Open("resources/controller2.xml")
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %v", err)
	}
	defer s2.Close()

	s3, err := os.Open("resources/controller3.xml")
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %v", err)
	}
	defer s3.Close()

	s4, err := os.Open("resources/controller4.xml")
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %v", err)
	}
	defer s4.Close()

	callbacks := make(map[string]func())
	callbacks["getData"] = energyCommunityConnector.getData
	callbacks["calculateSetPointsWithoutLimits"] = l.calculateSetPointsWithoutLimits
	callbacks["calculateSetPointsWithLimits"] = l.calculateSetPointsWithLimits
	callbacks["sendFlowProposal"] = dsoConnector.sendFlowProposal
	callbacks["sendSetPoints"] = energyCommunityConnector.sendSetPoints
	callbacks["setLimitsToZero"] = l.setLimitsToZero
	if sct, err := sct.NewSCT([]io.Reader{s1, s2, s3, s4}, callbacks); err != nil {
		return nil, err
	} else {
		l.sct = sct
	}

	return &l, nil
}

func (l *logic) start(ctx context.Context) error {
	eventChannel = make(chan string, 100)

	l.sct.Start(ctx)

	leaderCh := l.energyCommunityConnector.leaderCh(ctx)

	go func() {
		for {
			select {
			case v := <-leaderCh:
				if v {
					slog.Info("controller - I'm leader, starting logic")
					l.state.leader = true
					if !l.state.registeredAtDso {
						l.dsoConnector.registerAtDso(ctx)
					}
					l.timeoutTimer.Start(l.config.DsoNewRoundTriggerTimeout, l.timeout)
				} else {
					slog.Info("controller - lost leadership, stop logic")
					l.state.leader = false
				}
			case event := <-eventChannel:
				l.sct.AddEvent(event)
			case <-ctx.Done():
				slog.Debug("controller - shutdown leader channel observer")
				return
			}
		}
	}()

	return nil
}

func (l *logic) setLimitsToZero() {
	l.state.toplogy.setAllSensorLimits(0)
}

func (l *logic) timeout() {
	slog.Warn("controller - dso timeout")
	l.timeoutTimer.Reset(l.config.DsoNewRoundTriggerTimeout)
	addEvent("timeout")
}

func (l *logic) calculateSetPointsWithoutLimits() {
	l.state.toplogy.setAllSensorLimits(math.MaxFloat64)
	l.calculateSetPointsWithLimits()
}

func (l *logic) calculateSetPointsWithLimits() {
	l.timeoutTimer.Reset(l.config.DsoNewRoundTriggerTimeout)
	l.state.toplogy.rootSensor.reset()
	l.state.toplogy.rootSensor.setSetPoints()
}
