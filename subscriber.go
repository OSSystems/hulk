package main

import "fmt"

type Subscriber struct {
	Topics []string
}

func NewSubscriber() (*Subscriber, error) {
	subscriber := &Subscriber{}
	return subscriber, nil
}

func (s *Subscriber) Receiver(topic string) {
	fmt.Println(topic)
}
