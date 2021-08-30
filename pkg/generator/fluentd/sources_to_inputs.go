package fluentd

import (
	logging "github.com/openshift/cluster-logging-operator/pkg/apis/logging/v1"
	. "github.com/openshift/cluster-logging-operator/pkg/generator"
	. "github.com/openshift/cluster-logging-operator/pkg/generator/fluentd/elements"
	"github.com/openshift/cluster-logging-operator/pkg/generator/fluentd/helpers"
	"github.com/openshift/cluster-logging-operator/pkg/generator/fluentd/source"
)

func SourcesToInputs(spec *logging.ClusterLogForwarderSpec, o Options) []Element {
	var el []Element = make([]Element, 0)
	types := GatherSources(spec, o)
	types = AddLegacySources(types, o)
	if types.Has(logging.InputNameInfrastructure) {
		el = append(el, Match{
			Desc:      "Include Infrastructure logs",
			MatchTags: source.InfraTags,
			MatchElement: Relabel{
				OutLabel: helpers.SourceTypeLabelName(logging.InputNameInfrastructure),
			},
		})
	} else {
		el = append(el, ConfLiteral{
			Desc:         "Discard Infrastructure logs",
			Pattern:      source.InfraTags,
			TemplateName: "discardMatched",
			TemplateStr:  DiscardMatched,
		})
	}
	if types.Has(logging.InputNameApplication) {
		el = append(el, Match{
			Desc:      "Include Application logs",
			MatchTags: source.ApplicationTags,
			MatchElement: Relabel{
				OutLabel: helpers.SourceTypeLabelName(logging.InputNameApplication),
			},
		})
	} else {
		el = append(el, ConfLiteral{
			Desc:         "Discard Application logs",
			Pattern:      source.ApplicationTags,
			TemplateName: "discardMatched",
			TemplateStr:  DiscardMatched,
		})
	}
	if types.Has(logging.InputNameAudit) {
		el = append(el, Match{
			Desc:      "Include Audit logs",
			MatchTags: source.AuditTags,
			MatchElement: Relabel{
				OutLabel: helpers.SourceTypeLabelName(logging.InputNameAudit),
			},
		})
	} else {
		el = append(el, ConfLiteral{
			Desc:         "Discard Audit logs",
			Pattern:      source.AuditTags,
			TemplateName: "discardMatched",
			TemplateStr:  DiscardMatched,
		})
	}
	el = append(el, ConfLiteral{
		Desc:         "Send any remaining unmatched tags to stdout",
		TemplateName: "toStdout",
		Pattern:      "**",
		TemplateStr:  ToStdOut,
	})
	return el
}

const DiscardMatched string = `
{{define "discardMatched" -}}
# {{.Desc}}
<match {{.Pattern}}>
  @type null
</match>
{{end}}`
