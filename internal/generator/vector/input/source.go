package input

import (
	"fmt"
	"regexp"

	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	internalobs "github.com/openshift/cluster-logging-operator/internal/api/observability"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	"github.com/openshift/cluster-logging-operator/internal/factory"
	"github.com/openshift/cluster-logging-operator/internal/generator/framework"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/elements"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/helpers"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/source"
	"github.com/openshift/cluster-logging-operator/internal/utils/sets"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/set"
)

const (
	nsPodPathFmt       = "%s_%s-*/*/*.log"
	nsContainerPathFmt = "%s_*/%s/*.log"
)

var (
	//// TODO: Remove ES/Kibana from excludes
	loggingExcludes = source.NewContainerPathGlobBuilder().
			AddOther(
			fmt.Sprintf(nsPodPathFmt, constants.OpenshiftNS, constants.LogfilesmetricexporterName),
			fmt.Sprintf(nsPodPathFmt, constants.OpenshiftNS, constants.ElasticsearchName),
			fmt.Sprintf(nsPodPathFmt, constants.OpenshiftNS, constants.KibanaName),
		).AddOther(
		fmt.Sprintf(nsContainerPathFmt, constants.OpenshiftNS, "loki*"),
		fmt.Sprintf(nsContainerPathFmt, constants.OpenshiftNS, "gateway"),
		fmt.Sprintf(nsContainerPathFmt, constants.OpenshiftNS, "opa"),
	).AddExtensions(excludeExtensions...).
		Build()
	excludeExtensions = []string{"gz", "tmp", "log.*"}
	infraNamespaces   = []string{"default", "openshift", "openshift-*", "kube", "kube-*"}
	infraNSRegex      = regexp.MustCompile(
		`(?P<default>^default$)|(?P<openshift_dash>^openshift-.+$)|(?P<kube_dash>^kube-.+$)|(?P<openshift>^openshift$)|(?P<kube>^kube$)|(?P<openshift_all>^openshift\*$)|(?P<kube_all>^kube\*$)`)
)

// NewSource creates an input adapter to generate config for ViaQ sources to collect logs excluding the
// collector container logs from the namespace where the collector is deployed
func NewSource(input obs.InputSpec, collectorNS string, resNames factory.ForwarderResourceNames, secrets internalobs.Secrets, op framework.Options) ([]framework.Element, []string) {
	els := []framework.Element{}
	ids := []string{}
	// LOG-7196 temporary fix to set vector caching config
	// TODO: remove annotation logic and add to spec
	switch input.Type {
	case obs.InputTypeApplication:
		ib := source.NewContainerPathGlobBuilder()
		eb := source.NewContainerPathGlobBuilder()
		appIncludes := []string{}
		if input.Application != nil {
			if len(input.Application.Includes) > 0 {
				for _, in := range input.Application.Includes {
					ncs := source.NamespaceContainer{
						Namespace: in.Namespace,
						Container: in.Container,
					}
					ib.AddCombined(ncs)
					appIncludes = append(appIncludes, ncs.Namespace)
				}
			}
			// Need to remove any of the default excluded infra namespaces if they are part of the includes
			excludesList := pruneInfraNS(appIncludes)
			for _, ns := range excludesList {
				ncs := source.NamespaceContainer{
					Namespace: ns,
				}
				eb.AddCombined(ncs)
			}
			if len(input.Application.Excludes) > 0 {
				for _, ex := range input.Application.Excludes {
					ncs := source.NamespaceContainer{
						Namespace: ex.Namespace,
						Container: ex.Container,
					}
					eb.AddCombined(ncs)
				}
			}
		} else {
			// Need to remove any of the default excluded infra namespaces if they are part of the includes
			excludesList := pruneInfraNS(appIncludes)
			for _, ns := range excludesList {
				ncs := source.NamespaceContainer{
					Namespace: ns,
				}
				eb.AddCombined(ncs)
			}
		}
		eb.AddExtensions(excludeExtensions...)
		includes := ib.Build()
		excludes := eb.Build(infraNamespaces...)
		return NewContainerSource(input, collectorNS, includes, excludes, obs.InputTypeApplication, obs.InfrastructureSourceContainer)
	case obs.InputTypeInfrastructure:
		sources := set.Set[obs.InfrastructureSource]{}
		if input.Infrastructure == nil {
			sources.Insert(obs.InfrastructureSources...)
		} else {
			sources = set.New(input.Infrastructure.Sources...)
			if sources.Len() == 0 {
				sources.Insert(obs.InfrastructureSources...)
			}
		}
		if sources.Has(obs.InfrastructureSourceContainer) {
			infraIncludes := source.NewContainerPathGlobBuilder().AddNamespaces(infraNamespaces...).Build()
			cels, cids := NewContainerSource(input, collectorNS, infraIncludes, loggingExcludes, obs.InputTypeInfrastructure, obs.InfrastructureSourceContainer)
			els = append(els, cels...)
			ids = append(ids, cids...)
		}
		if sources.Has(obs.InfrastructureSourceNode) {
			jels, jids := NewJournalSource(input)
			els = append(els, jels...)
			ids = append(ids, jids...)
		}
		return els, ids
	case obs.InputTypeAudit:
		sources := set.Set[obs.AuditSource]{}
		if input.Audit == nil || len(input.Audit.Sources) == 0 {
			sources.Insert(obs.AuditSources...)
		} else {
			sources = set.New(input.Audit.Sources...)
			if sources.Len() == 0 {
				sources.Insert(obs.AuditSources...)
			}
		}
		if sources.Has(obs.AuditSourceAuditd) {
			cels, cids := NewAuditAuditdSource(input, op)
			els = append(els, cels...)
			ids = append(ids, cids...)
		}
		if sources.Has(obs.AuditSourceKube) {
			cels, cids := NewK8sAuditSource(input, op)
			els = append(els, cels...)
			ids = append(ids, cids...)
		}
		if sources.Has(obs.AuditSourceOpenShift) {
			cels, cids := NewOpenshiftAuditSource(input, op)
			els = append(els, cels...)
			ids = append(ids, cids...)
		}
		if sources.Has(obs.AuditSourceOVN) {
			cels, cids := NewOVNAuditSource(input, op)
			els = append(els, cels...)
			ids = append(ids, cids...)
		}
		return els, ids
	case obs.InputTypeReceiver:
		return NewViaqReceiverSource(input, resNames, secrets, op)
	}
	return els, ids
}

// NewContainerSource generates config elements and the id reference of this input and normalizes
func NewContainerSource(spec obs.InputSpec, namespace, includes, excludes string, logType obs.InputType, logSource interface{}) ([]framework.Element, []string) {
	base := helpers.MakeInputID(spec.Name, "container")
	var selector *metav1.LabelSelector
	maxMsgSize := int64(0)
	if spec.Application != nil {
		selector = spec.Application.Selector
		if spec.Application.Tuning != nil && spec.Application.Tuning.MaxMessageSize != nil {
			if size, ok := spec.Application.Tuning.MaxMessageSize.AsInt64(); ok {
				maxMsgSize = size
			}
		}
	}
	if spec.Infrastructure != nil {
		if (len(spec.Infrastructure.Sources) == 0 || set.New(spec.Infrastructure.Sources...).Has(obs.InfrastructureSourceContainer)) &&
			spec.Infrastructure.Tuning != nil && spec.Infrastructure.Tuning.Container != nil && spec.Infrastructure.Tuning.Container.MaxMessageSize != nil {
			if size, ok := spec.Infrastructure.Tuning.Container.MaxMessageSize.AsInt64(); ok {
				maxMsgSize = size
			}
		}
	}
	metaID := helpers.MakeID(base, "meta")
	k8sLogs := source.NewKubernetesLogs(base, includes, excludes, maxMsgSize)
	k8sLogs.ExtraLabelSelector = source.LabelSelectorFrom(selector)
	el := []framework.Element{
		k8sLogs,
		NewLogSourceAndType(metaID, logSource, logType, base, func(remap *elements.Remap) {
			remap.VRL = fmt.Sprintf(
				`
.log_source = %q
# If namespace is infra, label log_type as infra
if match_any(string!(.kubernetes.namespace_name), [r'^default$', r'^openshift(-.+)?$', r'^kube(-.+)?$']) {
    .log_type = %q
} else {
    .log_type = %q
}`,
				logSource,
				obs.InputTypeInfrastructure,
				obs.InputTypeApplication)
		}),
	}
	inputID := metaID
	//TODO: DETERMINE IF key field is correct and actually works
	if threshold, hasPolicy := internalobs.MaxRecordsPerSecond(spec); hasPolicy {
		throttleID := helpers.MakeID(base, "throttle")
		inputID = throttleID
		el = append(el, AddThrottleToInput(throttleID, metaID, threshold)...)
	}

	return el, []string{inputID}
}

// pruneInfraNS returns a pruned infra namespace list depending on which infra namespaces were included
// since the exclusion list includes all infra namespaces by default
// Example:
// Include: ["openshift-logging"]
// Default Exclude: ["default", "openshift", "openshift-*", "kube", "kube-*"]
// Final infra namespaces in Exclude: ["default", "openshift", "kube", "kube-*"]
func pruneInfraNS(includes []string) []string {
	foundInfraNamespaces := make(map[string]string)
	for _, ns := range includes {
		matches := infraNSRegex.FindStringSubmatch(ns)
		if matches != nil {
			for i, name := range infraNSRegex.SubexpNames() {
				if i != 0 && matches[i] != "" {
					foundInfraNamespaces[name] = matches[i]
				}
			}
		}
	}

	infraNSSet := sets.NewString(infraNamespaces...)
	// Remove infra namespace depending on the named capture group
	for k := range foundInfraNamespaces {
		switch k {
		case "default":
			infraNSSet.Remove("default")
		case "openshift":
			infraNSSet.Remove("openshift")
		case "openshift_dash":
			infraNSSet.Remove("openshift-*")
		case "openshift_all":
			infraNSSet.Remove("openshift")
			infraNSSet.Remove("openshift-*")
		case "kube":
			infraNSSet.Remove("kube")
		case "kube_dash":
			infraNSSet.Remove("kube-*")
		case "kube_all":
			infraNSSet.Remove("kube")
			infraNSSet.Remove("kube-*")
		}
	}
	return infraNSSet.List()
}
