package core

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	interpol "github.com/imkira/go-interpol"
	"github.com/joho/godotenv"

	"io/ioutil"

	yaml "gopkg.in/yaml.v2"
)

type Subscriber struct {
	Topics           []string          `yaml:"Topics"`
	ExtraTopics      string            `yaml:"ExtraTopics"`
	EnvironmentFiles []string          `yaml:"EnvironmentFiles"`
	Environment      map[string]string `yaml:"-"`

	SubscriberHooks `yaml:"Hooks"`
}

type SubscriberHooks struct {
	OnSubscribe string `yaml:"OnSubscribe"`
	OnPublish   string `yaml:"OnPublish"`
}

func NewSubscriber() *Subscriber {
	return &Subscriber{
		Environment: make(map[string]string),
	}
}

func (s *Subscriber) Receiver(topic string, payload []byte) {
	s.ExecuteHook(s.SubscriberHooks.OnPublish, payload)
}

func (s *Subscriber) LoadEnvironmentFiles() error {
	for _, file := range s.EnvironmentFiles {
		err := s.LoadEnvironment(file)
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *Subscriber) LoadEnvironment(file string) error {
	env, err := godotenv.Read(file)
	if err != nil {
		return err
	}

	for key, value := range env {
		if _, ok := s.Environment[key]; ok {
			return fmt.Errorf("Duplicated environment variable: %s", key)
		}

		s.Environment[key] = value
	}

	return nil
}

func (s *Subscriber) ExpandTopics() error {
	topics := s.Topics[:0]

	for _, topic := range s.Topics {
		expanded, err := interpol.WithMap(topic, s.Environment)
		if err != nil {
			return err
		}

		topics = append(topics, expanded)
	}

	return nil
}

func (s *Subscriber) CreateHookCommand(cmdLine string) *exec.Cmd {
	args := strings.Split(cmdLine, " ")
	command := args[0]

	if len(args) > 1 {
		args = args[1:]
	}

	cmd := exec.Command(command, args...)

	for key, value := range s.Environment {
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", key, value))
	}

	return cmd
}

func (s *Subscriber) ExecuteHook(cmdLine string, payload []byte) error {
	cmd := s.CreateHookCommand(cmdLine)

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return err
	}

	defer stdin.Close()

	err = cmd.Start()
	if err != nil {
		return err
	}

	stdin.Write(payload)

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
		subscriber := NewSubscriber()

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
