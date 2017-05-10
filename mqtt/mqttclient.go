package mqtt

type MqttClient interface {
	Connect() error
	Disconnect()
	IsConnected() bool
	Subscribe(topic string, qos byte, callback MqttMessageHandler) error
	Unsubscribe(topic string)
}

type MqttMessageHandler func(topic string, payload []byte)
