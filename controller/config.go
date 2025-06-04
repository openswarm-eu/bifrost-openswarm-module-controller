package controller

import (
	"time"
)

type Config struct {
	RegistrationTimeout       time.Duration
	NodeDataTimeout           time.Duration
	MaximumMessageAge         time.Duration
	DsoNewRoundTriggerTimeout time.Duration
}

func NewConfig() Config {
	return Config{
		RegistrationTimeout:       2000 * time.Millisecond,
		NodeDataTimeout:           100 * time.Millisecond,
		MaximumMessageAge:         100 * time.Millisecond,
		DsoNewRoundTriggerTimeout: 5000 * time.Millisecond,
	}
}
