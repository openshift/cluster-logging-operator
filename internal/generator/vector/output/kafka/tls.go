package kafka

import (
	"fmt"

	"github.com/openshift/cluster-logging-operator/internal/generator/vector/output/security"
)

type TLS security.TLS

func (t TLS) Name() string {
	return "kafkaTLSTemplate"
}

func (t TLS) Template() string {
	return fmt.Sprintf(`{{define "kafkaTLSTemplate" -}}
enabled = %t
{{end}}`, t)
}
