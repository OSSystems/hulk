package hulk

import (
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/OSSystems/pkg/log"
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
	enabled     bool
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
		enabled:     false,
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
	}).Debug("loading environment variables")

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
		}).Info("unsubscribe from topic")
		s.hulk.unsubscribe(topic, s)
	}

	s.topics = s.topics[:0]

	for _, topic := range s.manifest.Topics {
		expanded, err := template.Expand(topic, s.environment)
		if err != nil {
			if ve, ok := err.(*template.VariableExpandError); ok {
				logEntry := log.WithFields(logrus.Fields{
					"service":  s.name,
					"topic":    topic,
					"variable": ve.Name,
				})

				if ve.IsOptional {
					logEntry.Warn("no value for optional variable")

					// If value of variable is optional them ignore ONLY the current topic
					continue
				} else {
					logEntry.Data["reason"] = err
					logEntry.Error("service disabled")

					// If the value of variable is required them disable service,
					// clear the topic list and ignore ALL topics from manifest
					s.enabled = false
					s.topics = s.topics[:0]

					break
				}
			}
		}

		s.topics = append(s.topics, expanded...)
	}
}

// subscribe subscribes to topics
func (s *Service) subscribe() {
	if !s.enabled {
		log.WithFields(logrus.Fields{"service": s.name}).Debug("skipping subscribe while service is disabled")
		return
	}

	for _, topic := range s.topics {
		log.WithFields(logrus.Fields{
			"service": s.name,
			"topic":   topic,
		}).Info("subscribe to topic")

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
		}).Debug("cannot find hook or it is empty")
		return nil
	}

	err := hook.execute(payload)
	if err != nil {
		return errors.Wrapf(err, "failed to execute %s", HookNameToString(name))
	}

	return nil
}
