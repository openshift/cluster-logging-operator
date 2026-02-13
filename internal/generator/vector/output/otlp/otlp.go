package otlp

import (
	"fmt"
	"sort"
	"strings"

	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	"github.com/openshift/cluster-logging-operator/internal/api/observability"
	"github.com/openshift/cluster-logging-operator/internal/generator/framework"
	"github.com/openshift/cluster-logging-operator/internal/generator/adapters"
	genhelper "github.com/openshift/cluster-logging-operator/internal/generator/helpers"
	"github.com/openshift/cluster-logging-operator/internal/generator/url"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/api"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/api/sinks"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/elements"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/helpers"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/output/common"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/output/common/auth"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/output/common/tls"
	"github.com/openshift/cluster-logging-operator/internal/utils"
)

const (
	// OtlpLogSourcesOption Option identifier to restrict the generated code to this list of log sources
	OtlpLogSourcesOption = "otlpLogSourcesOption"
	// MigratedFromLokistackOption Option identifier to skip trace context extraction remap for outputs migrated from lokistack
	MigratedFromLokistackOption = "migratedFromLokistackOption"
)

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

func New(id string, o *adapters.Output, inputs []string, secrets observability.Secrets, op utils.Options) []framework.Element {
	if genhelper.IsDebugOutput(op) {
		return []framework.Element{
			elements.Debug(helpers.MakeID(id, "debug"), helpers.MakeInputs(inputs...)),
		}
	}
	var opSources, _ = utils.GetOption(op, OtlpLogSourcesOption, allLogSources)
	if len(opSources) == 0 {
		panic("InputSources not found while generating config")
	}
	sources := logSources(opSources)
	// TODO: create a pattern to filter by input so all this is not necessary
	var els []framework.Element
	rerouteID := helpers.MakeID(id, "reroute") // "output_my_id_reroute

	// Skip trace context extraction remap for outputs migrated from lokistack.
	// This is because the trace context extraction remap is already applied in the lokistack output.
	if fromLokistack, _ := utils.GetOption(op, MigratedFromLokistackOption, false); fromLokistack {
		els = append(els, RouteBySource(rerouteID, inputs, sources))
	} else {
		// Push all logs bound for OTLP through trace context extraction remap first if not from lokistack
		transformTraceContextID := helpers.MakeID(id, "trace", "context")
		els = append(els, TransformTraceContext(transformTraceContextID, inputs))
		els = append(els, RouteBySource(rerouteID, []string{transformTraceContextID}, sources))
	}
	// Creates reroutes for 'container','node','auditd','kubeAPI','openshiftAPI','ovn'
	els = append(els, elements.NewUnmatched(rerouteID, op, nil))

	groupBySourceInputs := []string{}
	groupByHostInputs := []string{}
	reduceInputs := []string{}
	// Container
	if sources.Has(logSourceContainer) {
		transformContainerID := helpers.MakeID(id, logSourceContainer)                       // "output_my_id_container"
		transformContainerInputID := helpers.MakeRouteInputID(rerouteID, logSourceContainer) // "output_my_id_reroute.container"
		reduceContainerID := helpers.MakeID(id, "groupby", "container")
		els = append(els, TransformContainer(transformContainerID, []string{transformContainerInputID}))
		// Group by cluster_id, namespace_name, pod_name, container_name
		els = append(els, GroupByContainer(reduceContainerID, []string{transformContainerID}))

		reduceInputs = append(reduceInputs, reduceContainerID)
	}
	if sources.Has(logSourceNode) {
		// Journal
		transformNodeID := helpers.MakeID(id, logSourceNode)
		transformNodeRouteID := helpers.MakeRouteInputID(rerouteID, logSourceNode)
		els = append(els, TransformJournal(transformNodeID, []string{transformNodeRouteID}))

		groupByHostInputs = append(groupByHostInputs, transformNodeID)
	}

	if sources.Has(logSourceAuditd) {
		// Audit
		transformAuditHostID := helpers.MakeID(id, logSourceAuditd)
		transformAuditHostRouteID := helpers.MakeRouteInputID(rerouteID, logSourceAuditd)
		els = append(els, TransformAuditHost(transformAuditHostID, []string{transformAuditHostRouteID}))
		groupByHostInputs = append(groupByHostInputs, transformAuditHostID)
	}
	if sources.Has(logSourceKubeAPI) {
		transformAuditKubeID := helpers.MakeID(id, logSourceKubeAPI)
		transformAuditKubeRouteID := helpers.MakeRouteInputID(rerouteID, logSourceKubeAPI)
		els = append(els, TransformAuditKube(transformAuditKubeID, []string{transformAuditKubeRouteID}))
		groupBySourceInputs = append(groupBySourceInputs, transformAuditKubeID)
	}
	if sources.Has(logSourceOpenshiftAPI) {

		transformAuditOpenshiftID := helpers.MakeID(id, logSourceOpenshiftAPI)
		transformAuditOpenshiftRouteID := helpers.MakeRouteInputID(rerouteID, logSourceOpenshiftAPI)
		els = append(els, TransformAuditOpenshift(transformAuditOpenshiftID, []string{transformAuditOpenshiftRouteID}))
		groupBySourceInputs = append(groupBySourceInputs, transformAuditOpenshiftID)
	}
	if sources.Has(logSourceOvn) {
		transformAuditOvnID := helpers.MakeID(id, logSourceOvn)
		transformAuditOvnRouteID := helpers.MakeRouteInputID(rerouteID, logSourceOvn)
		els = append(els, TransformAuditOvn(transformAuditOvnID, []string{transformAuditOvnRouteID}))
		groupBySourceInputs = append(groupBySourceInputs, transformAuditOvnID)
	}

	// Group by cluster_id, log_source
	if len(groupBySourceInputs) > 0 {
		reduceSourceID := helpers.MakeID(id, "groupby", "source")
		els = append(els, GroupBySource(reduceSourceID, groupBySourceInputs))
		reduceInputs = append(reduceInputs, reduceSourceID)
	}
	// Group by cluster_id, hostname
	if len(groupByHostInputs) > 0 {
		reduceHostID := helpers.MakeID(id, "groupby", "host")
		els = append(els, GroupByHost(reduceHostID, groupByHostInputs))
		reduceInputs = append(reduceInputs, reduceHostID)
	}

	// Normalize all into resource and scopeLogs objects
	formatResourceLogsID := helpers.MakeID(id, "resource", "logs")
	els = append(els, FormatResourceLog(formatResourceLogsID, reduceInputs))

	return MergeElements(
		els,
		[]Element{
			api.NewConfig(func(c *api.Config) {
				c.Sinks[id] = sinks.NewOpenTelemetry(o.OTLP.URL, func(s *sinks.OpenTelemetry) {
					s.Protocol.Type = "http"
					s.Protocol.Method = sinks.MethodTypePost
					s.Protocol.Encoding = common.NewApiEncoding(api.CodecTypeJSON)
					s.Protocol.PayloadPrefix = "{\"resourceLogs\":"
					s.Protocol.PayloadSuffix = "}"
					if o.OTLP.Tuning != nil {
						s.Protocol.Compression = sinks.CompressionType(o.OTLP.Tuning.Compression)
						s.Batch = common.NewApiBatch(o)
						s.Buffer = common.NewApiBuffer(o)
						s.Protocol.Request = common.NewApiRequest(o)
					}
					if o.TLS != nil && url.IsSecure(o.OTLP.URL) {
						s.Protocol.TLS = tls.NewTls(o, secrets, op)
					}
					s.Protocol.Auth = auth.NewHttpAuth(o.OTLP.Authentication, op)

				}, formatResourceLogsID)
			}),
		},
	)
}

func RouteBySource(id string, inputs []string, logSources []string) framework.Element {
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
