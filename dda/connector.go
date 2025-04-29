package dda

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"code.siemens.com/energy-community-controller/common"
	"github.com/coatyio/dda/config"
	"github.com/coatyio/dda/dda"
	"github.com/coatyio/dda/services/com/api"
	"github.com/google/uuid"
)

type Connector struct {
	*dda.Dda
	cfg            *common.Config
	leaderElection *LeaderElection
}

func NewConnector(cfg *common.Config) (*Connector, error) {
	c := Connector{cfg: cfg}

	ddaConfig := config.New()
	ddaConfig.Services.Com.Url = cfg.Url
	ddaConfig.Identity.Name = cfg.Name
	ddaConfig.Identity.Id = cfg.Id
	ddaConfig.Apis.Grpc.Disabled = true
	ddaConfig.Apis.GrpcWeb.Disabled = true
	ddaConfig.Cluster = cfg.EnergyCommunityId

	if cfg.Leader.Enabled {
		ddaConfig.Services.State.Protocol = "raft"
		ddaConfig.Services.State.Disabled = false
		ddaConfig.Services.State.Bootstrap = cfg.Leader.Bootstrap
		c.leaderElection = New(ddaConfig.Identity.Id, cfg.Leader.HeartbeatPeriode, cfg.Leader.HeartbeatTimeoutBase)
	}

	var err error
	if c.Dda, err = dda.New(ddaConfig); err != nil {
		return nil, err
	}

	return &c, nil
}

func (c *Connector) Open() error {
	if err := c.Dda.Open(5 * time.Second); err != nil {
		return err
	}

	if c.leaderElection != nil {
		if err := c.leaderElection.Open(c); err != nil {
			return err
		}
	}

	return nil
}

func (c *Connector) LeaderCh(ctx context.Context) <-chan bool {
	return c.leaderElection.LeaderCh(ctx)
}

func (c *Connector) Close() {
	log.Println("DdaClient: close")
	if c.leaderElection != nil {
		c.leaderElection.Close()
		// wait for short time to let all threads shutdown/send final messages
		time.Sleep(time.Millisecond * 50)
	}

	c.Dda.Close()
}

func (c *Connector) RegisterNode(nodeId string, sendorId string, nodeType string) error {
	registerMessage := common.DdaRegisterMessage{NodeId: nodeId, SensorId: sendorId, NodeType: nodeType, Timestamp: time.Now().Unix()}
	data, err := json.Marshal(registerMessage)

	if err != nil {
		return err

	}

	event := api.Event{Type: common.REGISTER_EVENT, Id: uuid.NewString(), Source: "controller", Data: data}
	return c.Dda.PublishEvent(event)
}

func (c *Connector) DeregisterNode(nodeId string, sendorId string, nodeType string) error {
	registerMessage := common.DdaRegisterMessage{NodeId: nodeId, SensorId: sendorId, NodeType: nodeType, Timestamp: time.Now().Unix()}
	data, err := json.Marshal(registerMessage)

	if err != nil {
		return err

	}

	event := api.Event{Type: common.DEREGISTER_EVENT, Id: uuid.NewString(), Source: "controller", Data: data}
	return c.Dda.PublishEvent(event)
}
