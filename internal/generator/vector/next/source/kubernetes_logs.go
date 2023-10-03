package source

import (
	"github.com/openshift/cluster-logging-operator/internal/generator"
)

type KubernetesLogs struct {
	generator.ComponentID
	Desc         string
	ExcludePaths string
}

func (kl KubernetesLogs) Name() string {
	return "k8s_logs_template"
}

func (kl KubernetesLogs) Template() string {
	return `{{define "` + kl.Name() + `" -}}
# {{.Desc}}
[sources.{{.ComponentID}}]
type = "kubernetes_logs"
max_read_bytes = 3145728
glob_minimum_cooldown_ms = 15000
auto_partial_merge = true
exclude_paths_glob_patterns = {{.ExcludePaths}}
{{end}}`
}
