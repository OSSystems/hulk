package hulk

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"

	interpol "github.com/imkira/go-interpol"
	"github.com/joho/godotenv"
	"github.com/pkg/errors"
)

// Service represents a Hulk service
type Service struct {
	hulk        *Hulk
	name        string
	manifest    Manifest
	topics      []string
	environment map[string]string
}

// NewService creates a new Service from manifest file
func NewService(hulk *Hulk, filename string) (*Service, error) {
	data, err := ioutil.ReadFile(filename)

	manifest, err := LoadManifest(data)
	if err != nil {
		return nil, err
	}

	return &Service{
		hulk:        hulk,
		name:        path.Base(filename),
		manifest:    manifest,
		environment: make(map[string]string),
	}, nil
}

// loadEnvironment loads environment variables from 'EnvironmentFiles' specified in the service manifest
func (s *Service) loadEnvironment() {
	s.environment = map[string]string{}

	for _, file := range s.manifest.EnvironmentFiles {
		s.loadEnvironmentFile(file)
	}
}

// loadEnvironmentFile loads specified environment file
func (s *Service) loadEnvironmentFile(file string) {
	s.hulk.logger.Info(fmt.Sprintf("[%s] loading environment variables from %s", s.name, file))

retry:
	env, err := godotenv.Read(file)
	if err != nil {
		if os.IsNotExist(err) {
			s.hulk.logger.Warn(fmt.Errorf("[%s] environment file does not exists: %s", s.name, file))

			// Create an empty environment file
			err := ioutil.WriteFile(file, []byte(""), 0666)
			if err != nil {
				s.hulk.logger.Warn(errors.Wrapf(err, "[%s] failed to create empty environment file: %s", s.name, file))
				return
			}

			// Try again
			goto retry
		} else {
			s.hulk.logger.Warn(errors.Wrapf(err, "[%s] failed to parse environment file: %s", s.name, file))
			return
		}
	}

	for key, value := range env {
		s.environment[key] = value
		s.hulk.logger.Debug(fmt.Sprintf("[%s] environment variable %s=%s loaded from %s", s.name, key, value, file))
	}
}

// expandTopics expands topics from service manifest
func (s *Service) expandTopics() {
	for _, topic := range s.topics {
		s.hulk.logger.Debug(fmt.Sprintf("[%s] unsubscribe from %s", s.name, topic))
		s.hulk.unsubscribe(topic, s)
	}

	s.topics = s.topics[:0]

	for _, topic := range s.manifest.Topics {
		expanded, err := interpol.WithMap(topic, s.environment)
		if err != nil {
			s.hulk.logger.Warn(errors.Wrapf(err, "[%s] failed to expand topic: %s", s.name, topic))
			continue
		}

		s.topics = append(s.topics, expanded)
	}
}

// subscribe subscribes to topics
func (s *Service) subscribe() {
	s.hulk.logger.Info(fmt.Sprintf("[%s] subscribing for topics", s.name))

	for _, topic := range s.topics {
		s.hulk.logger.Debug(fmt.Sprintf("[%s] subscribe %s", s.name, topic))

		err := s.hulk.subscribe(topic, s)
		if err != nil {
			s.hulk.logger.Warn(err)
		}
	}
}

// messageHandler handles received messages on topic
func (s *Service) messageHandler(topic string, payload []byte) {
	err := s.executeHook(OnReceiveHook, payload)
	if err != nil {
		s.hulk.logger.Warn(err)
	}
}

// executeHook executes hook name
func (s *Service) executeHook(name HookName, payload []byte) error {
	hook := NewHook(s, name)

	if hook == nil {
		s.hulk.logger.Debug(fmt.Sprintf("[%s] skipping hook executation: %s is empty", s.name, HookNameToString(name)))
		return nil
	}

	s.hulk.logger.Info(fmt.Sprintf("[%s] %s: %s", s.name, HookNameToString(name), hook.cmdLine()))

	err := hook.execute(payload)
	if err != nil {
		return errors.Wrapf(err, "[%s] failed to execute %s", s.name, HookNameToString(name))
	}

	return nil
}
