package normalize

type DetectExceptions struct {
	ComponentID string
	Inputs      string
}

func (d DetectExceptions) Name() string {
	return "detectExceptions"
}

func (d DetectExceptions) Template() string {
	return `{{define "detectExceptions" -}}
[transforms.{{.ComponentID}}]
type = "detect_exceptions"
inputs = {{.Inputs}}
languages = ["All"]
group_by = ["kubernetes.namespace_name","kubernetes.pod_name","kubernetes.container_name", "kubernetes.pod_id"]
expire_after_ms = 2000
multiline_flush_interval_ms = 1000
{{end}}`
}
