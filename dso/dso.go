package dso

import (
	"context"

	"code.siemens.com/energy-community-controller/common"
	"code.siemens.com/energy-community-controller/dda"
)

type Dso struct {
	connector *connector
	logic     *logic
}

func NewDso(config Config, ddaConnector *dda.Connector) (*Dso, error) {
	state := &state{
		energyCommunities:           make([]*energyCommunity, 0),
		topology:                    topology{Version: 0, Sensors: make(map[string]sensor)},
		leader:                      false,
		localSenorInformations:      make(map[string]*localSenorInformation),
		energyCommunitySensorLimits: make(map[string]common.EnergyCommunitySensorLimitMessage),
	}

	connector := newConnector(config, ddaConnector, state)
	logic, err := newLogic(config, connector, state)
	if err != nil {
		return nil, err
	}

	return &Dso{connector: connector, logic: logic}, nil
}

func (d *Dso) Start(ctx context.Context) error {
	if err := d.connector.start(ctx); err != nil {
		return err
	}
	if err := d.logic.start(ctx); err != nil {
		return err
	}

	return nil
}

func (d *Dso) Stop() {
	d.connector.stop()
}
