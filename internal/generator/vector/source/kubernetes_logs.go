package source

type KubernetesLogs struct {
	SourceID     string
	Desc         string
	SourceType   string
	ExcludePaths string
}

func (kl KubernetesLogs) ComponentID() string {
	return kl.SourceID
}

func (kl KubernetesLogs) Type() string {
	return kl.SourceType
}

func (kl KubernetesLogs) Name() string {
	return "kubernetes_logs"
}

func (kl KubernetesLogs) Template() string {
	return `{{define "` + kl.Name() + `" -}}
# {{.Desc}}
[sources.{{.SourceType}}]
  type = "{{.SourceType}}"
  auto_partial_merge = true
  exclude_paths_glob_patterns = {{.ExcludePaths}}
{{end}}`
}
