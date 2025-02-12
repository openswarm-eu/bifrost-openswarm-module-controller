package common

import (
	"testing"
	"time"
)

func TestTimer(t *testing.T) {
	subject := Timer{}
	count := 0

	subject.Start(time.Millisecond*100, func() {
		count++
	})
	time.Sleep(time.Millisecond * 300)

	if count != 1 {
		t.Errorf("Wrong number of invocations: %v", count)
		t.Fail()
	}
}

func TestTimerStop(t *testing.T) {
	subject := Timer{}
	count := 0

	subject.Start(time.Millisecond*100, func() {
		count++
	})
	time.Sleep(time.Millisecond * 50)
	subject.Stop()
	time.Sleep(time.Millisecond * 100)

	if count != 0 {
		t.Errorf("Wrong number of invocations: %v", count)
	}
}

func TestTimerStopAfterInvocation(t *testing.T) {
	subject := Timer{}
	count := 0

	subject.Start(time.Millisecond*100, func() {
		count++
	})

	time.Sleep(time.Millisecond * 150)
	if count != 1 {
		t.Errorf("Wrong number of invocations before Stop: %v", count)
	}

	subject.Stop()
	time.Sleep(time.Millisecond * 150)

	if count != 1 {
		t.Errorf("Wrong number of invocations after Stop: %v", count)
	}
}

func TestTimerStopInsideCallback(t *testing.T) {
	subject := Timer{}
	count := 0

	subject.Start(time.Millisecond*100, func() {
		count++
		subject.Stop()
	})

	time.Sleep(time.Millisecond * 150)

	if count != 1 {
		t.Errorf("Wrong number of invocations: %v", count)
	}
}

func TestTimerStopAfterStop(t *testing.T) {
	subject := Timer{}
	count := 0

	subject.Start(time.Millisecond*100, func() {
		count++
	})

	time.Sleep(time.Millisecond * 50)
	subject.Stop()
	subject.Stop()
	time.Sleep(time.Millisecond * 100)

	if count != 0 {
		t.Errorf("Wrong number of invocations after Stop: %v", count)
	}
}

func TestTimerReset(t *testing.T) {
	subject := Timer{}
	count := 0

	subject.Start(time.Millisecond*100, func() {
		count++
	})

	time.Sleep(time.Millisecond * 50)
	subject.Reset(time.Millisecond * 100)
	time.Sleep(time.Millisecond * 70)
	if count != 0 {
		t.Errorf("Wrong number of invocations after reStart: %v", count)
	}
}
