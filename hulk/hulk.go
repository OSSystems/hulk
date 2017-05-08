package hulk

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/OSSystems/hulk/api/types"
	"github.com/OSSystems/hulk/log"
	"github.com/OSSystems/hulk/mqtt"
	"github.com/Sirupsen/logrus"
	"github.com/fsnotify/fsnotify"
	"github.com/pkg/errors"
)

// Hulk represents a Hulk instance
type Hulk struct {
	path      string
	services  []*Service
	client    mqtt.MqttClient
	handlers  map[string][]*Service
	fswatcher *fsnotify.Watcher
}

// NewHulk initializes a new Hulk instance
func NewHulk(client mqtt.MqttClient, path string) (*Hulk, error) {
	stat, err := os.Stat(path)
	if err != nil {
		return nil, err
	}

	if !stat.IsDir() {
		return nil, fmt.Errorf("%s: not a directory", path)
	}

	fswatcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	return &Hulk{
		client:    client,
		handlers:  make(map[string][]*Service),
		path:      path,
		fswatcher: fswatcher,
	}, nil
}

// LoadServices loads services from a predefined directory
func (h *Hulk) LoadServices() error {
	files, err := filepath.Glob(filepath.Join(h.path, "*.yaml"))
	if err != nil {
		return err
	}

	for _, file := range files {
		service, err := NewService(h, file)
		if err != nil {
			log.Warn(err)
			continue
		}

		h.addService(service)

		// Prepare service for subscription
		service.loadEnvironment()
		service.expandTopics()
	}

	// Subscribe to all services topics
	for _, service := range h.services {
		service.subscribe()
	}

	return nil
}

func (h *Hulk) Services() []*types.Service {
	services := []*types.Service{}

	for _, service := range h.services {
		s := &types.Service{
			Name:    service.name,
			Enabled: service.enabled,
			Topics:  service.topics,
		}

		s.Hooks.OnReceive = service.manifest.Hooks.OnReceive

		services = append(services, s)
	}

	return services
}

// addService adds service to managed services by Hulk
func (h *Hulk) addService(service *Service) {
	h.services = append(h.services, service)

	log.WithFields(logrus.Fields{"service": service.name}).Info("Service added")

	// Watch environment files for changes
	for _, file := range service.manifest.EnvironmentFiles {
		if _, err := os.Stat(file); os.IsNotExist(err) {
			// Create an empty environment file
			err := ioutil.WriteFile(file, []byte(""), 0666)
			if err != nil {
				log.WithFields(logrus.Fields{
					"service": service.name,
					"file":    file,
				}).Warn(errors.Wrapf(err, "failed to create empty environment file"))
			}
		}

		err := h.fswatcher.Add(file)
		if err != nil {
			log.Warn(err)
		}
	}
}

// subscribe subscribes to service topics
func (h *Hulk) subscribe(topic string, service *Service) error {
	callback := func(topic string, payload []byte) {
		for _, s := range h.handlers[topic] {
			s.messageHandler(topic, payload)
		}
	}

	h.handlers[topic] = append(h.handlers[topic], service)

	return h.client.Subscribe(topic, 0, callback)
}

// unsubscribe unsubscribes service from topic
func (h *Hulk) unsubscribe(topic string, service *Service) {
	for i, s := range h.handlers[topic] {
		// Remove service handler
		if service == s {
			h.handlers[topic] = append(h.handlers[topic][:i], h.handlers[topic][i+1:]...)
		}
	}

	// Unsubscribe from topic if there is no handlers for topic
	if len(h.handlers[topic]) == 0 {
		log.WithFields(logrus.Fields{"topic": topic}).Debug("No handlers for topic: unsubscribing")
		h.client.Unsubscribe(topic)
	}
}

// reloadServices reloads services which depends on environment file
func (h *Hulk) reloadServices(file string) {
	for _, service := range h.services {
		for _, envfile := range service.manifest.EnvironmentFiles {
			if envfile == file {
				service.enabled = true
				service.loadEnvironment()
				service.expandTopics()
				service.subscribe()
			}
		}
	}
}

// Run runs the Hulk main loop
func (h *Hulk) Run() {
	done := make(chan bool)

	go func() {
		for {
			select {
			case event := <-h.fswatcher.Events:
				if event.Op == fsnotify.Write {
					log.WithFields(logrus.Fields{"file": event.Name}).Debug("Environment file changed")
					h.reloadServices(event.Name)
				}
			case err := <-h.fswatcher.Errors:
				log.Warn(err)
			}
		}
	}()

	<-done
}
