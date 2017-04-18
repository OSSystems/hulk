package core

import (
	"fmt"
)

type Broker struct {
	Subscriptions map[string]*Subscriber
}

func NewBroker() *Broker {
	return &Broker{
		Subscriptions: map[string]*Subscriber{},
	}
}

func (b *Broker) Subscribe(subscriber *Subscriber) error {
	for _, topic := range subscriber.topics {
		if s, ok := b.Subscriptions[topic]; ok {
			if s != subscriber {
				return fmt.Errorf("Topic \"%s\" is already subscribed by another subscriber", topic)
			}

			return fmt.Errorf("Topic \"%s\" is already subscribed by subscriber", topic)
		}

		b.Subscriptions[topic] = subscriber
	}

	return nil
}

func (b *Broker) Unsubscribe(subscriber *Subscriber, topic string) error {
	owner, ok := b.Subscriptions[topic]

	if !ok {
		return fmt.Errorf("Not-subscribed topic")
	}

	if subscriber != owner {
		return fmt.Errorf("Topic not owned by subscriber")
	}

	delete(b.Subscriptions, topic)

	return nil
}

func (b *Broker) Publish(topic string, payload []byte) {
	subscriber, _ := b.Subscriptions[topic]

	if subscriber == nil {
		return
	}

	subscriber.Receiver(topic, payload)
}
