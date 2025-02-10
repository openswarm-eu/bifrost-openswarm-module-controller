package leaderElection

import (
	"log"
	"math/rand"
	"sync"
	"time"
)

type state int

const (
	leader state = iota
	candidate
	follower
)

type event int

const (
	ownHeartbeatReceived event = iota
	differentHeartbeatReceived
	heartbeatTimeout
)

type transition func() state

type fsm struct {
	logic        logic
	currentState state
	transitions  map[state]map[event]transition

	heartbeatMonitor timer
	heartbeatSender  ticker

	timeout time.Duration

	mu sync.Mutex
}

func newFsm(logic logic, periode time.Duration, timeoutBase time.Duration) *fsm {
	f := fsm{
		logic:        logic,
		currentState: follower,
		transitions:  make(map[state]map[event]transition),
		timeout:      getRandomTimeout(timeoutBase),
	}

	f.transitions[leader] = map[event]transition{
		ownHeartbeatReceived: func() state {
			log.Println("leader: ownHeartbeatReceived --> leader")
			f.heartbeatMonitor.reset(timeoutBase)
			return leader
		},
		differentHeartbeatReceived: func() state {
			log.Println("leader: differentHeartbeatReceived --> follower")
			f.heartbeatSender.stop()
			logic.leaderCh() <- false
			f.timeout = getRandomTimeout(timeoutBase)
			f.heartbeatMonitor.reset(f.timeout)
			return follower
		},
		heartbeatTimeout: func() state {
			log.Println("leader: heartbeatTimeout --> follower")
			f.heartbeatSender.stop()
			logic.leaderCh() <- false
			f.timeout = getRandomTimeout(timeoutBase)
			f.heartbeatMonitor.reset(f.timeout)
			return follower
		},
	}

	f.transitions[candidate] = map[event]transition{
		ownHeartbeatReceived: func() state {
			log.Println("candidate: ownHeartbeatReceived --> leader")
			logic.leaderCh() <- true
			f.heartbeatMonitor.reset(f.timeout)
			return leader
		},
		differentHeartbeatReceived: func() state {
			log.Println("candidate: differentHeartbeatReceived --> follower")
			f.heartbeatSender.stop()
			f.timeout = getRandomTimeout(timeoutBase)
			f.heartbeatMonitor.reset(f.timeout)
			return follower
		},
	}

	f.transitions[follower] = map[event]transition{
		ownHeartbeatReceived: func() state {
			log.Println("follower: ownHeartbeatReceived --> follower")
			f.heartbeatMonitor.reset(f.timeout)
			return follower
		},
		differentHeartbeatReceived: func() state {
			log.Println("follower: differentHeartbeatReceived --> follower")
			f.heartbeatMonitor.reset(f.timeout)
			return follower
		},
		heartbeatTimeout: func() state {
			log.Println("follower: heartbeatTimeout --> candidate")
			f.heartbeatSender.start(periode, logic.sendHeartbeat)
			return candidate
		},
	}

	return &f
}

func (f *fsm) start() {
	f.heartbeatMonitor.start(f.timeout, f.logic.heartbeatTimeout)
}

func (f *fsm) applyEvent(event event) {
	defer f.mu.Unlock()
	f.mu.Lock()

	if transition, ok := f.transitions[f.currentState][event]; ok {
		f.currentState = transition()
	}
}

func (f *fsm) close() {
	f.heartbeatMonitor.stop()
	f.heartbeatSender.stop()
}

func getRandomTimeout(heartbeatTimeoutBase time.Duration) time.Duration {
	timeout := heartbeatTimeoutBase.Milliseconds() + int64(rand.Float64()*float64(heartbeatTimeoutBase.Milliseconds()))
	return time.Duration(timeout) * time.Millisecond
}

type logic interface {
	heartbeatTimeout()
	sendHeartbeat()
	leaderCh() chan bool
}
