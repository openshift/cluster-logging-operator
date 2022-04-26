package fluentd

import (
	"github.com/openshift/cluster-logging-operator/internal/generator"
	"github.com/openshift/cluster-logging-operator/internal/generator/fluentd/elements"
	"github.com/openshift/cluster-logging-operator/internal/generator/fluentd/helpers"
)

const OnError string = `
{{define "OnError" -}}
<filter **>
  @type record_transformer
  enable_ruby
  <record>
    message ${record['message'].encode!('UTF-8', :undef => :replace, :invalid => :replace, :replace => "?")}
  </record>
</filter>
{{end}}
`

func RetryError() []generator.Element {
	return []generator.Element{
		elements.Pipeline{
			InLabel: helpers.LabelName("ERROR"),
			Desc:    "Try to resend message in case error",
			SubElements: []generator.Element{
				generator.ConfLiteral{
					TemplateName: "OnError",
					TemplateStr:  OnError,
				},
				elements.Match{
					Desc:      "",
					MatchTags: "**",
					MatchElement: elements.Relabel{
						OutLabel: helpers.LabelName("DEFAULT"),
					},
				},
			},
		},
	}
}
