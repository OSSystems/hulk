package types

// Service contains response of Hulk API: GET /services
type Service struct {
	Name    string   `json:"Name"`
	Enabled bool     `json:"Enabled"`
	Topics  []string `json:"Topics"`
	Hooks   struct {
		OnReceive string `json:"OnReceive"`
	} `json:"Hooks"`
}
