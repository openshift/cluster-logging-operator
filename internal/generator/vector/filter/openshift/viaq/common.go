package viaq

const (
	ClusterID            = `.openshift.cluster_id = "${OPENSHIFT_CLUSTER_ID:-}"`
	FixTimestampField    = `if !exists(."@timestamp") {."@timestamp" = .timestamp}`
	InternalContext      = `._internal.message = .message`
	VRLOpenShiftSequence = `.openshift.sequence = to_unix_timestamp(now(), unit: "nanoseconds")`
)
