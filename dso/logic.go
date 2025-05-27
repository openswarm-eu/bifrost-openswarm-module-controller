package dso

import (
	"context"
	"log"

	"code.siemens.com/energy-community-controller/common"
	"code.siemens.com/energy-community-controller/sct"
)

var eventChannel chan string

func addEvent(event string) {
	eventChannel <- event
}

type logic struct {
	config    Config
	connector *connector
	state     *state
	sct       *sct.SCT
}

func newLogic(config Config, connector *connector, state *state) (*logic, error) {
	l := logic{config: config, connector: connector, state: state}

	/*s1, err := os.Open("resources/simpleController1.xml")
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %v", err)
	}
	defer s1.Close()

	s2, err := os.Open("resources/simpleController2.xml")
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %v", err)
	}
	defer s2.Close()*/

	callbacks := make(map[string]func())
	callbacks["triggerNewRound"] = connector.triggerNewRound
	callbacks["getData"] = connector.getSensorData
	callbacks["calculateLimits"] = l.calculateLimits
	callbacks["sendLimits"] = connector.sendLimits
	/*if sct, err := sct.NewSCT([]io.Reader{s1, s2}, callbacks); err != nil {
		return nil, err
	} else {
		l.sct = sct
	}*/

	return &l, nil
}

func (l *logic) start(ctx context.Context) error {
	eventChannel = make(chan string, 100)
	var ticker common.Ticker

	//l.sct.Start(ctx)

	leaderCh := l.connector.leaderCh(ctx)

	go func() {
		for {
			select {
			case v := <-leaderCh:
				if v {
					log.Println("dso - I'm leader, starting logic")
					l.state.leader = true
					ticker.Start(l.config.Periode, l.newRound)
				} else {
					log.Println("dso - lost leadership, stop logic")
					l.state.leader = false
					ticker.Stop()
				}
			case event := <-eventChannel:
				l.sct.AddEvent(event)
			case <-ctx.Done():
				log.Printf("dso - shutdown leader channel observer")
				ticker.Stop()
				return
			}
		}
	}()

	return nil
}

func (l *logic) newRound() {
	l.state.topology = l.state.newTopology
	//addEvent("newRound")
}

func (l *logic) calculateLimits() {
}
