package vector

import (
	"fmt"
	"sort"

	logging "github.com/openshift/cluster-logging-operator/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/internal/generator"
	. "github.com/openshift/cluster-logging-operator/internal/generator/vector/elements"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/helpers"
)

const (
	NsKube      = "kube"
	NsOpenshift = "openshift"
	NsDefault   = "default"

	K8sPodNamespace = ".kubernetes.namespace_name"
	K8sLabelKeyExpr = ".kubernetes.pod_labels.%s"

	InputContainerLogs   = "container_logs"
	InputJournalLogs     = "journal_logs"
	RouteApplicationLogs = "route_application_logs"

	SrcPassThrough = "."
)

var (
	InfraContainerLogs = OR(
		StartWith(K8sPodNamespace, NsKube),
		StartWith(K8sPodNamespace, NsOpenshift),
		Eq(K8sPodNamespace, NsDefault))
	AppContainerLogs = Neg(Paren(InfraContainerLogs))

	AddLogTypeApp   = fmt.Sprintf(".log_type = %q", logging.InputNameApplication)
	AddLogTypeInfra = fmt.Sprintf(".log_type = %q", logging.InputNameInfrastructure)
	AddLogTypeAudit = fmt.Sprintf(".log_type = %q", logging.InputNameAudit)

	MatchNS = func(ns string) string {
		return Eq(K8sPodNamespace, ns)
	}
	K8sLabelKey = func(k string) string {
		return fmt.Sprintf(K8sLabelKeyExpr, k)
	}
	MatchLabel = func(k, v string) string {
		return Eq(K8sLabelKey(k), v)
	}
)

// Inputs takes the raw log sources (container, journal, audit) and produces Inputs as defined by ClusterLogForwarder Api
func Inputs(spec *logging.ClusterLogForwarderSpec, o generator.Options) []generator.Element {
	el := []generator.Element{}

	types := generator.GatherSources(spec, o)
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
		r := Remap{
			Desc:        `Rename log stream to "application"`,
			ComponentID: logging.InputNameApplication,
			Inputs:      helpers.MakeInputs("route_container_logs.app"),
			VRL:         AddLogTypeApp,
		}
		el = append(el, r)
	}
	if types.Has(logging.InputNameInfrastructure) {
		r := Remap{
			Desc:        `Rename log stream to "infrastructure"`,
			ComponentID: logging.InputNameInfrastructure,
			Inputs:      helpers.MakeInputs("route_container_logs.infra", InputJournalLogs),
			VRL:         AddLogTypeInfra,
		}
		el = append(el, r)
	}
	if types.Has(logging.InputNameAudit) {
		r := Remap{
			Desc:        `Rename log stream to "audit"`,
			ComponentID: logging.InputNameAudit,
			Inputs:      helpers.MakeInputs("host_audit_logs", "k8s_audit_logs", "openshift_audit_logs", "ovn_audit_logs"),
			VRL:         AddLogTypeAudit,
		}
		el = append(el, r)
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

func UserDefinedAppRouting(spec *logging.ClusterLogForwarderSpec, o generator.Options) map[string]string {
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
