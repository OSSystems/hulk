package main

import (
	"fmt"
	"os"
	"path/filepath"

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
	if stat, err := os.Stat(h.path); err != nil {
		return err
	} else {
		if !stat.IsDir() {
			return fmt.Errorf("Not a directory")
		}
	}

	files, err := filepath.Glob(filepath.Join(h.path, "*.yaml"))
	if err != nil {
		return err
	}

	subscribers := []*core.Subscriber{}

	for _, file := range files {
		subscriber := core.NewSubscriber()

		if err := subscriber.LoadUnit(file); err != nil {
			return err
		}

		subscribers = append(subscribers, subscriber)
	}

	return h.InitializeSubscribers(subscribers)
}

func (h *Hulk) InitializeSubscribers(subscribers []*core.Subscriber) error {
	for _, subscriber := range subscribers {
		if err := subscriber.Initialize(); err != nil {
			h.logger.Warn(err)
			continue
		}

		if err := h.broker.Subscribe(subscriber); err != nil {
			h.logger.Warn(err)
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
