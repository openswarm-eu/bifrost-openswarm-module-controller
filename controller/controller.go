package controller

import (
	"context"

	"code.siemens.com/energy-community-controller/common"
	"code.siemens.com/energy-community-controller/dda"
	"github.com/coatyio/dda/services/com/api"
)

type Controller struct {
	connector *connector
	logic     *logic
}

func NewController(config common.ControllerConfig, id string, ddaConnector *dda.Connector) (*Controller, error) {
	state := &state{
		leader:              false,
		sensors:             make(map[string]*sensor),
		chargers:            make(map[string]*component),
		pvs:                 make(map[string]*component),
		rootSensor:          &sensor{id: "root", childSensors: make([]*sensor, 0)},
		registerCallbacks:   make(map[string]api.ActionCallback),
		deregisterCallbacks: make(map[string]api.ActionCallback),
	}
	connector := newConnector(config, id, ddaConnector, state)
	logic, err := newLogic(config, connector, state)
	if err != nil {
		return nil, err
	}

	c := Controller{connector: connector, logic: logic}

	return &c, nil
}

func (c *Controller) Start(ctx context.Context) error {
	if err := c.connector.start(ctx); err != nil {
		return err
	}
	if err := c.logic.start(ctx); err != nil {
		return err
	}

	return nil
}
