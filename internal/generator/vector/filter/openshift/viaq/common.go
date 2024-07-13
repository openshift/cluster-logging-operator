package viaq

const (
	ClusterID         = `.openshift.cluster_id = "${OPENSHIFT_CLUSTER_ID:-}"`
	FixTimestampField = `ts = del(.timestamp); if !exists(."@timestamp") {."@timestamp" = ts}`
	InternalContext   = `
._internal.message = .message
._internal.file = .file
`
)
