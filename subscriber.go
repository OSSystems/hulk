package main

import (
	"fmt"
	"os"
	"path/filepath"

	"io/ioutil"

	yaml "gopkg.in/yaml.v2"
)

type Subscriber struct {
	Topics    []string `yaml:"Topics"`
}

func NewSubscriber() (*Subscriber, error) {
	subscriber := &Subscriber{}
	return subscriber, nil
}

func (s *Subscriber) Receiver(topic string) {
	fmt.Println(topic)
}

func LoadSubscribers(path string) ([]*Subscriber, error) {
	if stat, err := os.Stat(path); err != nil {
		return nil, err
	} else {
		if !stat.IsDir() {
			return nil, fmt.Errorf("Not a directory")
		}
	}

	files, err := filepath.Glob(filepath.Join(path, "*.yaml"))
	if err != nil {
		return nil, err
	}

	subscribers := []*Subscriber{}

	for _, file := range files {
		subscriber, _ := NewSubscriber()

		data, err := ioutil.ReadFile(file)
		if err != nil {
			return nil, err
		}

		err = yaml.Unmarshal(data, subscriber)
		if err != nil {
			return nil, err
		}

		subscribers = append(subscribers, subscriber)
	}

	return subscribers, nil
}
