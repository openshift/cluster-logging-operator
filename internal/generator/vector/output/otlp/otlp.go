package otlp

import (
	obsv1 "github.com/openshift/cluster-logging-operator/api/observability/v1"
	. "github.com/openshift/cluster-logging-operator/internal/generator/framework"
	genhelper "github.com/openshift/cluster-logging-operator/internal/generator/helpers"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/elements"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/helpers"
	vectorhelpers "github.com/openshift/cluster-logging-operator/internal/generator/vector/helpers"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/output/common"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/output/common/auth"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/output/common/tls"
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
encoding.codec = "json"
{{.Compression}}
{{end}}
`
}

// TODO: test this for otlp
func (p *Otlp) SetCompression(algo string) {
	p.Compression.Value = algo
}

const (
	logSourceContainer    = "container"
	logSourceNode         = "node"
	logSourceAuditd       = "auditd"
	logSourceKubeAPI      = "kubeapi"
	logSourceOpenshiftAPI = "openshiftapi"
	logSourceOvn          = "ovn"
)

func New(id string, o obsv1.OutputSpec, inputs []string, secrets vectorhelpers.Secrets, strategy common.ConfigStrategy, op Options) []Element {
	if genhelper.IsDebugOutput(op) {
		return []Element{
			elements.Debug(helpers.MakeID(id, "debug"), vectorhelpers.MakeInputs(inputs...)),
		}
	}
	// TODO: create a pattern to filter by input so all this is not necessary
	var els []Element
	// Creates reroutes for 'container','node','auditd','kubeAPI','openshiftAPI','ovn'
	rerouteID := vectorhelpers.MakeID(id, "reroute")
	els = append(els, RouteBySource(rerouteID, inputs))
	// Container
	transformContainerID := vectorhelpers.MakeID(id, logSourceContainer)
	reduceContainerID := vectorhelpers.MakeID(id, "groupby", "container")
	els = append(els, TransformContainer(transformContainerID, []string{rerouteID + "." + logSourceContainer}))
	// Group by cluster_id, namespace_name, pod_name, container_name
	els = append(els, GroupByContainer(reduceContainerID, []string{transformContainerID}))
	// Journal
	transformNodeID := vectorhelpers.MakeID(id, logSourceNode)
	els = append(els, TransformJournal(transformNodeID, []string{rerouteID + "." + logSourceNode}))
	// Audit
	transformAuditHostID := vectorhelpers.MakeID(id, logSourceAuditd)
	transformAuditKubeID := vectorhelpers.MakeID(id, logSourceKubeAPI)
	transformAuditOpenshiftID := vectorhelpers.MakeID(id, logSourceOpenshiftAPI)
	transformAuditOvnID := vectorhelpers.MakeID(id, logSourceOvn)
	reduceSourceID := vectorhelpers.MakeID(id, "groupby", "source")
	els = append(els, TransformAuditHost(transformAuditHostID, []string{rerouteID + "." + logSourceAuditd}))
	els = append(els, TransformAuditKube(transformAuditKubeID, []string{rerouteID + "." + logSourceKubeAPI}))
	els = append(els, TransformAuditOpenshift(transformAuditOpenshiftID, []string{rerouteID + "." + logSourceOpenshiftAPI}))
	els = append(els, TransformAuditOvn(transformAuditOvnID, []string{rerouteID + "." + logSourceOvn}))
	// Group by cluster_id, hostname, log_type
	els = append(els, GroupBySource(reduceSourceID, []string{
		transformNodeID,
		transformAuditHostID,
		transformAuditKubeID,
		transformAuditOpenshiftID,
		transformAuditOvnID,
	}))
	// Normalize all into resource and scopeLogs objects
	formatID := vectorhelpers.MakeID(id, "resource", "logs")
	els = append(els, FormatResourceLog(formatID, []string{
		reduceContainerID,
		reduceSourceID,
		rerouteID + "._unmatched", // mostly for debug, but could be necessary?
	}))
	// Create sink and wrap in `resourceLogs`
	sink := Output(id, o, []string{formatID}, secrets, op)
	if strategy != nil {
		strategy.VisitSink(sink)
	}
	return MergeElements(
		els,
		[]Element{
			sink,
			common.NewAcknowledgments(id, strategy),
			tls.New(id, o.TLS, secrets, op),
			auth.HTTPAuth(id, o.OTLP.Authentication, secrets),
		},
	)
}

func Output(id string, o obsv1.OutputSpec, inputs []string, secrets vectorhelpers.Secrets, op Options) *Otlp {
	return &Otlp{
		ComponentID: id,
		Inputs:      vectorhelpers.MakeInputs(inputs...),
		URI:         o.OTLP.URL,
		RootMixin:   common.NewRootMixin(nil),
	}
}
