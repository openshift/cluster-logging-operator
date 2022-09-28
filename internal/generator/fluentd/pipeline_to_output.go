package fluentd

import (
	"encoding/json"
	"fmt"
	"github.com/openshift/cluster-logging-operator/internal/generator/fluentd/source"
	"sort"

	logging "github.com/openshift/cluster-logging-operator/apis/logging/v1"
	. "github.com/openshift/cluster-logging-operator/internal/generator"
	. "github.com/openshift/cluster-logging-operator/internal/generator/fluentd/elements"
	"github.com/openshift/cluster-logging-operator/internal/generator/fluentd/helpers"
)

const (
	JSONParseType               = "json"
	MultilineExceptionParseType = "multilineException"

	PipelineLabels = `
{{define "PipelineLabels" -}}
# {{.Desc}}
<filter **>
  @type record_transformer
  <record>
    openshift { "labels": %s }
  </record>
</filter>
{{end}}`

	JsonParseTemplate = `{{define "JsonParse" -}}
# {{.Desc}}
<filter %s>
  @type parser
  key_name message
  reserve_data true
  hash_value_field structured
  {{/* https://issues.redhat.com/browse/LOG-1806 
    A non-JSON log message is forwarded as it would be if JSON parsing was not enabled (e.g. to the app index).
    This warning is just fluentd detecting a non-JSON message, but it will still be forwarded to the non-JSON index. */}}
  emit_invalid_record_to_error false 
  remove_key_name_field true
  <parse>
    @type json
    json_parser oj
  </parse>
</filter>
{{end}}`
)

func PipelineToOutputs(spec *logging.ClusterLogForwarderSpec, op Options) []Element {
	var e []Element = []Element{}
	pipelines := spec.Pipelines
	sort.Slice(pipelines, func(i, j int) bool {
		return pipelines[i].Name < pipelines[j].Name
	})
	for _, p := range pipelines {
		po := FromLabel{
			Desc:    fmt.Sprintf("Copying pipeline %s to outputs", p.Name),
			InLabel: helpers.LabelName(p.Name),
		}
		if p.Labels != nil && len(p.Labels) != 0 {
			// ignoring error, because pre-check stage already checked if Labels can be marshalled
			s, _ := json.Marshal(p.Labels)
			po.SubElements = append(po.SubElements,
				ConfLiteral{
					Desc:         "Add User Defined labels to the output record",
					TemplateName: "PipelineLabels",
					TemplateStr:  fmt.Sprintf(PipelineLabels, string(s)),
				})
		}
		if p.DetectMultilineErrors {
			po.SubElements = append(po.SubElements,
				ConfLiteral{
					TemplateName: "matchMultilineDetectException",
					TemplateStr:  MultilineDetectExceptionTemplate,
				})
		}
		if p.Parse == JSONParseType {
			po.SubElements = append(po.SubElements,
				ConfLiteral{
					Desc:         "Parse the logs into json",
					TemplateName: "JsonParse",
					TemplateStr:  fmt.Sprintf(JsonParseTemplate, source.ApplicationTagsForMultilineEx),
				})
		}
		switch len(p.OutputRefs) {
		case 0:
			// should not happen
		case 1:
			po.SubElements = append(po.SubElements,
				Match{
					MatchTags: "**",
					MatchElement: Relabel{
						OutLabel: helpers.LabelName(p.OutputRefs[0]),
					},
				})
		default:
			po.SubElements = append(po.SubElements,
				Match{
					MatchTags: "**",
					MatchElement: Copy{
						DeepCopy: true,
						Stores:   CopyToLabels(helpers.LabelNames(p.OutputRefs)),
					},
				})
		}
		e = append(e, po)
	}
	return e
}
