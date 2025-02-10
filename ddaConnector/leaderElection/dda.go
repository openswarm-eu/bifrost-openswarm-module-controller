package leaderElection

import (
	"context"
	"encoding/json"
	"log"

	"github.com/coatyio/dda/dda"
	comAPI "github.com/coatyio/dda/services/com/api"
	"github.com/coatyio/dda/services/state/api"
	"github.com/google/uuid"
)

type ddaConsistencyProvider struct {
	id        string
	ddaClient *dda.Dda

	observerMembershipChannel  chan api.MembershipChange
	observerStateChangeChannel chan api.Input
}

func NewDdaConsistencyProvider(id string) *ddaConsistencyProvider {
	return &ddaConsistencyProvider{
		id:                         id,
		observerMembershipChannel:  make(chan api.MembershipChange, 256),
		observerStateChangeChannel: make(chan api.Input, 256),
	}
}

func (d *ddaConsistencyProvider) open(ddaClient *dda.Dda) {
	d.ddaClient = ddaClient
}

func (d *ddaConsistencyProvider) observeStateChange(ctx context.Context) (<-chan api.Input, error) {
	evts, err := d.ddaClient.SubscribeEvent(ctx, comAPI.SubscriptionFilter{Type: stateEvent})
	if err != nil {
		return nil, err
	}

	go func() {
		for {
			select {
			case event := <-evts:
				var stateEventMsg stateEventMsg
				if err := json.Unmarshal(event.Data, &stateEventMsg); err != nil {
					log.Printf("Could not unmarshal incomming state event message, %s", err)
					continue
				}
				d.observerStateChangeChannel <- api.Input{Key: stateEventMsg.Key, Value: stateEventMsg.Value, Op: stateEventMsg.Op}
			case <-ctx.Done():
				log.Printf("shutdown state change oberserver")
				return
			}
		}
	}()

	return d.observerStateChangeChannel, nil
}

func (d *ddaConsistencyProvider) proposeInput(ctx context.Context, in *api.Input) error {
	stateEventMsg := stateEventMsg{Key: in.Key, Value: in.Value, Op: in.Op}
	data, _ := json.Marshal(stateEventMsg)
	return d.ddaClient.PublishEvent(comAPI.Event{Type: stateEvent, Source: "ddaConsistencyProvider", Id: uuid.NewString(), Data: data})
}

const stateEvent = "com.siemens.openswarm.state"

type stateEventMsg struct {
	Key   string
	Value []byte
	Op    api.InputOp
}
