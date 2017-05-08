package hulk

import (
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/OSSystems/hulk/log"
	"github.com/OSSystems/hulk/template"
	"github.com/Sirupsen/logrus"
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

	basename := path.Base(filename)

	return &Service{
		hulk:        hulk,
		name:        strings.TrimSuffix(basename, filepath.Ext(basename)),
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
	log.WithFields(logrus.Fields{
		"service": s.name,
		"file":    file,
	}).Info("loading environment variables")

	env, err := godotenv.Read(file)
	if err != nil {
		if os.IsNotExist(err) {
			log.WithFields(logrus.Fields{
				"service": s.name,
				"file":    file,
			}).Warn("environment file does not exists")
		} else {
			log.WithFields(logrus.Fields{
				"service": s.name,
				"file":    file,
			}).Warn(errors.Wrapf(err, "failed to parse environment file"))
			return
		}
	}

	for key, value := range env {
		s.environment[key] = value
		log.WithFields(logrus.Fields{
			"service": s.name,
			"key":     key,
			"value":   value,
			"file":    file,
		}).Debug("environment variable loaded")
	}
}

// expandTopics expands topics from service manifest
func (s *Service) expandTopics() {
	for _, topic := range s.topics {
		log.WithFields(logrus.Fields{
			"service": s.name,
			"topic":   topic,
		}).Debug("unsubscribe from topic")
		s.hulk.unsubscribe(topic, s)
	}

	s.topics = s.topics[:0]

	for _, topic := range s.manifest.Topics {
		expanded, err := template.Expand(topic, s.environment)
		if err != nil {
			log.WithFields(logrus.Fields{
				"service": s.name,
				"topic":   topic,
			}).Warn(errors.Wrapf(err, "failed to expand topic"))
			continue
		}

		s.topics = append(s.topics, expanded...)
	}
}

// subscribe subscribes to topics
func (s *Service) subscribe() {
	log.WithFields(logrus.Fields{"service": s.name}).Info("subscribing to topics")

	for _, topic := range s.topics {
		log.WithFields(logrus.Fields{
			"service": s.name,
			"topic":   topic,
		}).Debug("subscribe to topic")

		err := s.hulk.subscribe(topic, s)
		if err != nil {
			log.Warn(err)
		}
	}
}

// messageHandler handles received messages on topic
func (s *Service) messageHandler(topic string, payload []byte) {
	err := s.executeHook(OnReceiveHook, topic, payload)
	if err != nil {
		log.Warn(err)
	}
}

// executeHook executes hook name
func (s *Service) executeHook(name HookName, topic string, payload []byte) error {
	hook := NewHook(s, name, topic)

	if hook == nil {
		log.WithFields(logrus.Fields{
			"service": s.name,
			"hook":    HookNameToString(name),
		}).Debug("skipping hook executation because it's empty")
		return nil
	}

	log.Infof("[%s] %s: %s", s.name, HookNameToString(name), hook.cmdLine())

	err := hook.execute(payload)
	if err != nil {
		return errors.Wrapf(err, "failed to execute %s", HookNameToString(name))
	}

	return nil
}
