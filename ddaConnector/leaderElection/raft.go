package leaderElection

import (
	"context"

	"github.com/coatyio/dda/dda"
	"github.com/coatyio/dda/services/state/api"
)

type raft struct {
	ddaClient *dda.Dda
}

func NewRaft() *raft {
	return &raft{}
}

func (r *raft) open(ddaClient *dda.Dda) error {
	r.ddaClient = ddaClient
	return nil
}

func (r *raft) observeMembershipChange(ctx context.Context) (<-chan api.MembershipChange, error) {
	return r.ddaClient.ObserveMembershipChange(ctx)
}

func (r *raft) observeStateChange(ctx context.Context) (<-chan api.Input, error) {
	return r.ddaClient.ObserveStateChange(ctx)
}

func (r *raft) proposeInput(ctx context.Context, in *api.Input) error {
	return r.ddaClient.ProposeInput(ctx, in)
}

func (r *raft) close() {}
