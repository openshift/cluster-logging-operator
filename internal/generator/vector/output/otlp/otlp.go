package otlp

import (
	"fmt"
	"github.com/openshift/cluster-logging-operator/internal/utils"
	"github.com/openshift/cluster-logging-operator/internal/utils/sets"
	"sort"
	"strings"

	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	"github.com/openshift/cluster-logging-operator/internal/api/observability"
	. "github.com/openshift/cluster-logging-operator/internal/generator/framework"
	genhelper "github.com/openshift/cluster-logging-operator/internal/generator/helpers"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/elements"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/helpers"
	vectorhelpers "github.com/openshift/cluster-logging-operator/internal/generator/vector/helpers"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/output/common"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/output/common/auth"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/output/common/tls"
)

const (
	// OtlpLogSourcesOption Option identifier to restrict the generated code to this list of log sources
	OtlpLogSourcesOption = "otlpLogSourcesOption"
)

type Otlp struct {
	ComponentID string
	Inputs      string
	URI         string
	common.RootMixin
}

func (p Otlp) Name() string {
	return "vectorOtlpTemplate"
}

func (p Otlp) Template() string {
	return `{{define "` + p.Name() + `" -}}
[sinks.{{.ComponentID}}]
type = "opentelemetry"
inputs = {{.Inputs}}
protocol.uri = "{{.URI}}"
protocol.method = "post"
protocol.encoding = "json"
protocol.payload_prefix = "{\"resourceLogs\":"
protocol.payload_suffix = "}"
{{end}}
`
}

func (p *Otlp) SetCompression(algo string) {
	p.Compression.Value = algo
}

const (
	logSourceContainer    = string(obs.ApplicationSourceContainer)
	logSourceNode         = string(obs.InfrastructureSourceNode)
	logSourceAuditd       = string(obs.AuditSourceAuditd)
	logSourceKubeAPI      = string(obs.AuditSourceKube)
	logSourceOpenshiftAPI = string(obs.AuditSourceOpenShift)
	logSourceOvn          = string(obs.AuditSourceOVN)
)

var (
	allLogSources = []string{logSourceContainer, logSourceNode, logSourceAuditd, logSourceKubeAPI, logSourceOpenshiftAPI, logSourceOvn}
)

func New(id string, o obs.OutputSpec, inputs []string, secrets observability.Secrets, strategy common.ConfigStrategy, op Options) []Element {
	if genhelper.IsDebugOutput(op) {
		return []Element{
			elements.Debug(helpers.MakeID(id, "debug"), vectorhelpers.MakeInputs(inputs...)),
		}
	}
	opLogSources, _ := utils.GetOption(op, OtlpLogSourcesOption, allLogSources)
	logSources := sets.NewString(opLogSources...)

	// TODO: create a pattern to filter by input so all this is not necessary
	var els []Element
	// Creates reroutes for 'container','node','auditd','kubeAPI','openshiftAPI','ovn'
	rerouteID := vectorhelpers.MakeID(id, "reroute") // "output_my_id_reroute
	els = append(els, RouteBySource(rerouteID, inputs, logSources.List()))

	groupBySourceInputs := []string{}
	groupByHostInputs := []string{}
	reduceInputs := []string{}
	// Container
	if logSources.Has(logSourceContainer) {
		transformContainerID := vectorhelpers.MakeID(id, logSourceContainer)                       // "output_my_id_container"
		transformContainerInputID := vectorhelpers.MakeRouteInputID(rerouteID, logSourceContainer) // "output_my_id_reroute.container"
		reduceContainerID := vectorhelpers.MakeID(id, "groupby", "container")
		els = append(els, TransformContainer(transformContainerID, []string{transformContainerInputID}))
		// Group by cluster_id, namespace_name, pod_name, container_name
		els = append(els, GroupByContainer(reduceContainerID, []string{transformContainerID}))

		reduceInputs = append(reduceInputs, reduceContainerID)
	}
	if logSources.Has(logSourceNode) {
		// Journal
		transformNodeID := vectorhelpers.MakeID(id, logSourceNode)
		transformNodeRouteID := vectorhelpers.MakeRouteInputID(rerouteID, logSourceNode)
		els = append(els, TransformJournal(transformNodeID, []string{transformNodeRouteID}))

		groupByHostInputs = append(groupByHostInputs, transformNodeID)
	}

	if logSources.Has(logSourceAuditd) || logSources.Has(logSourceKubeAPI) || logSources.Has(logSourceOpenshiftAPI) || logSources.Has(logSourceOvn) {
		// Audit
		transformAuditHostID := vectorhelpers.MakeID(id, logSourceAuditd)
		transformAuditHostRouteID := vectorhelpers.MakeRouteInputID(rerouteID, logSourceAuditd)
		transformAuditKubeID := vectorhelpers.MakeID(id, logSourceKubeAPI)
		transformAuditKubeRouteID := vectorhelpers.MakeRouteInputID(rerouteID, logSourceKubeAPI)
		transformAuditOpenshiftID := vectorhelpers.MakeID(id, logSourceOpenshiftAPI)
		transformAuditOpenshiftRouteID := vectorhelpers.MakeRouteInputID(rerouteID, logSourceOpenshiftAPI)
		transformAuditOvnID := vectorhelpers.MakeID(id, logSourceOvn)
		transformAuditOvnRouteID := vectorhelpers.MakeRouteInputID(rerouteID, logSourceOvn)

		els = append(els, TransformAuditHost(transformAuditHostID, []string{transformAuditHostRouteID}))
		els = append(els, TransformAuditKube(transformAuditKubeID, []string{transformAuditKubeRouteID}))
		els = append(els, TransformAuditOpenshift(transformAuditOpenshiftID, []string{transformAuditOpenshiftRouteID}))
		els = append(els, TransformAuditOvn(transformAuditOvnID, []string{transformAuditOvnRouteID}))

		groupBySourceInputs = append(groupBySourceInputs,
			transformAuditKubeID,
			transformAuditOpenshiftID,
			transformAuditOvnID)
		groupByHostInputs = append(groupByHostInputs, transformAuditHostID)
	}

	// Group by cluster_id, log_source
	if len(groupBySourceInputs) > 0 {
		reduceSourceID := vectorhelpers.MakeID(id, "groupby", "source")
		els = append(els, GroupBySource(reduceSourceID, groupBySourceInputs))
		reduceInputs = append(reduceInputs, reduceSourceID)
	}
	// Group by cluster_id, hostname
	if len(groupByHostInputs) > 0 {
		reduceHostID := vectorhelpers.MakeID(id, "groupby", "host")
		els = append(els, GroupByHost(reduceHostID, groupByHostInputs))
		reduceInputs = append(reduceInputs, reduceHostID)
	}

	// Normalize all into resource and scopeLogs objects
	formatResourceLogsID := vectorhelpers.MakeID(id, "resource", "logs")
	els = append(els, FormatResourceLog(formatResourceLogsID, reduceInputs))
	// Create sink and wrap in `resourceLogs`
	sink := Output(id, o, []string{formatResourceLogsID}, secrets, op)
	if strategy != nil {
		strategy.VisitSink(sink)
	}
	return MergeElements(
		els,
		[]Element{
			sink,
			common.NewEncoding(id, common.CodecJSON),
			common.NewAcknowledgments(id, strategy),
			common.NewBatch(id, strategy),
			common.NewBuffer(id, strategy),
			common.NewRequest(id, strategy),
			tls.New(id, o.TLS, secrets, op),
			auth.HTTPAuth(id, o.OTLP.Authentication, secrets, op),
		},
	)
}

func RouteBySource(id string, inputs []string, logSources []string) Element {
	// Sort to match the route vrl logic
	sort.Strings(logSources)
	routes := map[string]string{}
	for _, source := range logSources {
		routes[strings.ToLower(source)] = fmt.Sprintf("'.log_source == \"%s\"'", source)
	}

	return elements.Route{
		Desc:        "Route logs separately by log_source",
		ComponentID: id,
		Inputs:      helpers.MakeInputs(inputs...),
		Routes:      routes,
	}
}

func Output(id string, o obs.OutputSpec, inputs []string, secrets observability.Secrets, op Options) *Otlp {
	return &Otlp{
		ComponentID: id,
		Inputs:      vectorhelpers.MakeInputs(inputs...),
		URI:         o.OTLP.URL,
		RootMixin:   common.NewRootMixin(nil),
	}
}
