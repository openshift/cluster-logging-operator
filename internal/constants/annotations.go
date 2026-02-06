package constants

const (
	AnnotationDebugOutput = "logging.openshift.io/debug-output"

	// AnnotationEnableCollectorAsDeployment is to enable deploying the collector as a deployment
	// instead of a daemonset to support the HCP use case of using the collector for collecting
	// audit logs via a webhook.
	AnnotationEnableCollectorAsDeployment = "logging.openshift.io/dev-preview-enable-collector-as-deployment"

	// AnnotationVectorLogLevel is used to set the log level of vector.
	// Log level can be one of: trace, debug, info, warn, error, off.
	// CLO's default log level for vector is `warn`: https://issues.redhat.com/browse/LOG-3435
	AnnotationVectorLogLevel = "observability.openshift.io/log-level"

	AnnotationSecretHash = "observability.openshift.io/secret-hash"

	// AnnotationMaxUnavailable (Deprecated) configures the maximum number of DaemonSet pods that can be unavailable during a rolling update.
	// This can be an absolute number (e.g., 1) or a percentage (e.g., 10%). Default is 100%.
	AnnotationMaxUnavailable = "observability.openshift.io/max-unavailable-rollout"
)
