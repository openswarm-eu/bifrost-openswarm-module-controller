package mqtt

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/url"

	"code.siemens.com/energy-community-controller/common"
	"github.com/coatyio/dda/plog"
	"github.com/eclipse/paho.golang/autopaho"
	"github.com/eclipse/paho.golang/paho"
)

type Connector struct {
	config              *common.Config
	cliCfg              autopaho.ClientConfig
	mqttConnection      *autopaho.ConnectionManager
	router              paho.Router
	pvProductionChannel chan float64
}

func NewConnector(config *common.Config) (*Connector, error) {
	u, err := url.Parse(config.Url)
	if err != nil {
		return nil, err
	}

	connector := Connector{config: config, router: paho.NewStandardRouter()}

	connector.cliCfg = autopaho.ClientConfig{
		BrokerUrls:     []*url.URL{u},
		KeepAlive:      20,
		OnConnectionUp: func(cm *autopaho.ConnectionManager, connAck *paho.Connack) { log.Println("mqtt connection up") },
		OnConnectError: func(err error) { plog.Printf("error whilst attempting connection: %s", err) },
		ClientConfig: paho.ClientConfig{
			ClientID:      config.Id,
			Router:        connector.router,
			OnClientError: func(err error) { log.Printf("server requested disconnect: %s", err) },
			OnServerDisconnect: func(d *paho.Disconnect) {
				if d.Properties != nil {
					log.Printf("server requested disconnect: %s", d.Properties.ReasonString)
				} else {
					log.Printf("server requested disconnect; reason code: %d", d.ReasonCode)
				}
			},
		},
	}

	return &connector, nil
}

func (c *Connector) Open(ctx context.Context) error {
	connection, err := autopaho.NewConnection(ctx, c.cliCfg)
	if err != nil {
		return err
	}

	if err = connection.AwaitConnection(ctx); err != nil {
		return err
	}

	c.mqttConnection = connection

	return nil
}

func (c *Connector) Close() {
	if c.pvProductionChannel != nil {
		close(c.pvProductionChannel)
	}
	c.mqttConnection.Disconnect(context.Background())
}

func (c *Connector) PublishChargingSetPoint(ctx context.Context, chargingSetPoint float64) error {
	chargingSetPointMessage := chargingSetPointMessage{ChargingSetPoint: chargingSetPoint}
	payload, _ := json.Marshal(chargingSetPointMessage)

	_, err := c.mqttConnection.Publish(ctx, &paho.Publish{
		QoS:     1,
		Topic:   fmt.Sprintf("%s/%s", c.config.Id, charging_set_point_topic),
		Payload: payload,
	})

	return err
}

func (c *Connector) SubscribeToPvProduction(ctx context.Context) (<-chan float64, error) {
	c.pvProductionChannel = make(chan float64)
	topic := fmt.Sprintf("%s/%s", c.config.Id, production_topic)

	c.router.RegisterHandler(topic, func(p *paho.Publish) {
		var msg pvProductionMessage
		if err := json.Unmarshal(p.Payload, &msg); err != nil {
			log.Printf("Could not unmarshal incomming pv production message, %s", err)
			return
		}
		c.pvProductionChannel <- msg.Production
	})

	if _, err := c.mqttConnection.Subscribe(ctx, &paho.Subscribe{Subscriptions: []paho.SubscribeOptions{{Topic: topic, QoS: 1}}}); err != nil {
		return nil, err
	}

	return c.pvProductionChannel, nil
}

const production_topic = "production"

type pvProductionMessage struct {
	Production float64 `json:"production"`
}

const charging_set_point_topic = "chargingSetPoint"

type chargingSetPointMessage struct {
	ChargingSetPoint float64 `json:"chargingSetPoint"`
}
