package otlp

import (
	"fmt"
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	. "github.com/openshift/cluster-logging-operator/internal/generator/framework"
	genhelper "github.com/openshift/cluster-logging-operator/internal/generator/helpers"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/elements"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/helpers"
	vectorhelpers "github.com/openshift/cluster-logging-operator/internal/generator/vector/helpers"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/output/common"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/output/common/auth"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/output/common/tls"
	"sort"
	"strings"
)

type Otlp struct {
	ComponentID      string
	Inputs           string
	URI              string
	common.RootMixin //TODO: remove??
}

func (p Otlp) Name() string {
	return "vectorOtlpTemplate"
}

func (p Otlp) Template() string {
	return `{{define "` + p.Name() + `" -}}
[sinks.{{.ComponentID}}]
type = "http"
inputs = {{.Inputs}}
uri = "{{.URI}}"
method = "post"
payload_prefix = "{\"resourceLogs\":"
payload_suffix = "}"
{{.Compression}}
{{end}}
`
}

// TODO: test this for otlp
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

	// For now we are grouping all container logs by their ns, pod and container names
	// so there is no distinction here at the moment.
	// TODO: revise or remove
	logSourceContainerApp   = string(obs.ApplicationSourceContainer)
	logSourceContainerInfra = string(obs.InfrastructureSourceContainer)
)

func New(id string, o obs.OutputSpec, inputs []string, secrets vectorhelpers.Secrets, strategy common.ConfigStrategy, op Options) []Element {
	if genhelper.IsDebugOutput(op) {
		return []Element{
			elements.Debug(helpers.MakeID(id, "debug"), vectorhelpers.MakeInputs(inputs...)),
		}
	}
	// TODO: create a pattern to filter by input so all this is not necessary
	var els []Element
	// Creates reroutes for 'container','node','auditd','kubeAPI','openshiftAPI','ovn'
	rerouteID := vectorhelpers.MakeID(id, "reroute") // "output_my_id_reroute
	els = append(els, RouteBySource(rerouteID, inputs))
	// Container
	transformContainerID := vectorhelpers.MakeID(id, logSourceContainer)                       // "output_my_id_container"
	transformContainerInputID := vectorhelpers.MakeRouteInputID(rerouteID, logSourceContainer) // "output_my_id_reroute.container"
	reduceContainerID := vectorhelpers.MakeID(id, "groupby", "container")
	els = append(els, TransformContainer(transformContainerID, []string{transformContainerInputID}))
	// Group by cluster_id, namespace_name, pod_name, container_name
	els = append(els, GroupByContainer(reduceContainerID, []string{transformContainerID}))
	// Journal
	transformNodeID := vectorhelpers.MakeID(id, logSourceNode)
	transformNodeRouteID := vectorhelpers.MakeRouteInputID(rerouteID, logSourceNode)
	els = append(els, TransformJournal(transformNodeID, []string{transformNodeRouteID}))
	// Audit
	transformAuditHostID := vectorhelpers.MakeID(id, logSourceAuditd)
	transformAuditHostRouteID := vectorhelpers.MakeRouteInputID(rerouteID, logSourceAuditd)
	transformAuditKubeID := vectorhelpers.MakeID(id, logSourceKubeAPI)
	transformAuditKubeRouteID := vectorhelpers.MakeRouteInputID(rerouteID, logSourceKubeAPI)
	transformAuditOpenshiftID := vectorhelpers.MakeID(id, logSourceOpenshiftAPI)
	transformAuditOpenshiftRouteID := vectorhelpers.MakeRouteInputID(rerouteID, logSourceOpenshiftAPI)
	transformAuditOvnID := vectorhelpers.MakeID(id, logSourceOvn)
	transformAuditOvnRouteID := vectorhelpers.MakeRouteInputID(rerouteID, logSourceOvn)
	reduceSourceID := vectorhelpers.MakeID(id, "groupby", "source")
	reduceHostID := vectorhelpers.MakeID(id, "groupby", "host")
	els = append(els, TransformAuditHost(transformAuditHostID, []string{transformAuditHostRouteID}))
	els = append(els, TransformAuditKube(transformAuditKubeID, []string{transformAuditKubeRouteID}))
	els = append(els, TransformAuditOpenshift(transformAuditOpenshiftID, []string{transformAuditOpenshiftRouteID}))
	els = append(els, TransformAuditOvn(transformAuditOvnID, []string{transformAuditOvnRouteID}))
	// Group by cluster_id, log_source
	els = append(els, GroupBySource(reduceSourceID, []string{
		transformNodeID,
		transformAuditKubeID,
		transformAuditOpenshiftID,
		transformAuditOvnID,
	}))
	// Group by cluster_id, hostname
	els = append(els, GroupByHost(reduceHostID, []string{
		transformAuditHostID,
	}))

	// Normalize all into resource and scopeLogs objects
	formatResourceLogsID := vectorhelpers.MakeID(id, "resource", "logs")
	els = append(els, FormatResourceLog(formatResourceLogsID, []string{
		reduceContainerID,
		reduceSourceID,
		reduceHostID,
	}))
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
			tls.New(id, o.TLS, secrets, op),
			auth.HTTPAuth(id, o.OTLP.Authentication, secrets),
		},
	)
}

func RouteBySource(id string, inputs []string) Element {
	// TODO: refactor based on existing map of logSourceTypes?
	logSources := []string{
		logSourceContainer,
		logSourceNode,
		logSourceAuditd,
		logSourceKubeAPI,
		logSourceOpenshiftAPI,
		logSourceOvn,
	}
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

func Output(id string, o obs.OutputSpec, inputs []string, secrets vectorhelpers.Secrets, op Options) *Otlp {
	return &Otlp{
		ComponentID: id,
		Inputs:      vectorhelpers.MakeInputs(inputs...),
		URI:         o.OTLP.URL,
		RootMixin:   common.NewRootMixin(nil),
	}
}
