package common

import (
	"testing"
	"time"
)

func TestTicker(t *testing.T) {
	subject := Ticker{}
	count := 0

	subject.Start(time.Millisecond*50, func() {
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
	subject := Ticker{}
	count := 0

	subject.Start(time.Millisecond*100, func() {
		count++
	})
	time.Sleep(time.Millisecond * 50)
	subject.Stop()
	time.Sleep(time.Millisecond * 100)

	if count != 1 {
		t.Errorf("Wrong number of invocations: %v", count)
	}
}

func TestTickerStopAfterStop(t *testing.T) {
	subject := Ticker{}
	count := 0

	subject.Start(time.Millisecond*100, func() {
		count++
	})
	time.Sleep(time.Millisecond * 50)
	subject.Stop()
	subject.Stop()
	time.Sleep(time.Millisecond * 100)

	if count != 1 {
		t.Errorf("Wrong number of invocations after Stop: %v", count)
	}
}

func TestTickerOnlyOneActivationOnBlockingCallback(t *testing.T) {
	subject := Ticker{}
	count := 0

	subject.Start(time.Millisecond*50, func() {
		time.Sleep(time.Millisecond * 60)
		count++
	})
	time.Sleep(time.Millisecond * 100)
	subject.Stop()

	if count != 1 {
		t.Errorf("Wrong number of invocations after Stop: %v", count)
	}
}
