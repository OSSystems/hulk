package mqtt

import (
	"testing"

	MQTT "github.com/eclipse/paho.mqtt.golang"
	"github.com/stretchr/testify/assert"
)

func TestNewPahoClient(t *testing.T) {
	client := NewPahoClient(MQTT.NewClientOptions())

	assert.NotNil(t, client)
	assert.NotNil(t, client.(pahoClient).mqtt)
}

func TestPahoClientConnect(t *testing.T) {
	client := NewPahoClient(MQTT.NewClientOptions())

	assert.NoError(t, client.Connect())
}

func TestPahoClientSubscribe(t *testing.T) {
	client := NewPahoClient(MQTT.NewClientOptions())

	err := client.Subscribe("topic", 0, nil)

	assert.Error(t, err)
}
