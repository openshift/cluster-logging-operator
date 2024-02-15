package elasticsearch

import (
	"fmt"
	. "github.com/openshift/cluster-logging-operator/internal/generator/framework"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/normalize"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/output/common"
	"strings"

	logging "github.com/openshift/cluster-logging-operator/api/logging/v1"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	genhelper "github.com/openshift/cluster-logging-operator/internal/generator/helpers"
	. "github.com/openshift/cluster-logging-operator/internal/generator/vector/elements"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/helpers"
	corev1 "k8s.io/api/core/v1"
)

type Elasticsearch struct {
	ID_Key      string
	Desc        string
	ComponentID string
	Inputs      string
	Index       string
	Endpoint    string
	Version     int
	common.RootMixin
}

func (e Elasticsearch) Name() string {
	return "elasticsearchTemplate"
}

func (e Elasticsearch) Template() string {
	return `{{define "` + e.Name() + `" -}}
[sinks.{{.ComponentID}}]
type = "elasticsearch"
inputs = {{.Inputs}}
endpoints = ["{{.Endpoint}}"]
bulk.index = "{{ "{{ write_index }}" }}"
bulk.action = "create"
encoding.except_fields = ["write_index"]
id_key = "_id"
{{.Compression}}
{{- if ne .Version 0 }}
api_version = "v{{ .Version }}"
{{- end }}
{{end}}`
}

func SetESIndex(id string, inputs []string, o logging.OutputSpec, op Options) Element {

	setESIndex := `
index = "default"
if (.log_type == "application"){
  index = "app"
}
if (.log_type == "infrastructure"){
  index = "infra"
}
if (.log_type == "audit"){
  index = "audit"
}
.write_index = index + "-write"`

	addESId := `._id = encode_base64(uuid_v4())`
	removeFile := `del(.file)`
	removeTag := `del(.tag)`
	removeSourceType := `del(.source_type)`

	vrls := []string{setESIndex, addESId, removeFile, removeTag, removeSourceType}

	if o.Elasticsearch != nil {
		es := o.Elasticsearch
		switch {
		case es.StructuredTypeKey != "" && es.StructuredTypeName == "":
			changeIndexName := `
if .log_type == "application" && .structured != null {
  val = .%s
  if val != null {
    .write_index, err = "app-" + val + "-write"
  }
}  
`
			vrls = append(vrls, fmt.Sprintf(changeIndexName, o.Elasticsearch.StructuredTypeKey))
		case es.StructuredTypeKey == "" && es.StructuredTypeName != "":
			changeIndexName := `
if .log_type == "application" && .structured != null {
  .write_index = "app-%s-write"
}
`
			vrls = append(vrls, fmt.Sprintf(changeIndexName, o.Elasticsearch.StructuredTypeName))
		case es.StructuredTypeKey != "" && es.StructuredTypeName != "":
			changeIndexName := `
if .log_type == "application" && .structured != null {
  val = .%s
  if val != null {
    .write_index, err = "app-" + val + "-write"
  } else {
    .write_index = "app-%s-write"
  }
}
`
			vrls = append(vrls, fmt.Sprintf(changeIndexName, o.Elasticsearch.StructuredTypeKey, o.Elasticsearch.StructuredTypeName))
		}
		if es.EnableStructuredContainerLogs {
			vrls = append(vrls, `
  if .log_type == "application"  && .structured != null && .kubernetes.container_name != null && .kubernetes.annotations != null && length!(.kubernetes.annotations) > 0{
	key = join!(["containerType.logging.openshift.io", .kubernetes.container_name], separator: "/")
    index, err = get(value: .kubernetes.annotations, path: [key])
    if index != null && err == null {
      .write_index = join!(["app-",index,"-write"])
    } else {
       log(err, level: "error")
    }
  }
`)
		}
		vrls = append(vrls, `
  if .structured != null && .write_index == "app-write" {
    .message = encode_json(.structured)
    del(.structured)
  }
`)
	}

	return Remap{
		Desc:        "Set Elasticsearch index",
		ComponentID: id,
		Inputs:      helpers.MakeInputs(inputs...),
		VRL:         strings.Join(helpers.TrimSpaces(vrls), "\n"),
	}
}

func FlattenLabels(id string, inputs []string) Element {
	return Remap{
		ComponentID: id,
		Inputs:      helpers.MakeInputs(inputs...),
		VRL: strings.Join([]string{
			normalize.VRLOpenShiftSequence,
			normalize.VRLDedotLabels,
			VRLFlattenLabels,
			VRLPruneLabels,
		}, "\n"),
	}
}

const (
	VRLFlattenLabels = `if exists(.kubernetes.labels) {
  .kubernetes.flat_labels = []
  for_each(object!(.kubernetes.labels)) -> |k,v| {
    .kubernetes.flat_labels = push(.kubernetes.flat_labels, join!([string(k),"=",string!(v)]))
  }
}`
	VRLPruneLabels = `if exists(.kubernetes.labels) {
  exclusions = ["app_kubernetes_io_name", "app_kubernetes_io_instance", "app_kubernetes_io_version", "app_kubernetes_io_component", "app_kubernetes_io_part-of", "app_kubernetes_io_managed-by", "app_kubernetes_io_created-by"]
  keep = {}
  for_each(object!(.kubernetes.labels))->|k,v|{
    if !includes(exclusions, k) {
      .kubernetes.labels, err = remove(object!(.kubernetes.labels),[k],true)
      if err != null {
        log(err, level: "error")
      }
    }
  }
}`
)

func (e *Elasticsearch) SetCompression(algo string) {
	e.Compression.Value = algo
}

func New(id string, o logging.OutputSpec, inputs []string, secret *corev1.Secret, strategy common.ConfigStrategy, op Options) []Element {
	outputs := []Element{}

	esIndexID := helpers.MakeID(id, "add_es_index")
	dedotID := helpers.MakeID(id, "dedot_and_flatten")

	if genhelper.IsDebugOutput(op) {
		return []Element{
			SetESIndex(esIndexID, inputs, o, op),
			FlattenLabels(dedotID, []string{esIndexID}),
			Debug(id, helpers.MakeInputs([]string{dedotID}...)),
		}
	}
	sink := Output(id, o, []string{dedotID}, secret, op)
	if strategy != nil {
		strategy.VisitSink(sink)
	}
	request := common.NewRequest(id, strategy)
	request.TimeoutSecs.Value = 2147483648
	outputs = MergeElements(outputs,
		[]Element{
			SetESIndex(esIndexID, inputs, o, op),
			FlattenLabels(dedotID, []string{esIndexID}),
			Output(id, o, []string{dedotID}, secret, op),
			common.NewAcknowledgments(id, strategy),
			common.NewBatch(id, strategy),
			common.NewBuffer(id, strategy),
			request,
		},

		common.TLS(id, o, secret, op),
		BasicAuth(id, o, secret),
	)

	return outputs
}

func Output(id string, o logging.OutputSpec, inputs []string, secret *corev1.Secret, op Options) *Elasticsearch {
	es := Elasticsearch{
		ComponentID: id,
		Endpoint:    o.URL,
		Inputs:      helpers.MakeInputs(inputs...),
		RootMixin:   common.NewRootMixin(nil),
	}
	// If valid version is specified
	if o.Elasticsearch != nil && o.Elasticsearch.Version > 0 {
		es.Version = o.Elasticsearch.Version
	} else {
		es.Version = logging.DefaultESVersion
	}
	return &es
}

func BasicAuth(id string, o logging.OutputSpec, secret *corev1.Secret) []Element {

	conf := []Element{}

	if o.Secret != nil {
		hasBasicAuth := false
		conf = append(conf, BasicAuthConf{
			Desc:        "Basic Auth Config",
			ComponentID: id,
		})
		if common.HasUsernamePassword(secret) {
			hasBasicAuth = true
			up := UserNamePass{
				Username: common.GetFromSecret(secret, constants.ClientUsername),
				Password: common.GetFromSecret(secret, constants.ClientPassword),
			}
			conf = append(conf, up)
		}
		if !hasBasicAuth {
			return []Element{}
		}
	}

	return conf
}
