package kafka

type InsecureTLS struct {
	ComponentID string
}

func (i InsecureTLS) Name() string {
	return "kafkaInsecureTLSTemplate"
}

func (i InsecureTLS) Template() string {
	return `{{define "` + i.Name() + `" -}}
[sinks.{{.ComponentID}}.librdkafka_options]
"enable.ssl.certificate.verification" = "false"
{{- end}}`
}
