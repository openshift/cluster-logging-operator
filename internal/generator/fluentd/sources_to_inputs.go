package fluentd

import (
	logging "github.com/openshift/cluster-logging-operator/api/logging/v1"
	"github.com/openshift/cluster-logging-operator/internal/generator/fluentd/elements"
	"github.com/openshift/cluster-logging-operator/internal/generator/fluentd/helpers"
	"github.com/openshift/cluster-logging-operator/internal/generator/fluentd/source"
	"github.com/openshift/cluster-logging-operator/internal/generator/framework"
	"github.com/openshift/cluster-logging-operator/internal/generator/utils"
)

func SourcesToInputs(spec *logging.ClusterLogForwarderSpec, o framework.Options) []framework.Element {
	var el []framework.Element = make([]framework.Element, 0)
	types := utils.GatherSources(spec, o)
	if types.Has(logging.InputNameInfrastructure) {
		el = append(el, elements.Match{
			Desc:      "Include Infrastructure logs",
			MatchTags: source.InfraTags,
			MatchElement: elements.Relabel{
				OutLabel: helpers.SourceTypeLabelName(logging.InputNameInfrastructure),
			},
		})
	} else {
		el = append(el, framework.ConfLiteral{
			Desc:         "Discard Infrastructure logs",
			Pattern:      source.InfraTags,
			TemplateName: "discardMatched",
			TemplateStr:  DiscardMatched,
		})
	}
	if types.Has(logging.InputNameApplication) {
		el = append(el, elements.Match{
			Desc:      "Include Application logs",
			MatchTags: source.ApplicationTags,
			MatchElement: elements.Relabel{
				OutLabel: helpers.SourceTypeLabelName(logging.InputNameApplication),
			},
		})
	} else {
		el = append(el, framework.ConfLiteral{
			Desc:         "Discard Application logs",
			Pattern:      source.ApplicationTags,
			TemplateName: "discardMatched",
			TemplateStr:  DiscardMatched,
		})
	}
	if types.Has(logging.InputNameAudit) {
		el = append(el, elements.Match{
			Desc:      "Include Audit logs",
			MatchTags: source.AuditTags,
			MatchElement: elements.Relabel{
				OutLabel: helpers.SourceTypeLabelName(logging.InputNameAudit),
			},
		})
	} else {
		el = append(el, framework.ConfLiteral{
			Desc:         "Discard Audit logs",
			Pattern:      source.AuditTags,
			TemplateName: "discardMatched",
			TemplateStr:  DiscardMatched,
		})
	}
	el = append(el, framework.ConfLiteral{
		Desc:         "Send any remaining unmatched tags to stdout",
		TemplateName: "toStdout",
		Pattern:      "**",
		TemplateStr:  elements.ToStdOut,
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
