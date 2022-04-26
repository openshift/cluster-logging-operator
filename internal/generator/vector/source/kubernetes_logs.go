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
auto_partial_merge = true
exclude_paths_glob_patterns = {{.ExcludePaths}}
pod_annotation_fields.pod_labels = "kubernetes.labels"
pod_annotation_fields.pod_namespace = "kubernetes.namespace_name"
{{end}}`
}
