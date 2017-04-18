package mqtt

type MqttClient interface {
	Connect() error
	Subscribe(topic string, qos byte, callback MqttMessageHandler) error
}

type MqttMessageHandler func(topic string, payload []byte)
