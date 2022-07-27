package vector

import (
	"fmt"
	"sort"
	"strings"

	"github.com/openshift/cluster-logging-operator/internal/generator/vector/normalize"

	logging "github.com/openshift/cluster-logging-operator/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/internal/generator"
	. "github.com/openshift/cluster-logging-operator/internal/generator"
	. "github.com/openshift/cluster-logging-operator/internal/generator/vector/elements"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/helpers"
)

const (
	NsKube      = "kube"
	NsOpenshift = "openshift"
	NsDefault   = "default"

	K8sNamespaceName = ".kubernetes.namespace_name"
	K8sLabelKeyExpr  = ".kubernetes.labels.%q"

	InputContainerLogs = "container_logs"
	InputJournalLogs   = "journal_logs"

	RouteApplicationLogs    = "route_application_logs"
	SourceTransformThrottle = "source_throttle"

	SrcPassThrough = "."
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

	UserDefinedInput          = fmt.Sprintf("%s.%%s", RouteApplicationLogs)
	UserDefinedSourceThrottle = fmt.Sprintf("%s_%%s", SourceTransformThrottle)
	perContainerLimitKeyField = `"{{ file }}"`

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

func AddThrottle(spec *logging.InputSpec) []generator.Element {
	var (
		threshold    int64
		throttle_key string
	)

	el := []generator.Element{}
	input := fmt.Sprintf(UserDefinedInput, spec.Name)

	if spec.Application.ContainerLimit != nil {
		threshold = spec.Application.ContainerLimit.MaxRecordsPerSecond
		throttle_key = perContainerLimitKeyField
	} else {
		threshold = spec.Application.GroupLimit.MaxRecordsPerSecond
	}

	el = append(el, Throttle{
		ComponentID: fmt.Sprintf(UserDefinedSourceThrottle, spec.Name),
		Inputs:      helpers.MakeInputs([]string{input}...),
		Threshold:   threshold,
		KeyField:    throttle_key,
	})

	return el
}

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
				Desc:        `Set log_type to "audit"`,
				ComponentID: logging.InputNameAudit,
				Inputs:      helpers.MakeInputs(HostAuditLogs, K8sAuditLogs, OpenshiftAuditLogs, OvnAuditLogs),
				VRL: strings.Join(helpers.TrimSpaces([]string{
					AddLogTypeAudit,
					FixHostname,
					normalize.FixTimestampField,
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

		userDefined := spec.InputMap()
		for inRef := range userDefinedAppRouteMap {
			if input, ok := userDefined[inRef]; ok && input.HasPolicy() && input.GetMaxRecordsPerSecond() > 0 {
				// Vector Throttle component cannot have zero threshold
				el = append(el, AddThrottle(input)...)
			}
		}
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
					} else if input.HasPolicy() {
						routeMap[input.Name] = "'true'"
					}
				}
			}
		}
	}
	return routeMap
}
