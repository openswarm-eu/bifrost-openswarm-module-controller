package mqtt

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"log/slog"
	"net/url"

	"github.com/coatyio/dda/plog"
	"github.com/eclipse/paho.golang/autopaho"
	"github.com/eclipse/paho.golang/paho"
)

type Connector struct {
	config                   Config
	cliCfg                   autopaho.ClientConfig
	mqttConnection           *autopaho.ConnectionManager
	router                   paho.Router
	demandChannel            chan float64
	sensorMeasurementChannel chan float64
}

func NewConnector(config Config) (*Connector, error) {
	u, err := url.Parse(config.Url)
	if err != nil {
		return nil, err
	}

	connector := Connector{config: config, router: paho.NewStandardRouter()}

	connector.cliCfg = autopaho.ClientConfig{
		BrokerUrls:     []*url.URL{u},
		KeepAlive:      20,
		OnConnectionUp: func(cm *autopaho.ConnectionManager, connAck *paho.Connack) { slog.Info("mqtt - connection up") },
		OnConnectError: func(err error) { plog.Printf("error whilst attempting connection: %s", err) },
		ClientConfig: paho.ClientConfig{
			ClientID:      config.Id,
			Router:        connector.router,
			OnClientError: func(err error) { log.Printf("server requested disconnect: %s", err) },
			OnServerDisconnect: func(d *paho.Disconnect) {
				if d.Properties != nil {
					slog.Info("mqtt - server requested disconnect", "reason", d.Properties.ReasonString)
				} else {
					slog.Info("mqtt - server requested disconnect", "reason_code", d.ReasonCode)
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
	if c.demandChannel != nil {
		close(c.demandChannel)
	}
	if c.sensorMeasurementChannel != nil {
		close(c.sensorMeasurementChannel)
	}
	c.mqttConnection.Disconnect(context.Background())
}

func (c *Connector) PublishSetPoint(ctx context.Context, setPoint float64) error {
	setPointMessage := setPointMessage{SetPoint: setPoint}
	payload, _ := json.Marshal(setPointMessage)

	_, err := c.mqttConnection.Publish(ctx, &paho.Publish{
		QoS:     1,
		Topic:   fmt.Sprintf("%s/%s", c.config.Id, set_point_topic),
		Payload: payload,
	})

	return err
}

func (c *Connector) SubscribeToDemands(ctx context.Context) (<-chan float64, error) {
	c.demandChannel = make(chan float64)
	topic := fmt.Sprintf("%s/%s", c.config.Id, demand_topic)

	c.router.RegisterHandler(topic, func(p *paho.Publish) {
		var msg demandMessage
		if err := json.Unmarshal(p.Payload, &msg); err != nil {
			slog.Error("mqtt - could not unmarshal incomming demand message", "error", err)
			return
		}
		c.demandChannel <- msg.Demand
	})

	if _, err := c.mqttConnection.Subscribe(ctx, &paho.Subscribe{Subscriptions: []paho.SubscribeOptions{{Topic: topic, QoS: 1}}}); err != nil {
		return nil, err
	}

	return c.demandChannel, nil
}

func (c *Connector) SubscribeToSensorMeasurements(ctx context.Context) (<-chan float64, error) {
	c.sensorMeasurementChannel = make(chan float64)
	topic := fmt.Sprintf("%s/%s", c.config.Id, measurement_topic)

	c.router.RegisterHandler(topic, func(p *paho.Publish) {
		var msg measurementMessage
		if err := json.Unmarshal(p.Payload, &msg); err != nil {
			slog.Error("mqtt - could not unmarshal incomming sensor measurement message", "error", err)
			return
		}
		c.sensorMeasurementChannel <- msg.Measurement
	})

	if _, err := c.mqttConnection.Subscribe(ctx, &paho.Subscribe{Subscriptions: []paho.SubscribeOptions{{Topic: topic, QoS: 1}}}); err != nil {
		return nil, err
	}

	return c.sensorMeasurementChannel, nil
}

const demand_topic = "demand"

type demandMessage struct {
	Demand float64 `json:"demand"`
}

const set_point_topic = "setPoint"

type setPointMessage struct {
	SetPoint float64 `json:"setPoint"`
}

const measurement_topic = "measurement"

type measurementMessage struct {
	Measurement float64 `json:"measurement"`
}
