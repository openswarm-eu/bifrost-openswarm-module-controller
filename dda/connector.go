package dda

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"code.siemens.com/energy-community-controller/common"
	"code.siemens.com/energy-community-controller/dda/leaderElection"
	"github.com/coatyio/dda/config"
	"github.com/coatyio/dda/dda"
	"github.com/coatyio/dda/services/com/api"
	"github.com/google/uuid"
)

type Message struct {
	Id        string
	Timestamp time.Time
}

type Value struct {
	Message
	Value float64
}

type Connector struct {
	cfg            *common.Config
	ddaClient      *dda.Dda
	leaderElection *leaderElection.LeaderElection
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
		c.leaderElection = leaderElection.New(ddaConfig.Identity.Id, cfg.Leader.HeartbeatPeriode, cfg.Leader.HeartbeatTimeoutBase)
	}

	var err error
	if c.ddaClient, err = dda.New(ddaConfig); err != nil {
		return nil, err
	}

	return &c, nil
}

func (c *Connector) Open() error {
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

func (c *Connector) LeaderCh() <-chan bool {
	return c.leaderElection.LeaderCh()
}

func (c *Connector) Close() {
	log.Println("DdaClient: close")
	if c.leaderElection != nil {
		c.leaderElection.Close()
		// wait for short time to let all threads shutdown/send final messages
		time.Sleep(time.Millisecond * 50)
	}

	c.ddaClient.Close()
}

func (c *Connector) GetChargers(ctx context.Context) (<-chan Message, error) {
	action := api.Action{Type: CHARGER_ACTION, Id: uuid.NewString(), Source: "controller"}
	replies, err := c.ddaClient.PublishAction(ctx, action)

	if err != nil {
		return nil, err
	}

	result := make(chan Message)
	go func() {
		for reply := range replies {
			var msg ddaMessage
			if err := json.Unmarshal(reply.Data, &msg); err != nil {
				log.Printf("Could not unmarshal incoming charger message, %s", err)
				continue
			}
			result <- Message{Id: msg.Id, Timestamp: time.Unix(msg.Timestamp, 0)}
		}
		close(result)
	}()

	return result, nil
}

func (c *Connector) SubscribeGetChargers(ctx context.Context) (<-chan api.ActionWithCallback, error) {
	return c.ddaClient.SubscribeAction(ctx, api.SubscriptionFilter{Type: CHARGER_ACTION})
}

func (c *Connector) GetProduction(ctx context.Context) (<-chan Value, error) {
	action := api.Action{Type: PRODUCTION_ACTION, Id: uuid.NewString(), Source: "controller"}
	replies, err := c.ddaClient.PublishAction(ctx, action)

	if err != nil {
		return nil, err
	}

	result := make(chan Value, 256)
	go func() {
		for reply := range replies {
			var value ddaValue
			if err := json.Unmarshal(reply.Data, &value); err != nil {
				log.Printf("Could not unmarshal incoming charger message, %s", err)
				continue
			}
			result <- Value{Message: Message{Id: value.Id, Timestamp: time.Unix(value.Timestamp, 0)}, Value: value.Value}
		}
		close(result)
	}()

	return result, nil
}

func (c *Connector) SubscribeGetProduction(ctx context.Context) (<-chan api.ActionWithCallback, error) {
	return c.ddaClient.SubscribeAction(ctx, api.SubscriptionFilter{Type: PRODUCTION_ACTION})
}

func (c *Connector) SendChargingSetPoints(setPoints []Value) {
	for _, setPoint := range setPoints {
		msg := ddaValue{ddaMessage: ddaMessage{Id: setPoint.Id, Timestamp: setPoint.Timestamp.Unix()}, Value: setPoint.Value}
		data, _ := json.Marshal(msg)
		if err := c.ddaClient.PublishEvent(api.Event{Type: CHARGING_SET_POINT, Source: "ddaConsistencyProvider", Id: uuid.NewString(), Data: data}); err != nil {
			log.Printf("could not send charging set point - %s", err)
		}
	}
}

func (c *Connector) SubscribeChargingSetPoint(ctx context.Context) (<-chan Value, error) {
	events, err := c.ddaClient.SubscribeEvent(ctx, api.SubscriptionFilter{Type: CHARGING_SET_POINT})

	if err != nil {
		return nil, err
	}

	result := make(chan Value, 256)
	go func() {
		for event := range events {
			var value ddaValue
			if err := json.Unmarshal(event.Data, &value); err != nil {
				log.Printf("Could not unmarshal incoming charging set point, %s", err)
				continue
			}
			if value.Id == c.cfg.Id {
				result <- Value{Message: Message{Id: value.Id, Timestamp: time.Unix(value.Timestamp, 0)}, Value: value.Value}
			}
		}
		close(result)
	}()

	return result, nil
}

func (c *Connector) CreateGetChargerResponse() api.ActionResult {
	msg := ddaMessage{Id: c.cfg.Id, Timestamp: time.Now().Unix()}
	data, _ := json.Marshal(msg)
	return api.ActionResult{Data: data}
}

func (c *Connector) CreateGetProductionResponse(production float64) api.ActionResult {
	msg := ddaValue{ddaMessage: ddaMessage{Id: c.cfg.Id, Timestamp: time.Now().Unix()}, Value: production}
	data, _ := json.Marshal(msg)
	return api.ActionResult{Data: data}
}

const CHARGER_ACTION = "com.siemens.openswarm.charger"
const PRODUCTION_ACTION = "com.siemens.openswarm.production"
const CHARGING_SET_POINT = "com.siemens.openswarm.chargersetpoint"

type ddaMessage struct {
	Id        string
	Timestamp int64
}

type ddaValue struct {
	ddaMessage
	Value float64
}
