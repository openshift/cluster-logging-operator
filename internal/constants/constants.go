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
	AWSWebIdentityRoleKey       = "role_arn"        // key to expect for sts-formatted secret
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
)

const (
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
	CertEventName                 = "cluster-logging-certs-generate"
	ClusterInfrastructureInstance = "cluster"

	ContainerLogDir = "/var/log/containers"
	PodLogDir       = "/var/log/pods"
)

var ReconcileForGlobalProxyList = []string{CollectorTrustedCAName}
var ExtraNoProxyList = []string{ElasticsearchFQDN}
