package unit

import yaml "gopkg.in/yaml.v2"

type Subscriber struct {
	Topics           []string `yaml:"Topics"`
	ExtraTopics      string   `yaml:"ExtraTopics"`
	EnvironmentFiles []string `yaml:"EnvironmentFiles"`

	Hooks struct {
		OnSubscribe string `yaml:"OnSubscribe"`
		OnPublish   string `yaml:"OnPublish"`
	} `yaml:"Hooks"`
}

func SubscriberFromData(data []byte) (*Subscriber, error) {
	unit := &Subscriber{}

	if err := yaml.Unmarshal(data, unit); err != nil {
		return nil, err
	}

	return unit, nil
}
