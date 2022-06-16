package vector

import (
	"fmt"
	"sort"
	"strings"

	logging "github.com/openshift/cluster-logging-operator/apis/logging/v1"
	. "github.com/openshift/cluster-logging-operator/internal/generator"
	. "github.com/openshift/cluster-logging-operator/internal/generator/vector/elements"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/helpers"
)

const (
	NsKube      = "kube"
	NsOpenshift = "openshift"
	NsDefault   = "default"

	K8sNamespaceName = ".kubernetes.namespace_name"
	K8sLabelKeyExpr  = ".kubernetes.labels.%s"

	InputContainerLogs   = "container_logs"
	InputJournalLogs     = "journal_logs"
	RouteApplicationLogs = "route_application_logs"

	SrcPassThrough = "."

	HostAuditLogTag = ".linux-audit.log"
	HostAuditLogID  = "tagged_host_audit_logs"

	K8sAuditLogTag = ".k8s-audit.log"
	K8sAuditLogID  = "tagged_k8s_audit_logs"

	OpenAuditLogTag = ".openshift-audit.log"
	OpenAuditLogID  = "tagged_openshift_audit_logs"

	OvnAuditLogTag = ".ovn-audit.log"
	OvnAuditLogID  = "tagged_ovn_audit_logs"

	ParseAndFlatten = `. = merge(., parse_json!(string!(.message))) ?? .
del(.message)
`
)

var (
	InfraContainerLogs = OR(
		StartWith(K8sNamespaceName, NsKube+"-"),
		StartWith(K8sNamespaceName, NsOpenshift+"-"),
		Eq(K8sNamespaceName, NsDefault),
		Eq(K8sNamespaceName, NsOpenshift),
		Eq(K8sNamespaceName, NsKube))
	AppContainerLogs = Neg(Paren(InfraContainerLogs))

	AddLogTypeApp   = fmt.Sprintf(".log_type = %q", logging.InputNameApplication)
	AddLogTypeInfra = fmt.Sprintf(".log_type = %q", logging.InputNameInfrastructure)
	AddLogTypeAudit = fmt.Sprintf(".log_type = %q", logging.InputNameAudit)

	AddHostAuditTag = fmt.Sprintf(".tag = %q", HostAuditLogTag)
	AddK8sAuditTag  = fmt.Sprintf(".tag = %q", K8sAuditLogTag)
	AddOpenAuditTag = fmt.Sprintf(".tag = %q", OpenAuditLogTag)
	AddOvnAuditTag  = fmt.Sprintf(".tag = %q", OvnAuditLogTag)

	MatchNS = func(ns string) string {
		return Eq(K8sNamespaceName, ns)
	}
	K8sLabelKey = func(k string) string {
		return fmt.Sprintf(K8sLabelKeyExpr, k)
	}
	MatchLabel = func(k, v string) string {
		return Eq(K8sLabelKey(k), v)
	}
)

// Inputs takes the raw log sources (container, journal, audit) and produces Inputs as defined by ClusterLogForwarder Api
func Inputs(spec *logging.ClusterLogForwarderSpec, o Options) []Element {
	el := []Element{}

	types := GatherSources(spec, o)
	// route container_logs based on type
	if types.Has(logging.InputNameApplication) || types.Has(logging.InputNameInfrastructure) {
		r := Route{
			ComponentID: "route_container_logs",
			Inputs:      helpers.MakeInputs(InputContainerLogs),
			Routes:      map[string]string{},
		}
		if types.Has(logging.InputNameApplication) {
			r.Routes["app"] = Quote(AppContainerLogs)
		}
		if types.Has(logging.InputNameInfrastructure) {
			r.Routes["infra"] = Quote(InfraContainerLogs)
		}
		el = append(el, r)
	}

	if types.Has(logging.InputNameApplication) {
		el = append(el, Remap{
			Desc:        `Set log_type to "application"`,
			ComponentID: logging.InputNameApplication,
			Inputs:      helpers.MakeInputs("route_container_logs.app"),
			VRL:         AddLogTypeApp,
		})
	}
	if types.Has(logging.InputNameInfrastructure) {
		el = append(el, Remap{
			Desc:        `Set log_type to "infrastructure"`,
			ComponentID: logging.InputNameInfrastructure,
			Inputs:      helpers.MakeInputs("route_container_logs.infra", InputJournalLogs),
			VRL:         AddLogTypeInfra,
		})
	}
	if types.Has(logging.InputNameAudit) {
		el = append(el,
			Remap{
				Desc:        `Tag host audit files`,
				ComponentID: HostAuditLogID,
				Inputs:      helpers.MakeInputs(HostAuditLogs),
				VRL:         AddHostAuditTag,
			},
			Remap{
				Desc:        `Tag k8s audit files`,
				ComponentID: K8sAuditLogID,
				Inputs:      helpers.MakeInputs(K8sAuditLogs),
				VRL: strings.Join(helpers.TrimSpaces([]string{
					AddK8sAuditTag,
					ParseAndFlatten,
				}), "\n"),
			},
			Remap{
				Desc:        `Tag openshift audit files`,
				ComponentID: OpenAuditLogID,
				Inputs:      helpers.MakeInputs(OpenShiftAuditLogs),
				VRL: strings.Join(helpers.TrimSpaces([]string{
					AddOpenAuditTag,
					ParseAndFlatten,
				}), "\n"),
			},
			Remap{
				Desc:        `Tag ovn audit files`,
				ComponentID: OvnAuditLogID,
				Inputs:      helpers.MakeInputs(OvnAuditLogs),
				VRL:         AddOvnAuditTag,
			},
			Remap{
				Desc:        `Set log_type to "audit"`,
				ComponentID: logging.InputNameAudit,
				Inputs:      helpers.MakeInputs(HostAuditLogID, K8sAuditLogID, OpenAuditLogID, OvnAuditLogID),
				VRL: strings.Join(helpers.TrimSpaces([]string{
					AddLogTypeAudit,
					FixTimestampField,
				}), "\n"),
			})
	}

	userDefinedAppRouteMap := UserDefinedAppRouting(spec, o)
	if len(userDefinedAppRouteMap) != 0 {
		el = append(el, Route{
			ComponentID: RouteApplicationLogs,
			Inputs:      helpers.MakeInputs(logging.InputNameApplication),
			Routes:      userDefinedAppRouteMap,
		})
	}

	return el
}

func UserDefinedAppRouting(spec *logging.ClusterLogForwarderSpec, o Options) map[string]string {
	userDefined := spec.InputMap()
	routeMap := map[string]string{}
	for _, pipeline := range spec.Pipelines {
		for _, inRef := range pipeline.InputRefs {
			if input, ok := userDefined[inRef]; ok {
				// user defined input
				if input.Application != nil {
					app := input.Application
					matchNS := []string{}
					if len(app.Namespaces) != 0 {
						for _, ns := range app.Namespaces {
							matchNS = append(matchNS, MatchNS(ns))
						}
					}
					matchLabels := []string{}
					if app.Selector != nil && len(app.Selector.MatchLabels) != 0 {
						labels := app.Selector.MatchLabels
						keys := make([]string, 0, len(labels))
						for k := range labels {
							keys = append(keys, k)
						}
						sort.Strings(keys)
						for _, k := range keys {
							matchLabels = append(matchLabels, MatchLabel(k, labels[k]))
						}
					}
					if len(matchNS) != 0 || len(matchLabels) != 0 {
						routeMap[input.Name] = Quote(AND(OR(matchNS...), AND(matchLabels...)))
					}
				}
			}
		}
	}
	return routeMap
}
