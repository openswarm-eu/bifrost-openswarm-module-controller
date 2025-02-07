package leaderElection

import "time"

type ticker struct {
	quit    chan bool
	started bool
}

func (t *ticker) start(duration time.Duration, callback func()) {
	t.started = true
	t.quit = make(chan bool)

	ticker := time.NewTicker(duration)
	go callback()
	go func() {
		for {
			select {
			case <-ticker.C:
				go callback()
			case <-t.quit:
				ticker.Stop()
				t.started = false
				return
			}
		}
	}()
}

func (t *ticker) stop() {
	if t.started {
		select {
		case t.quit <- true:
		default:
		}
		close(t.quit)
	}
	t.started = false
}
