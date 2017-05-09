package types

// Service contains response of Hulk API: GET /services
type Service struct {
	Name        string   `json:"Name" yaml:"Name"`
	Description string   `json:"Description" yaml:"Description"`
	Enabled     bool     `json:"Enabled" yaml:"Enabled"`
	Topics      []string `json:"Topics" yaml:"Topics"`
	Hooks       struct {
		OnReceive string `json:"OnReceive" yaml:"OnReceive"`
	} `json:"Hooks" yaml:"Hooks"`
}
