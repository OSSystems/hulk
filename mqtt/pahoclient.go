package mqtt

import MQTT "github.com/eclipse/paho.mqtt.golang"

type PahoClient interface {
	Connect() error
	Disconnect()
	IsConnected() bool
	Subscribe(topic string, qos byte, callback MqttMessageHandler) error
	Unsubscribe(topic string)
}

type pahoClient struct {
	mqtt MQTT.Client
}

func NewPahoClient(opts *MQTT.ClientOptions) PahoClient {
	c := pahoClient{}
	c.mqtt = MQTT.NewClient(opts)

	return c
}

func (paho pahoClient) Connect() error {
	token := paho.mqtt.Connect()
	token.Wait()

	return token.Error()
}

func (paho pahoClient) Disconnect() {
	paho.mqtt.Disconnect(250)
}

func (paho pahoClient) IsConnected() bool {
	return paho.mqtt.IsConnected()
}

func (paho pahoClient) Subscribe(topic string, qos byte, callback MqttMessageHandler) error {
	pahoCallback := func(c MQTT.Client, msg MQTT.Message) {
		callback(msg.Topic(), msg.Payload())
	}

	token := paho.mqtt.Subscribe(topic, qos, pahoCallback)
	token.Wait()

	return token.Error()
}

func (paho pahoClient) Unsubscribe(topic string) {
	paho.mqtt.Unsubscribe(topic)
}
