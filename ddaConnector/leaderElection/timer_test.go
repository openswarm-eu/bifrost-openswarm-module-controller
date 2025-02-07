package leaderElection

import (
	"testing"
	"time"
)

func TestTimer(t *testing.T) {
	subject := timer{}
	count := 0

	subject.start(time.Millisecond*100, func() {
		count++
	})
	time.Sleep(time.Millisecond * 300)

	if count != 1 {
		t.Errorf("Wrong number of invocations: %v", count)
		t.Fail()
	}
}

func TestTimerStop(t *testing.T) {
	subject := timer{}
	count := 0

	subject.start(time.Millisecond*100, func() {
		count++
	})
	time.Sleep(time.Millisecond * 50)
	subject.stop()
	time.Sleep(time.Millisecond * 100)

	if count != 0 {
		t.Errorf("Wrong number of invocations: %v", count)
	}
}

func TestTimerStopAfterInvocation(t *testing.T) {
	subject := timer{}
	count := 0

	subject.start(time.Millisecond*100, func() {
		count++
	})

	time.Sleep(time.Millisecond * 150)
	if count != 1 {
		t.Errorf("Wrong number of invocations before stop: %v", count)
	}

	subject.stop()
	time.Sleep(time.Millisecond * 150)

	if count != 1 {
		t.Errorf("Wrong number of invocations after stop: %v", count)
	}
}

func TestTimerStopInsideCallback(t *testing.T) {
	subject := timer{}
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

func TestTimerStopAfterStop(t *testing.T) {
	subject := timer{}
	count := 0

	subject.start(time.Millisecond*100, func() {
		count++
	})

	time.Sleep(time.Millisecond * 50)
	subject.stop()
	subject.stop()
	time.Sleep(time.Millisecond * 100)

	if count != 0 {
		t.Errorf("Wrong number of invocations after stop: %v", count)
	}
}

func TestTimerReset(t *testing.T) {
	subject := timer{}
	count := 0

	subject.start(time.Millisecond*100, func() {
		count++
	})

	time.Sleep(time.Millisecond * 50)
	subject.reset(time.Millisecond * 100)
	time.Sleep(time.Millisecond * 70)
	if count != 0 {
		t.Errorf("Wrong number of invocations after restart: %v", count)
	}
}
