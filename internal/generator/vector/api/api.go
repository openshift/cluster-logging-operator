package api

// Api is the API vector configurations
type Api struct {
	Enabled    bool   `json:"enabled,omitempty"`
	Address    string `json:"address,omitempty"`
	Playground bool   `json:"playground,omitempty"`
}
