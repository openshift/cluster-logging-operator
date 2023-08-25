package constants

const (
	Enabled = "enabled"

	// UseOldRemoteSyslogPlugin use old syslog plugin (docebo/fluent-plugin-remote-syslog)
	// +deprecated
	UseOldRemoteSyslogPlugin = "clusterlogging.openshift.io/useoldremotesyslogplugin"

	AnnotationDebugOutput = "logging.openshift.io/debug-output"

	// OpenTelemetry is the annotation to enable OpenTelemetry output
	OpenTelemetry = "logging.openshift.io/opentelemetry"
)
