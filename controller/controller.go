package controller

import (
	"context"

	"code.siemens.com/energy-community-controller/common"
	"code.siemens.com/energy-community-controller/dda"
)

type Controller struct {
	connector *connector
	logic     *logic
}

func NewController(config common.ControllerConfig, ddaConnector *dda.Connector) (*Controller, error) {
	state := &state{pvProductionValues: []common.Value{}, chargerRequests: []common.Value{}, setPoints: []common.Value{}, topology: make(map[string][]string), limits: make(map[string]float64), utilizationProposal: make(map[string]float64)}
	connector := newConnector(config, ddaConnector, state)
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
