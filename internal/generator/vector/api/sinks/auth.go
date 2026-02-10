package sinks

type HttpAuthStrategy string

const (
	HttpAuthStrategyBasic  HttpAuthStrategy = "basic"
	HttpAuthStrategyBearer HttpAuthStrategy = "bearer"
)

type HttpAuth struct {
	Strategy HttpAuthStrategy `json:"strategy,omitempty" yaml:"strategy,omitempty" toml:"strategy,omitempty"`
	User     string           `json:"user,omitempty" yaml:"user,omitempty" toml:"user,omitempty"`
	Password string           `json:"password,omitempty" yaml:"password,omitempty" toml:"password,omitempty"`
	Token    string           `json:"token,omitempty" yaml:"token,omitempty" toml:"token,omitempty"`
}
