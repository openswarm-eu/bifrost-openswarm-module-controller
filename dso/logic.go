package dso

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"

	"code.siemens.com/energy-community-controller/common"
	"code.siemens.com/energy-community-controller/sct"
)

var eventChannel chan string

func addEvent(event string) {
	eventChannel <- event
}

type logic struct {
	config                         Config
	connector                      *connector
	energyCommunityTopologyUpdater *energyCommunityTopologyUpdater
	state                          *state
	sensorLimitsCalculator         sensorLimitsCalculator
	sct                            *sct.SCT
}

func newLogic(config Config, connector *connector, energyCommunityTopologyUpdater *energyCommunityTopologyUpdater, state *state) (*logic, error) {
	l := logic{
		config:                         config,
		connector:                      connector,
		energyCommunityTopologyUpdater: energyCommunityTopologyUpdater,
		sensorLimitsCalculator:         newSensorLimitsCalculator(state),
		state:                          state,
	}

	s1, err := os.Open("resources/sensor1.xml")
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %v", err)
	}
	defer s1.Close()

	s2, err := os.Open("resources/sensor2.xml")
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %v", err)
	}
	defer s2.Close()

	s3, err := os.Open("resources/sensor3.xml")
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %v", err)
	}
	defer s2.Close()

	callbacks := make(map[string]func())
	callbacks["getFlowProposals"] = connector.getFlowProposals
	callbacks["getSensorMeasurements"] = connector.getSensorMeasurements
	callbacks["calculateSensorLimits"] = l.sensorLimitsCalculator.calculateSensorLimits
	callbacks["sendSensorLimits"] = connector.sendSensorLimits
	if sct, err := sct.NewSCT([]io.Reader{s1, s2, s3}, callbacks); err != nil {
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

	leaderCh := l.connector.leaderCh(ctx)

	go func() {
		for {
			select {
			case v := <-leaderCh:
				if v {
					slog.Info("dso - I'm leader, starting logic")
					l.state.leader = true
					l.state.resetEnergyCommunitySensorLimits()
					l.energyCommunityTopologyUpdater.sendUpdatesToEnergyCommunities()
					ticker.Start(l.config.Periode, l.newRound)
				} else {
					slog.Info("dso - lost leadership, stop logic")
					l.state.leader = false
					ticker.Stop()
				}
			case event := <-eventChannel:
				l.sct.AddEvent(event)
			case <-ctx.Done():
				slog.Debug("dso - shutdown leader channel observer")
				ticker.Stop()
				return
			}
		}
	}()

	return nil
}

func (l *logic) newRound() {
	l.state.topology = l.state.newTopology
	l.state.updateLocalSensorInformation()
	l.energyCommunityTopologyUpdater.sendUpdatesToEnergyCommunities()

	addEvent("newRound")
}
