package input

import (
	"fmt"
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	internalobs "github.com/openshift/cluster-logging-operator/internal/api/observability"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	"github.com/openshift/cluster-logging-operator/internal/factory"
	"github.com/openshift/cluster-logging-operator/internal/generator/framework"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/helpers"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/source"
	"github.com/openshift/cluster-logging-operator/internal/utils/sets"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/set"
	"regexp"
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
	excludeExtensions = []string{"gz", "tmp"}
	infraNamespaces   = []string{"default", "openshift*", "kube*"}
	infraNSRegex      = regexp.MustCompile(`^(?P<default>default)|(?P<openshift>openshift.*)|(?P<kube>kube.*)$`)
)

// NewSource creates an input adapter to generate config for ViaQ sources to collect logs excluding the
// collector container logs from the namespace where the collector is deployed
func NewSource(input obs.InputSpec, collectorNS string, resNames factory.ForwarderResourceNames, secrets helpers.Secrets, op framework.Options) ([]framework.Element, []string) {
	els := []framework.Element{}
	ids := []string{}
	switch input.Type {
	case obs.InputTypeApplication:
		ib := source.NewContainerPathGlobBuilder()
		eb := source.NewContainerPathGlobBuilder()
		appIncludes := []string{}
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
		eb.AddExtensions(excludeExtensions...)
		includes := ib.Build()
		excludes := eb.Build(infraNamespaces...)
		return NewContainerSource(input, collectorNS, includes, excludes, obs.InputTypeApplication, obs.InfrastructureSourceContainer)
	case obs.InputTypeInfrastructure:
		sources := set.New(input.Infrastructure.Sources...)
		if sources.Len() == 0 {
			sources.Insert(obs.InfrastructureSources...)
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
		sources := set.New(input.Audit.Sources...)
		if sources.Len() == 0 {
			sources.Insert(obs.AuditSources...)
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
func NewContainerSource(spec obs.InputSpec, namespace, includes, excludes string, logType, logSource interface{}) ([]framework.Element, []string) {
	base := helpers.MakeInputID(spec.Name, "container")
	var selector *metav1.LabelSelector
	if spec.Application != nil {
		selector = spec.Application.Selector
	}
	metaID := helpers.MakeID(base, "meta")
	el := []framework.Element{
		source.KubernetesLogs{
			ComponentID:        base,
			Desc:               "Logs from containers (including openshift containers)",
			IncludePaths:       includes,
			ExcludePaths:       excludes,
			ExtraLabelSelector: source.LabelSelectorFrom(selector),
		},
		NewLogSourceAndType(metaID, logSource, logType, base),
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
// Default Exclude: ["default", "openshift*", "kube*"]
// Final infra namespaces in Exclude: ["default", "kube*"]
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
			infraNSSet.Remove("openshift*")
		case "kube":
			infraNSSet.Remove("kube*")
		}
	}
	return infraNSSet.List()
}
