package v1

// NOTE: The Enum validation on FilterSpec.Type must be updated if the list of
// known types changes.

// Filter type constants, must match JSON tags of FilterTypeSpec fields.
const (
	FilterKubeAPIAudit = "kubeAPIAudit"
)

// FilterTypeSpec is a union of filter specification types.
// The fields of this struct define the set of known filter types.
type FilterTypeSpec struct {
	// +optional
	KubeAPIAudit *KubeAPIAudit `json:"kubeAPIAudit,omitempty"`

	// NOTE more filter types expected in future, for example filtering on record fields (e.g. level).
}
