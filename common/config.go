package common

import (
	"time"

	"github.com/google/uuid"
)

type Config struct {
	Url               string
	Name              string
	Id                string
	EnergyCommunityId string
	Leader            LeaderConfig
	Controller        ControllerConfig
	Charger           ChargerConfig
}

type LeaderConfig struct {
	Enabled              bool
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
		Url:               "",
		Name:              "DDA",
		Id:                uuid.NewString(),
		EnergyCommunityId: "energyCommunity",
		Leader: LeaderConfig{
			Enabled:              false,
			Bootstrap:            false,
			HeartbeatPeriode:     1000 * time.Millisecond,
			HeartbeatTimeoutBase: 1200 * time.Millisecond,
		},
		Controller: ControllerConfig{
			Algorithm:         "equal",
			Periode:           1000 * time.Millisecond,
			WaitTimeForInputs: 100 * time.Millisecond,
		},
		Charger: ChargerConfig{
			MaximumAcceptableSetPointOffset: 1000 * time.Millisecond,
		},
	}
}
