package hulk

import yaml "gopkg.in/yaml.v2"

// Manifest represents a service manifest
type Manifest struct {
	Description      string        `yaml:"Description,omitempty"`
	Topics           []string      `yaml:"Topics"`
	EnvironmentFiles []string      `yaml:"EnvironmentFiles,omitempty"`
	Hooks            ManifestHooks `yaml:"Hooks,omitempty"`
}

// ManifestHooks represents the 'Hooks' section of a service manifest
type ManifestHooks struct {
	OnReceive string `yaml:"OnReceive,omitempty"`
}

// LoadManifest loads manifest from data
func LoadManifest(data []byte) (Manifest, error) {
	manifest := Manifest{}

	if err := yaml.Unmarshal(data, &manifest); err != nil {
		return manifest, err
	}

	return manifest, nil
}
