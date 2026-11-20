package sources

import (
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/api/types"
)

type KubernetesLogs struct {

	// Type is required to be 'kubernetes_logs'
	Type                      types.SourceType           `json:"type"`
	MaxReadBytes              uint                       `json:"max_read_bytes,omitempty"`
	GlobMinimumCooldownMillis uint                       `json:"glob_minimum_cooldown_ms,omitempty"`
	AutoPartialMerge          bool                       `json:"auto_partial_merge,omitempty"`
	MaxMergedLineBytes        uint64                     `json:"max_merged_line_bytes,omitempty"`
	IncludePathsGlobPatterns  []string                   `json:"include_paths_glob_patterns,omitempty"`
	ExcludePathsGlobPatterns  []string                   `json:"exclude_paths_glob_patterns,omitempty"`
	ExtraLabelSelector        string                     `json:"extra_label_selector,omitempty"`
	RotateWaitSecs            uint                       `json:"rotate_wait_secs,omitempty"`
	UseApiServerCache         bool                       `json:"use_apiserver_cache,omitempty"`
	PodAnnotationFields       *PodAnnotationFields       `json:"pod_annotation_fields,omitempty"`
	NamespaceAnnotationFields *NamespaceAnnotationFields `json:"namespace_annotation_fields,omitempty"`
	NodeAnnotationFields      *NodeAnnotationFields      `json:"node_annotation_fields,omitempty"`
}

func NewKubernetesLogs(init func(logs *KubernetesLogs)) *KubernetesLogs {
	k := &KubernetesLogs{
		Type: types.SourceTypeKubernetesLogs,
	}
	if init != nil {
		init(k)
	}
	return k
}

func (k *KubernetesLogs) SourceType() types.SourceType {
	return k.Type
}

type PodAnnotationFields struct {
	PodLabels      string `json:"pod_labels,omitempty"`
	PodNamespace   string `json:"pod_namespace,omitempty"`
	PodAnnotations string `json:"pod_annotations,omitempty"`
	PodUid         string `json:"pod_uid,omitempty"`
	PodNodeName    string `json:"pod_node_name,omitempty"`
}

type NamespaceAnnotationFields struct {
	NamespaceUid string `json:"namespace_uid,omitempty"`
}

// NodeAnnotationFields controls which node metadata fields the kubernetes_logs
// source captures. NodeLabels must NOT be tagged omitempty: an empty string is
// meaningful — it tells Vector to suppress capturing node labels, which are
// otherwise attached to every event and only deleted later by the VIAQ
// normalize filter (see LOG-10533).
type NodeAnnotationFields struct {
	NodeLabels string `json:"node_labels"`
}
