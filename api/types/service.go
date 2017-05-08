package types

// Service contains response of Hulk API: GET /services
type Service struct {
	Name   string   `json:"Name"`
	Topics []string `json:"Topics"`
	Hooks  struct {
		OnReceive string `json:"OnReceive"`
	} `json:"Hooks"`
}
