package kafka

type LibrdkafkaOptions struct {
	ComponentID string
	InsecureTLS bool
}

func (i LibrdkafkaOptions) Name() string {
	return "kafkaInsecureTLSTemplate"
}

// Template
//Starting from librdkafka 2.0.2 ssl.endpoint.identification.algorithm defaults to https.
//That means the broker certificate Common Name must correspond to the DNS domain name.
//Revert ssl.endpoint.identification.algorithm default to "none" for compatibility
//see: https://github.com/confluentinc/librdkafka/issues/4349
func (i LibrdkafkaOptions) Template() string {
	return `{{define "` + i.Name() + `" -}}
[sinks.{{.ComponentID}}.librdkafka_options]
{{- if .InsecureTLS}}
"enable.ssl.certificate.verification" = "false"
{{- end}}
"ssl.endpoint.identification.algorithm" = "none"
{{- end}}`
}
