package common

import "time"

type Ticker struct {
	quit    chan bool
	started bool
}

func (t *Ticker) Start(duration time.Duration, callback func()) {
	t.started = true
	t.quit = make(chan bool)

	ticker := time.NewTicker(duration)
	go callback()
	go func() {
		for {
			select {
			case <-ticker.C:
				callback()
			case <-t.quit:
				ticker.Stop()
				t.started = false
				return
			}
		}
	}()
}

func (t *Ticker) Stop() {
	if t.started {
		select {
		case t.quit <- true:
		default:
		}
		close(t.quit)
	}
	t.started = false
}
