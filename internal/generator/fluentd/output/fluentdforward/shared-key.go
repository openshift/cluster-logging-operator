package fluentdforward

import (
	"github.com/openshift/cluster-logging-operator/internal/generator/fluentd/output/security"
)

type SharedKey security.SharedKey

func (sk SharedKey) Name() string {
	return "fluentdforwardSharedKeyTemplate"
}

func (sk SharedKey) Template() string {
	return `{{define "` + sk.Name() + `" -}}
<security>
  self_hostname "#{ENV['NODE_NAME']}"
  shared_key "{{.Key}}"
</security>
{{- end}}`
}
