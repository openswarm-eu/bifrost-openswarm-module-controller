package dso

import (
	"time"
)

type Config struct {
	Periode           time.Duration
	WaitTimeForInputs time.Duration
}

func NewConfig() Config {
	return Config{
		Periode:           1000 * time.Millisecond,
		WaitTimeForInputs: 100 * time.Millisecond,
	}
}
