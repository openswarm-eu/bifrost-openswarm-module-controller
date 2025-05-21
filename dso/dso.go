package dso

import (
	"context"

	"code.siemens.com/energy-community-controller/common"
	"code.siemens.com/energy-community-controller/dda"
	"github.com/coatyio/dda/services/com/api"
)

type Dso struct {
	connector *connector
	logic     *logic
}

// fix which type of config we need
func NewDso(config common.ControllerConfig, ddaConnector *dda.Connector) (*Dso, error) {
	state := &state{
		leader:                             false,
		topology:                           topology{Version: 0, Sensors: make(map[string][]string)},
		registerCallbacks:                  make(map[string]api.ActionCallback),
		deregisterCallbacks:                make(map[string]api.ActionCallback),
		registerEnergyCommunityCallbacks:   make(map[string]api.ActionCallback),
		deregisterEnergyCommunityCallbacks: make(map[string]api.ActionCallback),
	}

	connector := newConnector(ddaConnector, state)
	logic, err := newLogic(config, connector, state)
	if err != nil {
		return nil, err
	}

	return &Dso{connector: connector, logic: logic}, nil
}

func (c *Dso) Start(ctx context.Context) error {
	if err := c.connector.start(ctx); err != nil {
		return err
	}
	if err := c.logic.start(ctx); err != nil {
		return err
	}

	return nil
}
