package controller

import (
	"context"

	"code.siemens.com/energy-community-controller/dda"
)

type Controller struct {
	energyCommunityconnector *energyCommunityConnector
	dsoConnector             *dsoConnector
	logic                    *logic

	ctx    context.Context
	cancel context.CancelFunc
}

func NewController(config Config, energyCommunityId string, ddaConnectorEnergyCommunity *dda.Connector, ddaConnectorDso *dda.Connector) (*Controller, error) {
	state := &state{
		leader:          false,
		registeredAtDso: false,
		toplogy:         newToplogy(),
	}
	energyCommunityConnector := newEnergyCommunityConnector(config, energyCommunityId, ddaConnectorEnergyCommunity, state)
	dsoConnector := newDsoConnector(config, energyCommunityId, ddaConnectorDso, energyCommunityConnector, state)
	logic, err := newLogic(config, energyCommunityConnector, dsoConnector, state)
	if err != nil {
		return nil, err
	}

	c := Controller{energyCommunityconnector: energyCommunityConnector, dsoConnector: dsoConnector, logic: logic}

	return &c, nil
}

func (c *Controller) Start() error {
	c.ctx, c.cancel = context.WithCancel(context.Background())

	if err := c.energyCommunityconnector.start(c.ctx); err != nil {
		return err
	}

	if err := c.dsoConnector.start(c.ctx); err != nil {
		return err
	}

	if err := c.logic.start(c.ctx); err != nil {
		return err
	}

	return nil
}

func (c *Controller) Stop() {
	c.cancel()
	c.dsoConnector.stop()
}
