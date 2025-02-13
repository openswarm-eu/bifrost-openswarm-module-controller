package leaderElection

import (
	"context"
	"log"
	"time"

	"github.com/coatyio/dda/dda"
	"github.com/coatyio/dda/services/state/api"
)

const LEADER_KEY = "leader"

type consistencyProvider interface {
	open(ddaClient *dda.Dda)
	observeStateChange(ctx context.Context) (<-chan api.Input, error)
	proposeInput(ctx context.Context, in *api.Input) error
}

type LeaderElection struct {
	id                  string
	consistencyProvider consistencyProvider

	leaderChannel chan bool
	fsm           *fsm

	ctx    context.Context
	cancel context.CancelFunc
}

func New(id string, consistencyProvider consistencyProvider, heartbeatPeriode int, heartbeatTimeoutBase int) *LeaderElection {
	ctx, cancel := context.WithCancel(context.Background())
	le := &LeaderElection{
		id:                  id,
		consistencyProvider: consistencyProvider,
		leaderChannel:       make(chan bool, 1),
		ctx:                 ctx,
		cancel:              cancel,
	}

	le.fsm = newFsm(le, time.Millisecond*time.Duration(heartbeatPeriode), time.Millisecond*time.Duration(heartbeatTimeoutBase))

	return le
}

func (le *LeaderElection) Open(ddaClient *dda.Dda) error {
	le.consistencyProvider.open(ddaClient)

	sc, err := le.consistencyProvider.observeStateChange(le.ctx)
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
	log.Printf("%d: %s-%s", change.Op, change.Key, change.Value)

	if change.Key != LEADER_KEY {
		return
	}

	if change.Op != api.InputOpSet {
		return
	}

	if le.id == string(change.Value) {
		le.fsm.applyEvent(ownHeartbeatReceived)
	} else {
		le.fsm.applyEvent(differentHeartbeatReceived)
	}
}

func (le *LeaderElection) heartbeatTimeout() {
	log.Println("heartbeat timeout")
	le.fsm.applyEvent(heartbeatTimeout)
}

func (le *LeaderElection) sendHeartbeat() {
	log.Println("sending heartbeat")
	input := api.Input{
		Op:    api.InputOpSet,
		Key:   LEADER_KEY,
		Value: []byte(le.id),
	}

	ctx, cancel := context.WithCancel(le.ctx)
	go func() {
		if err := le.consistencyProvider.proposeInput(ctx, &input); err != nil {
			log.Printf("Could not send heartbeat: %s", err)
		}
	}()

	<-time.After(1000 * time.Millisecond)
	cancel()
}

func (le *LeaderElection) leaderCh() chan bool {
	return le.leaderChannel
}
