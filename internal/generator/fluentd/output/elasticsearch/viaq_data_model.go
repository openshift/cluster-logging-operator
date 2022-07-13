package elasticsearch

import (
	"fmt"
	logging "github.com/openshift/cluster-logging-operator/apis/logging/v1"
	. "github.com/openshift/cluster-logging-operator/internal/generator"
	. "github.com/openshift/cluster-logging-operator/internal/generator/fluentd/elements"
	corev1 "k8s.io/api/core/v1"
)

type Viaq struct {
	Elasticsearch *logging.Elasticsearch
}

const (
	AnnotationPrefix = "containerType.logging.openshift.io"
)

func ViaqDataModel(bufspec *logging.FluentdBufferSpec, secret *corev1.Secret, o logging.OutputSpec, op Options) []Element {
	elements := []Element{
		Viaq{
			Elasticsearch: o.Elasticsearch,
		},
	}
	if o.Elasticsearch == nil || (o.Elasticsearch.StructuredTypeKey == "" && o.Elasticsearch.StructuredTypeName == "" && !o.Elasticsearch.EnableStructuredContainerLogs) {
		recordModifier := RecordModifier{
			RemoveKeys: []string{KeyStructured},
		}
		if op[CharEncoding] != nil {
			recordModifier.CharEncoding = fmt.Sprintf("%v", op[CharEncoding])
		}
		elements = append(elements, Filter{
			Desc:      "remove structured field if present",
			MatchTags: "**",
			Element:   recordModifier,
		})
	}
	return elements
}

func (im Viaq) StructuredTypeKey() string {
	if im.Elasticsearch != nil && im.Elasticsearch.StructuredTypeKey != "" {
		return im.Elasticsearch.StructuredTypeKey
	}
	return ""
}
func (im Viaq) StructuredTypeName() string {
	if im.Elasticsearch != nil && im.Elasticsearch.StructuredTypeName != "" {
		return im.Elasticsearch.StructuredTypeName
	}
	return ""
}
func (im Viaq) StructuredTypeAnnotationPrefix() string {
	if im.Elasticsearch != nil && im.Elasticsearch.EnableStructuredContainerLogs {
		return AnnotationPrefix
	}
	return ""
}

func (im Viaq) Name() string {
	return "viaqDataIndexModel"
}

func (im Viaq) Template() string {
	return `{{define "viaqDataIndexModel" -}}
# Viaq Data Model
<filter **>
  @type viaq_data_model
  enable_openshift_model false
  enable_prune_empty_fields false
  rename_time false
  undefined_dot_replace_char UNUSED
  elasticsearch_index_prefix_field 'viaq_index_name'
  <elasticsearch_index_name>
    enabled 'true'
    tag "kubernetes.var.log.pods.openshift_** kubernetes.var.log.pods.openshift-*_** kubernetes.var.log.pods.default_** kubernetes.var.log.pods.kube-*_** journal.system** system.var.log**"
    name_type static
    static_index_name infra-write
{{if (ne .StructuredTypeKey "") -}}
    structured_type_key {{ .StructuredTypeKey }}
{{ end -}}
{{if (ne .StructuredTypeName "") -}}
    structured_type_name {{ .StructuredTypeName }}
{{ end -}}
{{if (ne .StructuredTypeAnnotationPrefix "") -}}
    structured_type_annotation_prefix {{ .StructuredTypeAnnotationPrefix }}
{{ end -}}
  </elasticsearch_index_name>
  <elasticsearch_index_name>
    enabled 'true'
    tag "linux-audit.log** k8s-audit.log** openshift-audit.log** ovn-audit.log**"
    name_type static
    static_index_name audit-write
  </elasticsearch_index_name>
  <elasticsearch_index_name>
    enabled 'true'
    tag "**"
    name_type structured
    static_index_name app-write
{{if (ne .StructuredTypeKey "") -}}
    structured_type_key {{ .StructuredTypeKey }}
{{ end -}}
{{if (ne .StructuredTypeName "") -}}
    structured_type_name {{ .StructuredTypeName }}
{{ end -}}
{{if (ne .StructuredTypeAnnotationPrefix "") -}}
    structured_type_annotation_prefix {{ .StructuredTypeAnnotationPrefix }}
{{ end -}}
  </elasticsearch_index_name>
</filter>
<filter **>
  @type viaq_data_model
  enable_prune_labels true
  enable_openshift_model false
  rename_time false
  undefined_dot_replace_char UNUSED
  prune_labels_exclusions app.kubernetes.io/name,app.kubernetes.io/instance,app.kubernetes.io/version,app.kubernetes.io/component,app.kubernetes.io/part-of,app.kubernetes.io/managed-by,app.kubernetes.io/created-by
</filter>
{{end}}
`
}
