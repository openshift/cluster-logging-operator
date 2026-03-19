package api

// LogSchema default log schema for all events.
type LogSchema struct {
	// The name of the event field to treat as the host which sent the message.
	HostKey string `json:"host_key,omitempty" yaml:"host_key,omitempty" toml:"host_key,omitempty"`
}
