package fluentd

import (
	"encoding/json"
	"fmt"
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
<filter **>
  @type parser
  key_name message
  reserve_data yes
  hash_value_field structured
  <parse>
    @type json
    json_parser oj
  </parse>
</filter>
{{end}}`

	LogLevelSetterTemplate = `{{define "LogLevelSetter" -}}
# {{.Desc}}
<filter **>
  @type record_modifier
  remove_keys _dummy_
  <record>
    _dummy_ ${struct_level = record.dig('structured','level'); level = record['level']; if ![nil, level].include?(struct_level) && [nil, 'default'].include?(level); record['level']=struct_level; end; nil}
  </record>
</filter>
{{end}}
`
)

func PipelineToOutputs(spec *logging.ClusterLogForwarderSpec, op Options) []Element {
	var e []Element
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
					TemplateStr:  JsonParseTemplate,
				},
				ConfLiteral{
					Desc:         "Extract log level from structured if exist",
					TemplateName: "LogLevelSetter",
					TemplateStr:  LogLevelSetterTemplate,
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
