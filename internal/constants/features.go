package constants

const (
	Enabled = "enabled"

	AnnotationDebugOutput = "logging.openshift.io/debug-output"

	// AnnotationEnableSchema is the annotation to enable alternate output formats of logs.
	// Currently only viaq & opentelemetry are supported
	AnnotationEnableSchema = "logging.openshift.io/enableschema"

	// AnnotationEnableCollectorAsDeployment is to enable deploying the collector as a deployment
	// instead of a daemonset to support the HCP use case of using the collector for collecting
	// audit logs via a webhook.
	AnnotationEnableCollectorAsDeployment = "logging.openshift.io/dev-preview-enable-collector-as-deployment"

	// AnnotationVectorLogLevel is used to set the log level of vector.
	// Log level can be one of: trace, debug, info, warn, error, off.
	// CLO's default log level for vector is `warn`: https://issues.redhat.com/browse/LOG-3435
	AnnotationVectorLogLevel = "observability.openshift.io/log-level"
)
