package leaderElection

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/coatyio/dda/dda"
	"github.com/coatyio/dda/services/state/api"
)

const LEADER_KEY = "leader"

type leaderHeartbeat struct {
	Term     uint64
	LeaderId string
}

type LeaderElection struct {
	ddaClient *dda.Dda

	leaderChannel chan bool
	fsm           *fsm

	ctx    context.Context
	cancel context.CancelFunc
}

func New(id string, heartbeatPeriode time.Duration, heartbeatTimeoutBase time.Duration) *LeaderElection {
	ctx, cancel := context.WithCancel(context.Background())
	le := &LeaderElection{
		leaderChannel: make(chan bool, 1),
		ctx:           ctx,
		cancel:        cancel,
	}

	le.fsm = newFsm(id, le, heartbeatPeriode, heartbeatTimeoutBase)
	return le
}

func (le *LeaderElection) Open(ddaClient *dda.Dda) error {
	le.ddaClient = ddaClient
	sc, err := le.ddaClient.ObserveStateChange(le.ctx)
	if err != nil {
		return err
	}

	go func() {
		for stateChange := range sc {
			le.handleStateUpdate(stateChange)
		}
	}()

	le.fsm.start()
	return nil
}

func (le *LeaderElection) LeaderCh() <-chan bool {
	return le.leaderChannel
}

func (le *LeaderElection) Close() {
	le.leaderChannel <- false
	le.fsm.close()
	le.cancel()
	close(le.leaderChannel)
}

func (le *LeaderElection) handleStateUpdate(change api.Input) {
	if change.Key != LEADER_KEY {
		return
	}

	if change.Op != api.InputOpSet {
		return
	}

	log.Printf("leader election - received heartbeat %s", change.Value)

	var leaderHeartbeat leaderHeartbeat
	if err := json.Unmarshal(change.Value, &leaderHeartbeat); err != nil {
		log.Printf("leader election - error unmarshalling leader heartbeat: %s", err)
		return
	}

	le.fsm.handleHeartbeat(leaderHeartbeat)
}

func (le *LeaderElection) sendHeartbeat(id string, term uint64) {
	log.Println("leader election - sending heartbeat")

	leaderHeartbeat := leaderHeartbeat{
		Term:     term,
		LeaderId: id,
	}

	value, _ := json.Marshal(leaderHeartbeat)

	input := api.Input{
		Op:    api.InputOpSet,
		Key:   LEADER_KEY,
		Value: value,
	}

	if err := le.ddaClient.ProposeInput(le.ctx, &input); err != nil {
		log.Printf("leader election - Could not send heartbeat: %s", err)
	}
}

func (le *LeaderElection) leaderCh() chan bool {
	return le.leaderChannel
}
