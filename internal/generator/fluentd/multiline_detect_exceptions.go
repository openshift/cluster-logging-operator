package fluentd

import (
	logging "github.com/openshift/cluster-logging-operator/apis/logging/v1"
	. "github.com/openshift/cluster-logging-operator/internal/generator"
	. "github.com/openshift/cluster-logging-operator/internal/generator/fluentd/elements"
	"github.com/openshift/cluster-logging-operator/internal/generator/fluentd/helpers"
)

func MultilineDetectExceptions(spec *logging.ClusterLogForwarderSpec, o Options) []Element {
	return []Element{
		Pipeline{
			InLabel: helpers.LabelName("_MULITLINE_DETECT"),
			SubElements: MergeElements([]Element{
				ConfLiteral{
					TemplateName: "matchMultilineDetectException",
					TemplateStr: `
{{define "matchMultilineDetectException" -}}
<match kubernetes.**>
  @id multiline-detect-except
  @type detect_exceptions
  remove_tag_prefix 'kubernetes'
  message log
  force_line_breaks true
  multiline_flush_interval .2
</match>
<match **>
  @type relabel
  @label @INGRESS
</match>
{{end}}
`,
				},
			}),
		},
	}
}
