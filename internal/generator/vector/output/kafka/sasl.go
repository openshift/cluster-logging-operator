package kafka

const (
	SASLMechanismPlain = "PLAIN"
	SASLMechamisnSSL   = "SCRAM-SHA-256"
)

type SASL struct {
	Desc        string
	ComponentID string
	Username    string
	Password    string
	Mechanism   string
}

func (t SASL) Name() string {
	return "vectorKafkaSasl"
}

func (t SASL) Template() string {
	return `{{define "vectorKafkaSasl"}}
# {{.Desc}}
[sinks.{{.ComponentID}}.sasl]
enabled = true
username = "{{.Username}}"
password = "{{.Password}}"
mechanism = "{{.Mechanism}}"
{{end}}`
}
