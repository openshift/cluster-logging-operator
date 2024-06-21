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

func New(id string, o obsv1.OutputSpec, inputs []string, secrets vectorhelpers.Secrets, strategy common.ConfigStrategy, op Options) []Element {
	if genhelper.IsDebugOutput(op) {
		return []Element{
			elements.Debug(helpers.MakeID(id, "debug"), vectorhelpers.MakeInputs(inputs...)),
		}
	}
	// TODO: create a pattern to filter by input so all this is not necessary
	var els []Element
	// Creates reroutes for 'container','journal','linux','kube','openshift','ovn'
	rerouteID := vectorhelpers.MakeID(id, "reroute")
	els = append(els, RouteBySource(rerouteID, inputs))
	// Container
	transformContainerID := vectorhelpers.MakeID(id, "otlp_container")
	reduceContainerID := vectorhelpers.MakeID(id, "group_by_container")
	els = append(els, TransformContainer(transformContainerID, []string{rerouteID + ".container"}))
	// Group by cluster_id, namespace_name, pod_name, container_name
	els = append(els, GroupByContainer(reduceContainerID, []string{transformContainerID}))
	// Journal
	transformNodeID := vectorhelpers.MakeID(id, "otlp_node")
	els = append(els, TransformJournal(transformNodeID, []string{rerouteID + ".journal"}))
	// Audit
	transformAuditHostID := vectorhelpers.MakeID(id, "otlp_audit_linux")
	transformAuditKubeID := vectorhelpers.MakeID(id, "otlp_audit_kube")
	transformAuditOpenshiftID := vectorhelpers.MakeID(id, "otlp_audit_openshift")
	transformAuditOvnID := vectorhelpers.MakeID(id, "otlp_audit_ovn")
	reduceSourceID := vectorhelpers.MakeID(id, "group_by_source")
	els = append(els, TransformAuditHost(transformAuditHostID, []string{rerouteID + ".linux"}))
	els = append(els, TransformAuditKube(transformAuditKubeID, []string{rerouteID + ".kube"}))
	els = append(els, TransformAuditOpenshift(transformAuditOpenshiftID, []string{rerouteID + ".openshift"}))
	els = append(els, TransformAuditOvn(transformAuditOvnID, []string{rerouteID + ".ovn"}))
	// Group by cluster_id, hostname, log_type
	els = append(els, GroupBySource(reduceSourceID, []string{
		transformNodeID,
		transformAuditHostID,
		transformAuditKubeID,
		transformAuditOpenshiftID,
		transformAuditOvnID,
	}))
	// Normalize all into resource and scopeLogs objects
	formatID := vectorhelpers.MakeID(id, "final_otlp")
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
