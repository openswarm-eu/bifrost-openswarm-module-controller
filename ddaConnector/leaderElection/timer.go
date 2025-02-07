package leaderElection

import (
	"time"
)

type timer struct {
	timer *time.Timer

	quit    chan bool
	started bool
}

func (t *timer) start(duration time.Duration, callback func()) {
	t.started = true
	t.quit = make(chan bool)

	t.timer = time.NewTimer(duration)

	go func() {
		select {
		case <-t.timer.C:
			go callback()
		case <-t.quit:
			if !t.timer.Stop() {
				<-t.timer.C
			}
		}
		t.started = false
	}()
}

func (t *timer) stop() {
	if t.started {
		select {
		case t.quit <- true:
		default:
		}
		close(t.quit)
	}
	t.started = false
}

func (t *timer) reset(duration time.Duration) {
	t.timer.Reset(duration)
}
