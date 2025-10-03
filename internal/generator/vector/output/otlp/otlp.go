package otlp

import (
	"fmt"
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
	"github.com/openshift/cluster-logging-operator/internal/utils"
)

const (
	// OtlpLogSourcesOption Option identifier to restrict the generated code to this list of log sources
	OtlpLogSourcesOption = "otlpLogSourcesOption"
)

type Otlp struct {
	ComponentID string
	Inputs      string
	URI         string
	Compression genhelper.OptionalPair
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
protocol.type = "http"
protocol.method = "post"
protocol.encoding.codec = "json"
protocol.encoding.except_fields = ["_internal"]
protocol.payload_prefix = "{\"resourceLogs\":"
protocol.payload_suffix = "}"
{{.Compression}}
{{end}}
`
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

type logSources []string

func (ls logSources) Has(source string) bool {
	for _, e := range ls {
		if e == source {
			return true
		}
	}
	return false
}

func New(id string, o obs.OutputSpec, inputs []string, secrets observability.Secrets, strategy common.ConfigStrategy, op Options) []Element {
	if genhelper.IsDebugOutput(op) {
		return []Element{
			elements.Debug(helpers.MakeID(id, "debug"), vectorhelpers.MakeInputs(inputs...)),
		}
	}
	var opSources, _ = utils.GetOption(op, OtlpLogSourcesOption, allLogSources)
	if len(opSources) == 0 {
		panic("InputSources not found while generating config")
	}
	sources := logSources(opSources)
	// TODO: create a pattern to filter by input so all this is not necessary
	var els []Element
	// Creates reroutes for 'container','node','auditd','kubeAPI','openshiftAPI','ovn'
	rerouteID := vectorhelpers.MakeID(id, "reroute") // "output_my_id_reroute
	els = append(els, RouteBySource(rerouteID, inputs, sources))
	els = append(els, elements.NewUnmatched(rerouteID, op, nil))

	groupBySourceInputs := []string{}
	groupByHostInputs := []string{}
	reduceInputs := []string{}
	// Container
	if sources.Has(logSourceContainer) {
		transformContainerID := vectorhelpers.MakeID(id, logSourceContainer)                       // "output_my_id_container"
		transformContainerInputID := vectorhelpers.MakeRouteInputID(rerouteID, logSourceContainer) // "output_my_id_reroute.container"
		reduceContainerID := vectorhelpers.MakeID(id, "groupby", "container")
		els = append(els, TransformContainer(transformContainerID, []string{transformContainerInputID}))
		// Group by cluster_id, namespace_name, pod_name, container_name
		els = append(els, GroupByContainer(reduceContainerID, []string{transformContainerID}))

		reduceInputs = append(reduceInputs, reduceContainerID)
	}
	if sources.Has(logSourceNode) {
		// Journal
		transformNodeID := vectorhelpers.MakeID(id, logSourceNode)
		transformNodeRouteID := vectorhelpers.MakeRouteInputID(rerouteID, logSourceNode)
		els = append(els, TransformJournal(transformNodeID, []string{transformNodeRouteID}))

		groupByHostInputs = append(groupByHostInputs, transformNodeID)
	}

	if sources.Has(logSourceAuditd) {
		// Audit
		transformAuditHostID := vectorhelpers.MakeID(id, logSourceAuditd)
		transformAuditHostRouteID := vectorhelpers.MakeRouteInputID(rerouteID, logSourceAuditd)
		els = append(els, TransformAuditHost(transformAuditHostID, []string{transformAuditHostRouteID}))
		groupByHostInputs = append(groupByHostInputs, transformAuditHostID)
	}
	if sources.Has(logSourceKubeAPI) {
		transformAuditKubeID := vectorhelpers.MakeID(id, logSourceKubeAPI)
		transformAuditKubeRouteID := vectorhelpers.MakeRouteInputID(rerouteID, logSourceKubeAPI)
		els = append(els, TransformAuditKube(transformAuditKubeID, []string{transformAuditKubeRouteID}))
		groupBySourceInputs = append(groupBySourceInputs, transformAuditKubeID)
	}
	if sources.Has(logSourceOpenshiftAPI) {

		transformAuditOpenshiftID := vectorhelpers.MakeID(id, logSourceOpenshiftAPI)
		transformAuditOpenshiftRouteID := vectorhelpers.MakeRouteInputID(rerouteID, logSourceOpenshiftAPI)
		els = append(els, TransformAuditOpenshift(transformAuditOpenshiftID, []string{transformAuditOpenshiftRouteID}))
		groupBySourceInputs = append(groupBySourceInputs, transformAuditOpenshiftID)
	}
	if sources.Has(logSourceOvn) {
		transformAuditOvnID := vectorhelpers.MakeID(id, logSourceOvn)
		transformAuditOvnRouteID := vectorhelpers.MakeRouteInputID(rerouteID, logSourceOvn)
		els = append(els, TransformAuditOvn(transformAuditOvnID, []string{transformAuditOvnRouteID}))
		groupBySourceInputs = append(groupBySourceInputs, transformAuditOvnID)
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

	protocolId := id + ".protocol"
	return MergeElements(
		els,
		[]Element{
			sink,
			common.NewAcknowledgments(id, strategy),
			common.NewBatch(protocolId, strategy),
			common.NewBuffer(id, strategy),
			common.NewRequest(protocolId, strategy),
			tls.New(protocolId, o.TLS, secrets, op),
			auth.HTTPAuth(protocolId, o.OTLP.Authentication, secrets, op),
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
	compression := genhelper.NewOptionalPair("protocol.compression", nil)
	if o.OTLP.Tuning != nil && o.OTLP.Tuning.Compression != "" {
		compression = genhelper.NewOptionalPair("protocol.compression", o.OTLP.Tuning.Compression)
	}
	return &Otlp{
		ComponentID: id,
		Inputs:      vectorhelpers.MakeInputs(inputs...),
		URI:         o.OTLP.URL,
		Compression: compression,
	}
}
