package input

import (
	"fmt"

	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	internalobs "github.com/openshift/cluster-logging-operator/internal/api/observability"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	"github.com/openshift/cluster-logging-operator/internal/generator/framework"
	helpers2 "github.com/openshift/cluster-logging-operator/internal/generator/helpers"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/api"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/api/sources"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/helpers"
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
	loggingExcludes = helpers2.NewContainerPathGlobBuilder().
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
	infraNamespaces   = []string{"default", "openshift*", "kube*"}
)

// NewContainerSource generates config elements and the id reference of this input and normalizes
func NewContainerSource(spec obs.InputSpec, includes, excludes []string, logType obs.InputType, logSource interface{}) ([]framework.Element, []string) {
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
	k8sLogs := sources.NewKubernetesLogs(func(kl *sources.KubernetesLogs) {
		kl.MaxReadBytes = 3145728
		kl.GlobMinimumCooldownMillis = 15000
		kl.AutoPartialMerge = true
		kl.MaxMergedLineBytes = uint64(maxMsgSize)
		kl.IncludePathsGlobPatterns = includes
		kl.ExcludePathsGlobPatterns = excludes
		kl.ExtraLabelSelector = helpers2.LabelSelectorFrom(selector)
		kl.PodAnnotationFields = &sources.PodAnnotationFields{
			PodLabels:      "kubernetes.labels",
			PodNamespace:   "kubernetes.namespace_name",
			PodAnnotations: "kubernetes.annotations",
			PodUid:         "kubernetes.pod_id",
			PodNodeName:    "hostname",
		}
		kl.NamespaceAnnotationFields = &sources.NamespaceAnnotationFields{
			NamespaceUid: "kubernetes.namespace_id",
		}
		kl.RotateWaitSecs = 5
		kl.UseApiServerCache = true
	})
	el := []framework.Element{
		api.Config{
			Sources: map[string]interface{}{
				base: k8sLogs,
			},
		},
		NewInternalNormalization(metaID, logSource, logType, base),
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
		matches := internalobs.InfraNSRegex.FindStringSubmatch(ns)
		if matches != nil {
			for i, name := range internalobs.InfraNSRegex.SubexpNames() {
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
