package dda

import (
	"time"

	"github.com/google/uuid"
)

type Config struct {
	Url     string
	Name    string
	Id      string
	Cluster string
	Leader  LeaderConfig
}

type LeaderConfig struct {
	Enabled              bool
	Bootstrap            bool
	HeartbeatPeriode     time.Duration
	HeartbeatTimeoutBase time.Duration
}

func NewConfig() Config {
	return Config{
		Url:     "",
		Name:    "DDA",
		Id:      uuid.NewString(),
		Cluster: "cluster",
		Leader: LeaderConfig{
			Enabled:              false,
			Bootstrap:            false,
			HeartbeatPeriode:     1000 * time.Millisecond,
			HeartbeatTimeoutBase: 1200 * time.Millisecond,
		},
	}
}
