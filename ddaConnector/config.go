package ddaConnector

type Config struct {
	Url    string
	Name   string
	Leader LeaderConfig
}

type LeaderConfig struct {
	Disabled             bool
	Protocol             string
	Bootstrap            bool
	HeartbeatPeriode     int
	HeartbeatTimeoutBase int
}

func NewConfig() *Config {
	return &Config{
		Url:  "",
		Name: "DDA",
		Leader: LeaderConfig{
			Disabled:             true,
			Protocol:             "raft",
			Bootstrap:            false,
			HeartbeatPeriode:     1000,
			HeartbeatTimeoutBase: 1200,
		},
	}
}
