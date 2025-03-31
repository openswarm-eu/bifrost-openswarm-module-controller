package leaderElection

import (
	"log"
	"math/rand"
	"sync"
	"time"

	"code.siemens.com/energy-community-controller/common"
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
	id               string
	api              leaderElectionAPI
	mu               sync.Mutex
	heartbeatMonitor common.Timer
	heartbeatSender  common.Ticker

	currentState        state
	transitions         map[state]map[event]transition
	timeout             time.Duration
	highestReceivedTerm uint64
	currentTerm         uint64
	currentLeader       string
}

func newFsm(id string, api leaderElectionAPI, periode time.Duration, timeoutBase time.Duration) *fsm {
	f := fsm{
		id:                  id,
		api:                 api,
		currentState:        follower,
		transitions:         make(map[state]map[event]transition),
		timeout:             getRandomTimeout(timeoutBase),
		highestReceivedTerm: 0,
	}

	f.transitions[leader] = map[event]transition{
		ownHeartbeatReceived: func() state {
			log.Println("leader election - leader: ownHeartbeatReceived --> leader")
			f.heartbeatMonitor.Reset(timeoutBase)
			return leader
		},
		differentHeartbeatReceived: func() state {
			log.Println("leader election - leader: differentHeartbeatReceived --> follower")
			f.heartbeatSender.Stop()
			api.leaderCh() <- false
			f.timeout = getRandomTimeout(timeoutBase)
			f.heartbeatMonitor.Reset(f.timeout)
			return follower
		},
		heartbeatTimeout: func() state {
			log.Println("leader election - leader: heartbeatTimeout --> follower")
			f.heartbeatSender.Stop()
			api.leaderCh() <- false
			f.heartbeatMonitor.Start(f.timeout, f.heartbeatTimeout)
			return follower
		},
	}

	f.transitions[candidate] = map[event]transition{
		ownHeartbeatReceived: func() state {
			log.Println("leader election - candidate: ownHeartbeatReceived --> leader")
			api.leaderCh() <- true
			f.heartbeatMonitor.Start(f.timeout, f.heartbeatTimeout)
			return leader
		},
		differentHeartbeatReceived: func() state {
			log.Println("leader election - candidate: differentHeartbeatReceived --> follower")
			f.heartbeatSender.Stop()
			f.timeout = getRandomTimeout(timeoutBase)
			f.heartbeatMonitor.Start(f.timeout, f.heartbeatTimeout)
			return follower
		},
	}

	f.transitions[follower] = map[event]transition{
		ownHeartbeatReceived: func() state {
			log.Println("leader election - follower: ownHeartbeatReceived --> follower")
			f.heartbeatMonitor.Reset(f.timeout)
			return follower
		},
		differentHeartbeatReceived: func() state {
			log.Println("leader election - follower: differentHeartbeatReceived --> follower")
			f.heartbeatMonitor.Reset(f.timeout)
			return follower
		},
		heartbeatTimeout: func() state {
			log.Println("leader election - follower: heartbeatTimeout --> candidate")
			f.currentTerm = f.highestReceivedTerm + 1
			f.heartbeatSender.Start(periode, f.sendHeartbeat)
			return candidate
		},
	}

	return &f
}

func (f *fsm) start() {
	f.heartbeatMonitor.Start(f.timeout, f.heartbeatTimeout)
}

func (f *fsm) handleHeartbeat(leaderHeartbeat leaderHeartbeat) {
	if leaderHeartbeat.Term == f.highestReceivedTerm && leaderHeartbeat.LeaderId == f.id {
		f.applyEvent(ownHeartbeatReceived)
	} else if leaderHeartbeat.Term == f.highestReceivedTerm && leaderHeartbeat.LeaderId == f.currentLeader {
		// ignore heartbeats in the same term but from different candidates
		f.applyEvent(differentHeartbeatReceived)
	} else if leaderHeartbeat.Term > f.highestReceivedTerm {
		f.highestReceivedTerm = leaderHeartbeat.Term
		f.currentLeader = leaderHeartbeat.LeaderId
		if leaderHeartbeat.LeaderId == f.id {
			f.applyEvent(ownHeartbeatReceived)
		} else {
			f.applyEvent(differentHeartbeatReceived)
		}
	} else {
		log.Printf("leader election - ignoring heartbeat! Highest reveid term: %d - current leader: %s", f.highestReceivedTerm, f.currentLeader)
	}
}

func (f *fsm) applyEvent(event event) {
	defer f.mu.Unlock()
	f.mu.Lock()

	if transition, ok := f.transitions[f.currentState][event]; ok {
		f.currentState = transition()
	}
}

func (f *fsm) close() {
	f.heartbeatMonitor.Stop()
	f.heartbeatSender.Stop()
}

func (f *fsm) heartbeatTimeout() {
	log.Println("leader election - heartbeat timeout")
	f.applyEvent(heartbeatTimeout)
}

func (f *fsm) sendHeartbeat() {
	f.api.sendHeartbeat(f.id, f.currentTerm)
}

func getRandomTimeout(heartbeatTimeoutBase time.Duration) time.Duration {
	timeout := heartbeatTimeoutBase.Milliseconds() + int64(rand.Float64()*float64(heartbeatTimeoutBase.Milliseconds()))
	return time.Duration(timeout) * time.Millisecond
}

type leaderElectionAPI interface {
	sendHeartbeat(id string, term uint64)
	leaderCh() chan bool
}
