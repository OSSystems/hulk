package main

import (
	"github.com/OSSystems/hulk/core"
	"github.com/OSSystems/hulk/mqtt"
	"github.com/Sirupsen/logrus"
)

type Hulk struct {
	path   string
	broker *core.Broker
	client mqtt.MqttClient
	logger *logrus.Logger
}

func NewHulk(client mqtt.MqttClient, path string, logger *logrus.Logger) *Hulk {
	return &Hulk{
		client: client,
		path:   path,
		broker: core.NewBroker(),
		logger: logger,
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
			h.logger.Warn(err)
		}

		err = subscriber.ExpandTopics()
		if err != nil {
			h.logger.Warn(err)
		}

		if err = h.broker.Subscribe(subscriber); err != nil {
			continue
		}

		for _, topic := range subscriber.GetTopics() {
			handler := func(topic string, payload []byte) {
				h.broker.Publish(topic, payload)
			}

			h.client.Subscribe(topic, 0, handler)
		}
	}

	return nil
}
