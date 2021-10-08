package transform

import "github.com/openshift/cluster-logging-operator/internal/generator/vector/helpers"

type JournalTransform struct {
	SourceID      string
	Desc          string
	SrcType       string
	TranType      string
	TransformName string
	InputPipeline []string
}

func (jt JournalTransform) ComponentID() string {
	jt.SourceID = jt.SourceID + ".infra"
	return jt.SourceID
}

func (jt JournalTransform) SourceType() string {
	return jt.SrcType
}

func (jt JournalTransform) TransformType() string {
	return jt.TranType
}

func (jt JournalTransform) Name() string {
	return "journal_normalizer"
}

func (jt JournalTransform) Template() string {
	return `{{define "` + jt.Name() + `" -}}
[transforms.{{.SourceID}}]
  inputs = ` + helpers.ConcatArrays(jt.InputPipeline) + `
  type = "{{.TransformType}}"
  route.infra = '(starts_with!(.kubernetes.pod_namespace,"kube") && starts_with!(.kubernetes.pod_namespace,"openshift") && .kubernetes.pod_namespace == "default")'
{{end}}`
}
