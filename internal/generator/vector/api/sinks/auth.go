package sinks

type HttpAuthStrategy string

const (
	HttpAuthStrategyBasic  HttpAuthStrategy = "basic"
	HttpAuthStrategyBearer HttpAuthStrategy = "bearer"
)

type HttpAuth struct {
	Strategy HttpAuthStrategy `json:"strategy,omitempty"`
	HttpAuthBasic
	Token string `json:"token,omitempty"`
}

type HttpAuthBasic struct {
	User     string `json:"user,omitempty"`
	Password string `json:"password,omitempty"`
}
