package sources

type KubernetesLogs struct {

	// Type is required to be 'kubernetes_logs'
	Type SourceType `json:"type" yaml:"type" toml:"type"`

	MaxReadBytes              uint     `json:"max_read_bytes,omitempty" toml:"max_read_bytes,omitempty"`
	GlobMinimumCooldownMillis uint     `json:"glob_minimum_cooldown_ms,omitempty" toml:"glob_minimum_cooldown_ms,omitempty"`
	AutoPartialMerge          bool     `json:"auto_partial_merge,omitempty" toml:"auto_partial_merge,omitempty" toml:"auto_partial_merge,omitempty"`
	MaxMergedLineBytes        uint64   `json:"max_merged_line_bytes,omitempty" toml:"max_merged_line_bytes,omitempty"`
	IncludePathsGlobPatterns  []string `json:"include_paths_glob_patterns,omitempty" toml:"include_paths_glob_patterns,omitempty"`
	ExcludePathsGlobPatterns  []string `json:"exclude_paths_glob_patterns,omitempty" toml:"exclude_paths_glob_patterns,omitempty"`
	ExtraLabelSelector        string   `json:"extra_label_selector,omitempty" toml:"extra_label_selector,omitempty"`
	RotateWaitSecs            uint     `json:"rotate_wait_secs,omitempty" toml:"rotate_wait_secs,omitempty"`
	UseApiServerCache         bool     `json:"use_apiserver_cache,omitempty" toml:"use_apiserver_cache,omitempty"`

	PodAnnotationFields       *PodAnnotationFields       `json:"pod_annotation_fields,omitempty" toml:"pod_annotation_fields,omitempty"`
	NamespaceAnnotationFields *NamespaceAnnotationFields `json:"namespace_annotation_fields,omitempty" toml:"namespace_annotation_fields,omitempty"`
}

func NewKubernetesLogs(init func(logs *KubernetesLogs)) *KubernetesLogs {
	k := &KubernetesLogs{
		Type: SourceTypeKubernetesLogs,
	}
	if init != nil {
		init(k)
	}
	return k
}

type PodAnnotationFields struct {
	PodLabels      string `json:"pod_labels,omitempty" toml:"pod_labels,omitempty"`
	PodNamespace   string `json:"pod_namespace,omitempty" toml:"pod_namespace,omitempty"`
	PodAnnotations string `json:"pod_annotations,omitempty" toml:"pod_annotations,omitempty"`
	PodUid         string `json:"pod_uid,omitempty" toml:"pod_uid,omitempty"`
	PodNodeName    string `json:"pod_node_name,omitempty" toml:"pod_node_name,omitempty"`
}

type NamespaceAnnotationFields struct {
	NamespaceUid string `json:"namespace_uid,omitempty" toml:"namespace_uid,omitempty"`
}
