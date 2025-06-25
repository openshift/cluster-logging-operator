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

	// AnnotationOtlpOutputTechPreview is used to enable tech-preview of OTLP output and LokiStack with Otel data model
	AnnotationOtlpOutputTechPreview = "observability.openshift.io/tech-preview-otlp-output"

	AnnotationSecretHash = "observability.openshift.io/secret-hash"

	// AnnotationKubeCache is used to enable caching for requests to the kube-apiserver using vector kubernetes_logs source.
	// Tech-Preview feature
	//
	// While enabling cache can significantly reduce Kubernetes control plane
	// memory pressure, the trade-off is a chance of receiving stale data.
	AnnotationKubeCache = "observability.openshift.io/use-apiserver-cache"

	// AnnotationMaxUnavailable configures the maximum number of DaemonSet pods that can be unavailable during a rolling update.
	// Tech-Preview feature
	//
	// This can be an absolute number (e.g., 1) or a percentage (e.g., 10%). Default is 100%.
	AnnotationMaxUnavailable = "observability.openshift.io/max-unavailable-rollout"
)
