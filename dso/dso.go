package dso

import (
	"context"

	"code.siemens.com/energy-community-controller/common"
	"code.siemens.com/energy-community-controller/dda"
)

type Dso struct {
	connector                      *connector
	logic                          *logic
	energyCommunityTopologyUpdater *energyCommunityTopologyUpdater
}

func NewDso(config Config, ddaConnector *dda.Connector) (*Dso, error) {
	state := &state{
		energyCommunities:           make(map[string]int),
		newEnergyCommunities:        make(map[string]int),
		topology:                    topology{Version: 0, Sensors: make(map[string]*sensor)},
		newTopology:                 topology{Version: 0, Sensors: make(map[string]*sensor)},
		leader:                      false,
		energyCommunitySensorLimits: make(map[string]common.EnergyCommunitySensorLimitMessage),
	}

	connector := newConnector(config, ddaConnector, state)
	energyCommunityTopologyUpdater := newEnergyCommunityTopologyUpdater(config, ddaConnector, state, connector.writeEnergyCommunityToLog)
	logic, err := newLogic(config, connector, energyCommunityTopologyUpdater, state)
	if err != nil {
		return nil, err
	}

	return &Dso{connector: connector, logic: logic, energyCommunityTopologyUpdater: energyCommunityTopologyUpdater}, nil
}

func (d *Dso) Start(ctx context.Context) error {
	d.energyCommunityTopologyUpdater.setContext(ctx)
	if err := d.connector.start(ctx); err != nil {
		return err
	}
	if err := d.logic.start(ctx); err != nil {
		return err
	}

	return nil
}
