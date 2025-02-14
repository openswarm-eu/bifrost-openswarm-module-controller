package leaderElection

import (
	"context"

	"github.com/coatyio/dda/dda"
	"github.com/coatyio/dda/services/state/api"
)

type raftConsistencyProvider struct {
	ddaClient *dda.Dda
}

func NewRaftConsistencyProvider() *raftConsistencyProvider {
	return &raftConsistencyProvider{}
}

func (r *raftConsistencyProvider) open(ddaClient *dda.Dda) {
	r.ddaClient = ddaClient
}

func (r *raftConsistencyProvider) observeStateChange(ctx context.Context) (<-chan api.Input, error) {
	return r.ddaClient.ObserveStateChange(ctx)
}

func (r *raftConsistencyProvider) proposeInput(ctx context.Context, in *api.Input) error {
	return r.ddaClient.ProposeInput(ctx, in)
}
