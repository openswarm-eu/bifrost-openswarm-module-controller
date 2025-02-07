package internal

import (
	"context"
	"net/url"

	"github.com/coatyio/dda/plog"
	"github.com/eclipse/paho.golang/autopaho"
	"github.com/eclipse/paho.golang/paho"
)

type MqttConnector struct {
	cliCfg         autopaho.ClientConfig
	mqttConnection *autopaho.ConnectionManager
	subscriptions  subscriptionManager
}

func NewMqttConnector(brokerAddress string) *MqttConnector {

	u, err := url.Parse(brokerAddress)
	if err != nil {
		panic(err)
	}

	subscriptions := newSubscriptionManager()

	cliCfg := autopaho.ClientConfig{
		BrokerUrls:     []*url.URL{u},
		KeepAlive:      20,
		OnConnectionUp: func(cm *autopaho.ConnectionManager, connAck *paho.Connack) { plog.Println("mqtt connection up") },
		OnConnectError: func(err error) { plog.Printf("error whilst attempting connection: %s", err) },
		ClientConfig: paho.ClientConfig{
			ClientID: "TestClient",
			Router: paho.NewSingleHandlerRouter(func(m *paho.Publish) {
				plog.Printf("received message on topic %s; body: %s (retain: %t)", m.Topic, m.Payload, m.Retain)
				if c, err := subscriptions.get(m.Topic); err == nil {
					c <- m.Payload
				}
			}),
			OnClientError: func(err error) { plog.Printf("server requested disconnect: %s", err) },
			OnServerDisconnect: func(d *paho.Disconnect) {
				if d.Properties != nil {
					plog.Printf("server requested disconnect: %s", d.Properties.ReasonString)
				} else {
					plog.Printf("server requested disconnect; reason code: %d", d.ReasonCode)
				}
			},
		},
	}

	return &MqttConnector{cliCfg: cliCfg, subscriptions: subscriptions}
}

func (c *MqttConnector) Connect(ctx context.Context) error {
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

func (c *MqttConnector) Disconnect(ctx context.Context) {
	c.mqttConnection.Disconnect(ctx)
}

func (c *MqttConnector) Publish(ctx context.Context, topic string, msg []byte) error {
	_, err := c.mqttConnection.Publish(ctx, &paho.Publish{
		QoS:     1,
		Topic:   topic,
		Payload: msg,
	})
	return err
}

func (c *MqttConnector) Subscribe(ctx context.Context, topic string, msg []byte) (chan []byte, error) {
	ch := make(chan []byte)
	c.subscriptions.add(topic, ch)
	if _, err := c.mqttConnection.Subscribe(ctx, &paho.Subscribe{Subscriptions: []paho.SubscribeOptions{{Topic: topic, QoS: 1}}}); err != nil {
		c.subscriptions.remove(topic)
		close(ch)
		return nil, err
	}

	return ch, nil
}
