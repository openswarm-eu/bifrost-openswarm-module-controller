package ddaConnector

import "github.com/google/uuid"

type Config struct {
	Url    string
	Name   string
	Id     string
	Leader LeaderConfig
}

type LeaderConfig struct {
	Enabled              bool
	Protocol             string
	Bootstrap            bool
	HeartbeatPeriode     int
	HeartbeatTimeoutBase int
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
			HeartbeatPeriode:     5000,
			HeartbeatTimeoutBase: 5200,
		},
	}
}
