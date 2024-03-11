package v1

// NOTE: The Enum validation on FilterSpec.Type must be updated if the list of
// known types changes.

// Filter type constants, must match JSON tags of FilterTypeSpec fields.
const (
	FilterKubeAPIAudit = "kubeAPIAudit"
	FilterDrop         = "drop"
	FilterPrune        = "prune"
)

// FilterTypeSpec is a union of filter specification types.
// The fields of this struct define the set of known filter types.
type FilterTypeSpec struct {
	// +optional
	KubeAPIAudit *KubeAPIAudit `json:"kubeAPIAudit,omitempty"`

	// NOTE more filter types expected in future, for example filtering on record fields (e.g. level).

	// A drop filter applies a sequence of tests to a log record and drops the record if any test passes.
	// Each test contains a sequence of conditions, all conditions must be true for the test to pass.
	// A DropTestsSpec contains an array of tests which contains an array of conditions
	// +optional
	DropTestsSpec *[]DropTest `json:"drop,omitempty"`

	// The PruneFilterSpec consists of two arrays, namely in and notIn, which dictate the fields to be pruned.
	// +optional
	PruneFilterSpec *PruneFilterSpec `json:"prune,omitempty"`
}

type DropTest struct {
	// DropConditions is an array of DropCondition which are conditions that are ANDed together
	// +optional
	DropConditions []DropCondition `json:"test,omitempty"`
}

type DropCondition struct {
	// A dot delimited path to a field in the log record. It must start with a `.`.
	// The path can contain alpha-numeric characters and underscores (a-zA-Z0-9_).
	// If segments contain characters outside of this range, the segment must be quoted.
	// Examples: `.kubernetes.namespace_name`, `.log_type`, '.kubernetes.labels.foobar', `.kubernetes.labels."foo-bar/baz"`
	// +optional
	Field string `json:"field,omitempty"`

	// A regular expression that the field will match.
	// If the value of the field defined in the DropTest matches the regular expression, the log record will be dropped.
	// Must define only one of matches OR notMatches
	// +optional
	Matches string `json:"matches,omitempty"`

	// A regular expression that the field does not match.
	// If the value of the field defined in the DropTest does not match the regular expression, the log record will be dropped.
	// Must define only one of matches or notMatches
	// +optional
	NotMatches string `json:"notMatches,omitempty"`
}

type PruneFilterSpec struct {
	// `In` is an array of dot-delimited field paths. Fields included here are removed from the log record.
	// Each field path expression must start with a `.`.
	// The path can contain alpha-numeric characters and underscores (a-zA-Z0-9_).
	// If segments contain characters outside of this range, the segment must be quoted otherwise paths do NOT need to be quoted.
	// Examples: `.kubernetes.namespace_name`, `.log_type`, '.kubernetes.labels.foobar', `.kubernetes.labels."foo-bar/baz"`
	// NOTE1: `In` CANNOT contain `.log_type` or `.message` as those fields are required and cannot be pruned.
	// NOTE2: If this filter is used in a pipeline with GoogleCloudLogging, `.hostname` CANNOT be added to this list as it is a required field.
	// +optional
	In []string `json:"in,omitempty"`

	// `NotIn` is an array of dot-delimited field paths. All fields besides the ones listed here are removed from the log record
	// Each field path expression must start with a `.`.
	// The path can contain alpha-numeric characters and underscores (a-zA-Z0-9_).
	// If segments contain characters outside of this range, the segment must be quoted otherwise paths do NOT need to be quoted.
	// Examples: `.kubernetes.namespace_name`, `.log_type`, '.kubernetes.labels.foobar', `.kubernetes.labels."foo-bar/baz"`
	// NOTE1: `NotIn` MUST contain `.log_type` and `.message` as those fields are required and cannot be pruned.
	// NOTE2: If this filter is used in a pipeline with GoogleCloudLogging, `.hostname` MUST be added to this list as it is a required field.
	// +optional
	NotIn []string `json:"notIn,omitempty"`
}
