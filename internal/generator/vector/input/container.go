package input

import (
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	internalobs "github.com/openshift/cluster-logging-operator/internal/api/observability"
	"github.com/openshift/cluster-logging-operator/internal/generator/forwarder/adapters"
	commoncontainer "github.com/openshift/cluster-logging-operator/internal/generator/forwarder/input/container"
	"github.com/openshift/cluster-logging-operator/internal/generator/helpers"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/api"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/api/sources"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/api/types"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/set"
)

// NewContainerSource generates the source and transforms to support this input and
// returns an identifier id, source, transforms, and sets the list of ids to use for downstream components.
func NewContainerSource(spec *adapters.Input, includes, excludes []string, logType obs.InputType, logSource interface{}) (id string, source types.Source, tfs api.Transforms) {
	tfs = api.Transforms{}
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
	source = sources.NewKubernetesLogs(func(kl *sources.KubernetesLogs) {
		kl.MaxReadBytes = commoncontainer.MaxReadBytes
		kl.GlobMinimumCooldownMillis = 15000
		kl.AutoPartialMerge = true
		kl.MaxMergedLineBytes = uint64(maxMsgSize)
		kl.IncludePathsGlobPatterns = includes
		kl.ExcludePathsGlobPatterns = excludes
		kl.ExtraLabelSelector = helpers.LabelSelectorFrom(selector)
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
		// Node labels are not part of the output data model (the VIAQ normalize
		// filter deletes them), so suppress capturing them at the source.
		// Capturing them only inflates memory usage proportional to the number
		// of labels on the node and can OOMKill the collector (LOG-10533).
		kl.NodeAnnotationFields = &sources.NodeAnnotationFields{
			NodeLabels: "",
		}
		kl.RotateWaitSecs = 5
		kl.UseApiServerCache = true
	})
	tfs.Add(metaID, NewInternalNormalization(logSource, logType, base))
	id = metaID

	//TODO: DETERMINE IF key field is correct and actually works
	if threshold, hasPolicy := internalobs.MaxRecordsPerSecond(spec.InputSpec); hasPolicy {
		throttleID := helpers.MakeID(base, "throttle")
		id = throttleID
		tfs.Add(throttleID, AddThrottleToInput(metaID, threshold))
	}
	spec.Ids = append(spec.Ids, id)
	return base, source, tfs
}
