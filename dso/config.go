package dso

import (
	"time"
)

type Config struct {
	Periode                              time.Duration
	FlowProposalsTimeout                 time.Duration
	SensorMeasurementsTimeout            time.Duration
	TopologyUpdateAcknowledgementTimeout time.Duration
	MaximumMessageAge                    time.Duration
}

func NewConfig() Config {
	return Config{
		Periode:                              1000 * time.Millisecond,
		FlowProposalsTimeout:                 250 * time.Millisecond,
		SensorMeasurementsTimeout:            100 * time.Millisecond,
		TopologyUpdateAcknowledgementTimeout: 100 * time.Millisecond,
		MaximumMessageAge:                    100 * time.Millisecond,
	}
}
