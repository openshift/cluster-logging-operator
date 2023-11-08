package conf

import (
	"github.com/openshift/cluster-logging-operator/internal/collector/vector"
	"github.com/openshift/cluster-logging-operator/internal/generator/framework"
)

func Global(namespace, forwarderName string) []framework.Element {
	dataDir := vector.GetDataPath(namespace, forwarderName)
	if dataDir == vector.DefaultDataPath {
		dataDir = ""
	}
	return []framework.Element{
		GlobalOptions{
			ExpireMetricsSecs: 60,
			DataDir:           dataDir,
		},
	}
}

type GlobalOptions struct {
	ExpireMetricsSecs int
	DataDir           string
}

func (GlobalOptions) Name() string {
	return "globalOptionsTemplate"
}

func (g GlobalOptions) Template() string {
	return `
{{define "` + g.Name() + `" -}}
expire_metrics_secs = {{.ExpireMetricsSecs}}
{{ if .DataDir}}
data_dir = "{{.DataDir}}"
{{end}}

[api]
enabled = true
{{end}}
`
}
