package common

import (
	"time"

	"github.com/google/uuid"
)

type Config struct {
	Url        string
	Name       string
	Id         string
	Leader     LeaderConfig
	Controller ControllerConfig
	Charger    ChargerConfig
}

type LeaderConfig struct {
	Enabled              bool
	Protocol             string
	Bootstrap            bool
	HeartbeatPeriode     time.Duration
	HeartbeatTimeoutBase time.Duration
}

type ControllerConfig struct {
	Algorithm         string
	Periode           time.Duration
	WaitTimeForInputs time.Duration
}

type ChargerConfig struct {
	MaximumAcceptableSetPointOffset time.Duration
}

func NewConfig() *Config {
	return &Config{
		Url:  "",
		Name: "DDA",
		Id:   uuid.NewString(),
		Leader: LeaderConfig{
			Enabled:              false,
			Protocol:             "raft",
			Bootstrap:            false,
			HeartbeatPeriode:     5000 * time.Millisecond,
			HeartbeatTimeoutBase: 5200 * time.Millisecond,
		},
		Controller: ControllerConfig{
			Algorithm:         "equal",
			Periode:           10000 * time.Millisecond,
			WaitTimeForInputs: 100 * time.Millisecond,
		},
		Charger: ChargerConfig{
			MaximumAcceptableSetPointOffset: 1000 * time.Millisecond,
		},
	}
}
