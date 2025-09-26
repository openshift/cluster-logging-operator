package http

import (
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/api/sinks"
)

// HTTPSinkConfig represents the top-level configuration for the HTTP sink.
type HTTPSinkConfig struct {

	// id is the unique config identifier of the component
	id string

	// Type is required and must be "http".
	Type string `json:"type" yaml:"type" toml:"type"`
	// Inputs specifies the IDs of the sources or transforms that feed into this sink.
	Inputs []string `json:"inputs" yaml:"inputs" toml:"inputs"`
	// URI is the destination URL for the HTTP requests.
	URI string `json:"uri" yaml:"uri" toml:"uri"`
	// Auth holds authentication-related configuration.
	Auth *Auth `json:"auth,omitempty" yaml:"auth,omitempty" toml:"auth,omitempty"`
	// Batch configures how events are batched before being sent.
	Batch *sinks.Batch `json:"batch,omitempty" yaml:"batch,omitempty" toml:"batch,omitempty"`
	// Encoding specifies how events are encoded.
	Encoding *sinks.Encoding `json:"encoding,omitempty" yaml:"encoding,omitempty" toml:"encoding,omitempty"`
	// Request provides options for the HTTP request itself, such as headers and method.
	Request *Request `json:"request,omitempty" yaml:"request,omitempty" toml:"request,omitempty"`
}

func New(id string, inputs []string, uri string) *HTTPSinkConfig {
	return &HTTPSinkConfig{
		id:     id,
		Type:   "http",
		Inputs: inputs,
		URI:    uri,
	}
}

func (c *HTTPSinkConfig) GetID() string {
	return c.id
}

// Auth configures the authentication strategy for the sink.
type Auth struct {
	Strategy string `json:"strategy" yaml:"strategy" toml:"strategy"`

	// These fields are conditionally required based on the 'strategy' field.
	// This validation logic is typically handled by a Go validator function.
	// The declarative tags below indicate their type and constraints.

	// Fields for 'aws' strategy
	AccessKeyID     string `json:"access_key_id,omitempty" yaml:"access_key_id,omitempty" toml:"access_key_id,omitempty"`
	SecretAccessKey string `json:"secret_access_key,omitempty" yaml:"secret_access_key,omitempty" toml:"secret_access_key,omitempty"`
	AssumeRole      string `json:"assume_role,omitempty" yaml:"assume_role,omitempty" toml:"assume_role,omitempty"`
	CredentialsFile string `json:"credentials_file,omitempty" yaml:"credentials_file,omitempty" toml:"credentials_file,omitempty"`
	ExternalID      string `json:"external_id,omitempty" yaml:"external_id,omitempty" toml:"external_id,omitempty"`
	LoadTimeoutSecs uint   `json:"load_timeout_secs,omitempty" yaml:"load_timeout_secs,omitempty" toml:"load_timeout_secs,omitempty"`
	Profile         string `json:"profile,omitempty" yaml:"profile,omitempty" toml:"profile,omitempty"`
	Region          string `json:"region,omitempty" yaml:"region,omitempty" toml:"region,omitempty"`
	SessionName     string `json:"session_name,omitempty" yaml:"session_name,omitempty" toml:"session_name,omitempty"`
	SessionToken    string `json:"session_token,omitempty" yaml:"session_token,omitempty" toml:"session_token,omitempty"`
	// Fields for 'basic' strategy
	User     string `json:"user,omitempty" yaml:"user,omitempty" toml:"user,omitempty"`
	Password string `json:"password,omitempty" yaml:"password,omitempty" toml:"password,omitempty"`
	// Fields for 'bearer' strategy
	Token string `json:"token,omitempty" yaml:"token,omitempty" toml:"token,omitempty"`
}

// Request provides options for the HTTP request.
type Request struct {
	Method string `json:"method,omitempty" yaml:"method,omitempty" toml:"method,omitempty"`
	// Concurrency can be "adaptive" or a number.
	// The `int` validator only works for a number, and a string type is flexible.
	Concurrency string            `json:"concurrency,omitempty" yaml:"concurrency,omitempty" toml:"concurrency,omitempty"`
	Headers     map[string]string `json:"headers,omitempty" yaml:"headers,omitempty" toml:"headers,omitempty"`
}
