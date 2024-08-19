package viaq

const (
	ClusterID            = `.openshift.cluster_id = "${OPENSHIFT_CLUSTER_ID:-}"`
	SetTimestampField    = `if !exists(."@timestamp") {."@timestamp" = .timestamp}`
	Message              = `.message = ._internal.message`
	SetOpenShift         = `if exists(._internal.openshift) {.openshift = ._internal.openshift}`
	VRLOpenShiftSequence = `.openshift.sequence = to_unix_timestamp(now(), unit: "nanoseconds")`
)
