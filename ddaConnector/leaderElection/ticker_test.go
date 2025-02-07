package leaderElection

import (
	"testing"
	"time"
)

func TestTicker(t *testing.T) {
	subject := ticker{}
	count := 0

	subject.start(time.Millisecond*50, func() {
		count++
	})

	time.Sleep(time.Millisecond * 20)
	if count != 1 {
		t.Errorf("Wrong number of invocations: %v", count)
		t.Fail()
	}

	time.Sleep(time.Millisecond * 100)

	if count != 3 {
		t.Errorf("Wrong number of invocations: %v", count)
		t.Fail()
	}
}

func TestTickerStop(t *testing.T) {
	subject := ticker{}
	count := 0

	subject.start(time.Millisecond*100, func() {
		count++
	})
	time.Sleep(time.Millisecond * 50)
	subject.stop()
	time.Sleep(time.Millisecond * 100)

	if count != 1 {
		t.Errorf("Wrong number of invocations: %v", count)
	}
}

func TestTickerStopInsideCallback(t *testing.T) {
	subject := ticker{}
	count := 0

	subject.start(time.Millisecond*100, func() {
		count++
		subject.stop()
	})

	time.Sleep(time.Millisecond * 150)

	if count != 1 {
		t.Errorf("Wrong number of invocations: %v", count)
	}
}

func TestTickerStopAfterStop(t *testing.T) {
	subject := ticker{}
	count := 0

	subject.start(time.Millisecond*100, func() {
		count++
	})
	time.Sleep(time.Millisecond * 50)
	subject.stop()
	subject.stop()
	time.Sleep(time.Millisecond * 100)

	if count != 1 {
		t.Errorf("Wrong number of invocations after stop: %v", count)
	}
}
