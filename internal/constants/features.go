package constants

const (
	Enabled = "enabled"

	// PreviewTLSSecurityProfile is the annotation to enable TLS security profiles
	PreviewTLSSecurityProfile = "logging.openshift.io/preview-tls-security-profile"

	// UseOldRemoteSyslogPlugin use old syslog plugin (docebo/fluent-plugin-remote-syslog)
	// +deprecated
	UseOldRemoteSyslogPlugin = "clusterlogging.openshift.io/useoldremotesyslogplugin"

	AnnotationDebugOutput = "logging.openshift.io/debug-output"
)
