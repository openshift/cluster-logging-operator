package constants

const (
	// Keys used in ClusterLogForwarder.Output Secrets keys.
	// Documented with OutputSpec.Secret in /apis/logging/v1/cluster_log_forwarder_types.go
	//
	// WARNING: changing or removing values here is a breaking API change.

	// TLS keys, used by any output that supports TLS.

	ClientCertKey      = "tls.crt"
	ClientPrivateKey   = "tls.key"
	TrustedCABundleKey = "ca-bundle.crt"
	Passphrase         = "passphrase"

	// Username/Password keys, used by any output with username/password authentication.

	ClientUsername = "username"
	ClientPassword = "password"

	// SASL keys, used by any output that supports SASL.

	SASLEnable        = "sasl.enable"
	SASLMechanisms    = "sasl.mechanisms"
	SASLAllowInsecure = "sasl.allow-insecure"

	// Output-specific keys

	SharedKey             = "shared_key"            // fluent forward
	DeprecatedSaslOverSSL = "sasl_over_ssl"         // Kafka
	AWSSecretAccessKey    = "aws_secret_access_key" //nolint:gosec
	AWSAccessKeyID        = "aws_access_key_id"
)
const (
	SingletonName = "instance"
	OpenshiftNS   = "openshift-logging"
	// global proxy / trusted ca bundle consts
	ProxyName = "cluster"

	InjectTrustedCABundleLabel = "config.openshift.io/inject-trusted-cabundle"
	TrustedCABundleMountFile   = "tls-ca-bundle.pem"
	TrustedCABundleMountDir    = "/etc/pki/ca-trust/extracted/pem/"
	TrustedCABundleHashName    = "logging.openshift.io/hash"
	SecretHashPrefix           = "logging.openshift.io/"
	KibanaTrustedCAName        = "kibana-trusted-ca-bundle"
	// internal elasticsearch FQDN to prevent to connect to the global proxy
	ElasticsearchFQDN          = "elasticsearch.openshift-logging.svc"
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

	FluentdImageEnvVar            = "FLUENTD_IMAGE"
	VectorImageEnvVar             = "VECTOR_IMAGE"
	LogfilesmetricImageEnvVar     = "LOGFILEMETRICEXPORTER_IMAGE"
	CertEventName                 = "cluster-logging-certs-generate"
	ClusterInfrastructureInstance = "cluster"

	ContainerLogDir = "/var/log/containers"
	PodLogDir       = "/var/log/pods"
)

var ReconcileForGlobalProxyList = []string{CollectorTrustedCAName}
