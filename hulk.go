package main

import (
	"fmt"

	"github.com/OSSystems/hulk/core"
	"github.com/OSSystems/hulk/mqtt"
)

type Hulk struct {
	path   string
	broker *core.Broker
	client mqtt.MqttClient
}

func NewHulk(client mqtt.MqttClient, path string) *Hulk {
	return &Hulk{
		client: client,
		path:   path,
		broker: core.NewBroker(),
	}
}

func (h *Hulk) LoadSubscribers() error {
	subscribers, err := core.LoadSubscribers(h.path)
	if err != nil {
		return err
	}

	for _, subscriber := range subscribers {
		err := subscriber.LoadEnvironmentFiles()
		if err != nil {
			fmt.Println(err)
		}

		err = subscriber.ExpandTopics()
		if err != nil {
			fmt.Println(err)
		}

		if err = h.broker.Subscribe(subscriber); err != nil {
			continue
		}

		for _, topic := range subscriber.Topics {
			handler := func(topic string, payload []byte) {
				h.broker.Publish(topic, payload)
			}

			h.client.Subscribe(topic, 0, handler)
		}
	}

	return nil
}
