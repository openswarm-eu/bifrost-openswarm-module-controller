package controller

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"code.siemens.com/energy-community-controller/common"
	"code.siemens.com/energy-community-controller/dda"
	"github.com/coatyio/dda/services/com/api"
	"github.com/google/uuid"
)

type dsoConnector struct {
	energyCommunityId        string
	ddaConnector             *dda.Connector
	energyCommunityConnector *energyCommunityConnector
	state                    *state

	ctx context.Context
}

func newDsoConnector(energyCommunityId string, ddaConnector *dda.Connector, energyCommunityConnector *energyCommunityConnector, state *state) *dsoConnector {
	return &dsoConnector{
		energyCommunityId:        energyCommunityId,
		ddaConnector:             ddaConnector,
		energyCommunityConnector: energyCommunityConnector,
		state:                    state,
	}
}

func (c *dsoConnector) start(ctx context.Context) error {
	c.ctx = ctx

	return nil
}

func (c *dsoConnector) stop() {
	if c.state.registeredAtDso && c.state.clusterMembers == 1 {
		registerMessage := common.DdaRegisterEnergyCommunityMessage{EnergyCommunityId: c.energyCommunityId, Timestamp: time.Now().Unix()}
		data, _ := json.Marshal(registerMessage)

		for {
			log.Println("controller - trying to unregister energy community at DSO")

			registerContext, registerCancel := context.WithCancel(context.Background())

			result, err := c.ddaConnector.PublishAction(registerContext, api.Action{Type: common.DEREGISTER_ENERGY_COMMUNITY_ACTION, Id: uuid.NewString(), Source: c.energyCommunityId, Params: data})
			if err != nil {
				log.Fatalln(err)
			}

			select {
			case <-result:
				log.Println("controller - energy community deregistered")
				registerCancel()
				return
			case <-time.After(5 * time.Second):
				registerCancel()
			case <-registerContext.Done():
				registerCancel()
				return
			}
		}
	}
}

func (c *dsoConnector) registerAtDso(ctx context.Context) {
	go func() {
		registerMessage := common.DdaRegisterEnergyCommunityMessage{EnergyCommunityId: c.energyCommunityId, Timestamp: time.Now().Unix()}
		data, _ := json.Marshal(registerMessage)

		for {
			log.Println("controller - trying to register energy community at DSO")

			registerContext, registerCancel := context.WithCancel(ctx)

			result, err := c.ddaConnector.PublishAction(registerContext, api.Action{Type: common.REGISTER_ENERGY_COMMUNITY_ACTION, Id: uuid.NewString(), Source: c.energyCommunityId, Params: data})
			if err != nil {
				log.Fatalln(err)
			}

			select {
			case <-result:
				log.Println("controller - energy community registered")
				registerCancel()
				c.energyCommunityConnector.writeSuccessfullDsoRegistrationToLog()
				return
			case <-time.After(5 * time.Second):
				registerCancel()
			case <-registerContext.Done():
				registerCancel()
				return
			}
		}
	}()
}
