package hulk

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/OSSystems/hulk/mqtt"
	"github.com/Sirupsen/logrus"
	"github.com/fsnotify/fsnotify"
)

// Hulk represents a Hulk instance
type Hulk struct {
	path      string
	services  []*Service
	client    mqtt.MqttClient
	handlers  map[string][]*Service
	fswatcher *fsnotify.Watcher
	logger    *logrus.Logger
}

// NewHulk initializes a new Hulk instance
func NewHulk(client mqtt.MqttClient, path string, logger *logrus.Logger) (*Hulk, error) {
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
		logger:    logger,
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
			h.logger.Warn(err)
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

// addService adds service to managed services by Hulk
func (h *Hulk) addService(service *Service) {
	h.services = append(h.services, service)

	h.logger.Info(fmt.Sprintf("Service added: %s", service.name))

	// Watch environment files for changes
	for _, file := range service.manifest.EnvironmentFiles {
		if _, err := os.Stat(file); os.IsNotExist(err) {
			continue
		}

		err := h.fswatcher.Add(file)
		if err != nil {
			h.logger.Warn(err)
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
		h.logger.Debug(fmt.Sprintf("No handlers for %s topic: unsubscribing", topic))
		h.client.Unsubscribe(topic)
	}
}

// reloadServices reloads services which depends on environment file
func (h *Hulk) reloadServices(file string) {
	for _, service := range h.services {
		for _, envfile := range service.manifest.EnvironmentFiles {
			if envfile == file {
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
					h.logger.Debug(fmt.Sprintf("Environment file %s changed", event.Name))
					h.reloadServices(event.Name)
				}
			case err := <-h.fswatcher.Errors:
				h.logger.Warn(err)
			}
		}
	}()

	<-done
}
