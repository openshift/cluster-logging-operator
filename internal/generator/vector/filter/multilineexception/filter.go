package multilineexception

import (
	"github.com/openshift/cluster-logging-operator/internal/generator/framework"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/helpers"
)

func NewDetectException(id string, inputs ...string) framework.Element {
	return DetectExceptions{
		ComponentID: id,
		Inputs:      helpers.MakeInputs(inputs...),
	}
}

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
group_by = ["._internal.kubernetes.namespace_name","._internal.kubernetes.pod_name","._internal.kubernetes.container_name", "._internal.kubernetes.pod_id"]
expire_after_ms = 2000
multiline_flush_interval_ms = 1000
message_key = "._internal.message"
{{end}}`
}
