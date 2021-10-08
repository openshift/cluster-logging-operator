package transform

import "github.com/openshift/cluster-logging-operator/internal/generator/vector/helpers"

type KubernetesTransform struct {
	SourceID      string
	Desc          string
	SrcType       string
	TranType      string
	TransformName string
	InputPipeline []string
}

func (kt KubernetesTransform) ComponentID() string {
	kt.SourceID = kt.SourceID + ".app"
	return kt.SourceID
}

func (kt KubernetesTransform) SourceType() string {
	return kt.SrcType
}

func (kt KubernetesTransform) TransformType() string {
	return kt.TranType
}

func (kt KubernetesTransform) Name() string {
	return "kube_normalizer"
}

func (kt KubernetesTransform) Template() string {
	return `{{define "` + kt.Name() + `" -}}
[transforms.{{.SourceID}}]
  inputs = ` + helpers.ConcatArrays(kt.InputPipeline) + `
  type = "{{.TransformType}}"
  route.app = '!(starts_with!(.kubernetes.pod_namespace,"kube") && starts_with!(.kubernetes.pod_namespace,"openshift") && .kubernetes.pod_namespace == "default")'
{{end}}`
}
