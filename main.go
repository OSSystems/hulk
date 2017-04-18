package main

import (
	"github.com/OSSystems/hulk/mqtt"
	"github.com/Sirupsen/logrus"
	MQTT "github.com/eclipse/paho.mqtt.golang"
)

func main() {
	logger := logrus.New()

	opts := MQTT.NewClientOptions()
	opts.AddBroker("tcp://localhost:1883")

	client := mqtt.NewPahoClient(opts)

	if err := client.Connect(); err != nil {
		logger.Fatal(err)
	}

	h := NewHulk(client, "/tmp/", logger)

	done := make(chan bool)

	if err := h.LoadSubscribers(); err != nil {
		logger.Fatal(err)
	}

	<-done
}
