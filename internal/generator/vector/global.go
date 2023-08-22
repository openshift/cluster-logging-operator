package vector

import (
	"github.com/openshift/cluster-logging-operator/internal/generator"
)

func Global() []generator.Element {
	return []generator.Element{
		GlobalOptions{
			ExpireMetricsSecs: 60,
		},
	}
}

type GlobalOptions struct {
	ExpireMetricsSecs int
}

func (GlobalOptions) Name() string {
	return "globalOptionsTemplate"
}

func (g GlobalOptions) Template() string {
	return `
{{define "` + g.Name() + `" -}}
expire_metrics_secs = {{.ExpireMetricsSecs}}
{{end}}
`
}
