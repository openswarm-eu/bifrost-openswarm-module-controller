package common

import (
	"time"
)

type Timer struct {
	timer *time.Timer

	quit    chan bool
	started bool
}

func (t *Timer) Start(duration time.Duration, callback func()) {
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

func (t *Timer) Stop() {
	if t.started {
		select {
		case t.quit <- true:
		default:
		}
		close(t.quit)
	}
	t.started = false
}

func (t *Timer) Reset(duration time.Duration) {
	t.timer.Reset(duration)
}
