package constants

const (
	// Keys used in ClusterLogForwarder.Output Secrets keys.
	// Documented with OutputSpec.Secret in ../../apis/logging/v1/cluster_log_forwarder_types.go
	//
	// WARNING: changing or removing values here is a breaking API change.

	// TLS keys, used by any output that supports TLS.

	ClientCertKey      = "tls.crt"
	ClientPrivateKey   = "tls.key"
	TrustedCABundleKey = "ca-bundle.crt"
	Passphrase         = "passphrase"
	BearerTokenFileKey = "token"
	TLSInsecure        = "tls.insecure"

	// Username/Password keys, used by any output with username/password authentication.

	ClientUsername = "username"
	ClientPassword = "password"

	// SASL keys, used by any output that supports SASL.

	SASLEnable        = "sasl.enable"
	SASLMechanisms    = "sasl.mechanisms"
	SASLAllowInsecure = "sasl.allow-insecure"

	// Output-specific keys

	SharedKey                   = "shared_key"            // fluent forward
	DeprecatedSaslOverSSL       = "sasl_over_ssl"         // Kafka
	AWSSecretAccessKey          = "aws_secret_access_key" //nolint:gosec
	AWSAccessKeyID              = "aws_access_key_id"
	AWSRoleSessionName          = "cluster-logging" // identifier for role logging session
	AWSCredentialsKey           = "credentials"     // credrequest key to check for sts-formatted secret
	AWSWebIdentityRoleKey       = "role_arn"        // manual key to check for sts-formatted secret
	AWSWebIdentityTokenName     = "collector-sts-token"
	AWSWebIdentityTokenMount    = "/var/run/secrets/openshift/serviceaccount" //nolint:gosec // default location for volume mount
	AWSWebIdentityTokenFilePath = "token"                                     // file containing token relative to mount

	TokenKey          = "token"
	LogCollectorToken = "logcollector-token"

	UnHealthyStatus = "0"
	HealthyStatus   = "1"
	UnManagedStatus = "0"
	ManagedStatus   = "1"
	IsPresent       = "1"
	IsNotPresent    = "0"

	SingletonName = "instance"
	OpenshiftNS   = "openshift-logging"
	// global proxy / trusted ca bundle consts
	ProxyName = "cluster"

	InjectTrustedCABundleLabel = "config.openshift.io/inject-trusted-cabundle"
	TrustedCABundleMountFile   = "tls-ca-bundle.pem"
	TrustedCABundleMountDir    = "/etc/pki/ca-trust/extracted/pem/"
	SecretHashPrefix           = "logging.openshift.io/"
	KibanaTrustedCAName        = "kibana-trusted-ca-bundle"
	ElasticsearchFQDN          = "elasticsearch"
	ElasticsearchName          = "elasticsearch"
	ElasticsearchPort          = "9200"
	FluentdName                = "fluentd"
	VectorName                 = "vector"
	KibanaName                 = "kibana"
	KibanaProxyName            = "kibana-proxy"
	CuratorName                = "curator"
	LogfilesmetricexporterName = "logfilesmetricexporter"
	ConsolePluginName          = "consoleplugin"
	LokiStackName              = "lokistack"
	LogStoreURL                = "https://" + ElasticsearchFQDN + ":" + ElasticsearchPort
	MasterCASecretName         = "master-certs"
	CollectorSecretName        = "collector"
	// Disable gosec linter, complains "possible hard-coded secret"
	CollectorSecretsDir     = "/var/run/ocp-collector/secrets" //nolint:gosec
	KibanaSessionSecretName = "kibana-session-secret"          //nolint:gosec

	CollectorName             = "collector"
	CollectorConfigSecretName = "collector-config"
	CollectorMetricSecretName = "collector-metrics"
	CollectorMonitorJobLabel  = "monitor-collector"
	CollectorTrustedCAName    = "collector-trusted-ca-bundle"

	CollectorServiceAccountName = "logcollector"

	LegacySecureforward = "_LEGACY_SECUREFORWARD"
	LegacySyslog        = "_LEGACY_SYSLOG"

	FluentdImageEnvVar            = "RELATED_IMAGE_FLUENTD"
	VectorImageEnvVar             = "RELATED_IMAGE_VECTOR"
	LogfilesmetricImageEnvVar     = "RELATED_IMAGE_LOG_FILE_METRIC_EXPORTER"
	ConsolePluginImageEnvVar      = "RELATED_IMAGE_LOGGING_CONSOLE_PLUGIN"
	CertEventName                 = "cluster-logging-certs-generate"
	ClusterInfrastructureInstance = "cluster"

	ContainerLogDir = "/var/log/containers"
	PodLogDir       = "/var/log/pods"

	// Annotation Names
	AnnotationServingCertSecretName = "service.alpha.openshift.io/serving-cert-secret-name"

	// K8s recommended label names: https://kubernetes.io/docs/concepts/overview/working-with-objects/common-labels/
	LabelK8sName      = "app.kubernetes.io/name"       // The name of the application (string)
	LabelK8sInstance  = "app.kubernetes.io/instance"   // A unique name identifying the instance of an application (string)
	LabelK8sVersion   = "app.kubernetes.io/version"    // The current version of the application (e.g., a semantic version, revision hash, etc.) (string)
	LabelK8sComponent = "app.kubernetes.io/component"  // The component within the architecture (string)
	LabelK8sPartOf    = "app.kubernetes.io/part-of"    // The name of a higher level application this one is part of (string)
	LabelK8sManagedBy = "app.kubernetes.io/managed-by" // The tool being used to manage the operation of an application (string)
	LabelK8sCreatedBy = "app.kubernetes.io/created-by" // The controller/user who created this resource (string)

	// Commonly-used label names.
	LabelApp = "app"
)

var ReconcileForGlobalProxyList = []string{CollectorTrustedCAName}
var ExtraNoProxyList = []string{ElasticsearchFQDN}
