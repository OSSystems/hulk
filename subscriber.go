package main

import (
	"fmt"
	"os"
	"path/filepath"

	interpol "github.com/imkira/go-interpol"
	"github.com/joho/godotenv"

	"io/ioutil"

	yaml "gopkg.in/yaml.v2"
)

type Subscriber struct {
	Topics           []string `yaml:"Topics"`
	ExtraTopics      string   `yaml:"ExtraTopics"`
	EnvironmentFiles []string `yaml:"EnvironmentFiles"`

	SubscriberHooks `yaml:"Hooks"`
}

type SubscriberHooks struct {
	OnSubscribe string `yaml:"OnSubscribe"`
	OnPublish   string `yaml:"OnPublish"`
}

func NewSubscriber() (*Subscriber, error) {
	subscriber := &Subscriber{}
	return subscriber, nil
}

func (s *Subscriber) Receiver(topic string, payload []byte) {
	fmt.Println(topic)
}

func (s *Subscriber) LoadEnvironment() (map[string]string, error) {
	environment := make(map[string]string)

	for _, file := range s.EnvironmentFiles {
		env, err := godotenv.Read(file)
		if err != nil {
			return nil, err
		}

		for key, value := range env {
			if _, ok := environment[key]; ok {
				return nil, fmt.Errorf("Duplicated environment variable: %s", key)
			}

			environment[key] = value
		}
	}

	return environment, nil
}

func (s *Subscriber) ExpandTopics() error {
	env, err := s.LoadEnvironment()
	if err != nil {
		return err
	}

	topics := s.Topics[:0]

	for _, topic := range s.Topics {
		expanded, err := interpol.WithMap(topic, env)
		if err != nil {
			return err
		}

		topics = append(topics, expanded)
	}

	return nil
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
