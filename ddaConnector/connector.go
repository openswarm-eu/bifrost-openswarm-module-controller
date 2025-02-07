package ddaConnector

import (
	"time"

	"code.siemens.com/energy-community-controller/ddaConnector/leaderElection"
	"github.com/coatyio/dda/config"
	"github.com/coatyio/dda/dda"
)

type DdaClient struct {
	cfg            *Config
	ddaClient      *dda.Dda
	leaderElection *leaderElection.LeaderElection
}

func NewConnector(cfg *Config) (*DdaClient, error) {
	c := DdaClient{cfg: cfg}

	ddaConfig := config.New()
	ddaConfig.Services.Com.Url = cfg.Url
	ddaConfig.Identity.Name = cfg.Name
	ddaConfig.Apis.Grpc.Disabled = true
	ddaConfig.Apis.GrpcWeb.Disabled = true
	if !cfg.Leader.Disabled {
		if cfg.Leader.Protocol == "raft" {
			ddaConfig.Services.State.Protocol = cfg.Leader.Protocol
			ddaConfig.Services.State.Disabled = false
			ddaConfig.Services.State.Bootstrap = cfg.Leader.Bootstrap
			c.leaderElection = leaderElection.New(ddaConfig.Identity.Id, leaderElection.NewRaft(), cfg.Leader.HeartbeatPeriode, cfg.Leader.HeartbeatTimeoutBase)
		}
	}

	var err error
	if c.ddaClient, err = dda.New(ddaConfig); err != nil {
		return nil, err
	}

	return &c, nil
}

func (c *DdaClient) Open() error {
	if err := c.ddaClient.Open(5 * time.Second); err != nil {
		return err
	}

	if c.leaderElection != nil {
		if err := c.leaderElection.Open(c.ddaClient); err != nil {
			return err
		}
	}

	return nil
}

func (c *DdaClient) LeaderCh() <-chan bool {
	return c.leaderElection.LeaderCh()
}

func (c *DdaClient) Close() {
	if c.leaderElection != nil {
		c.leaderElection.Close()
	}

	c.ddaClient.Close()
}

// PV Communication
// Charger Communication
