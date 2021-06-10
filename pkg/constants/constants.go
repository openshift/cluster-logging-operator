package constants

const (
	SingletonName = "instance"
	OpenshiftNS   = "openshift-logging"
	// global proxy / trusted ca bundle consts
	ProxyName                  = "cluster"
	TrustedCABundleKey         = "ca-bundle.crt"
	AWSSecretAccessKey         = "aws_secret_access_key" //nolint:gosec
	AWSAccessKeyID             = "aws_access_key_id"
	ClientCertKey              = "tls.crt"
	ClientPrivateKey           = "tls.key"
	ClientUsername             = "username"
	ClientPassword             = "password"
	InjectTrustedCABundleLabel = "config.openshift.io/inject-trusted-cabundle"
	TrustedCABundleMountFile   = "tls-ca-bundle.pem"
	TrustedCABundleMountDir    = "/etc/pki/ca-trust/extracted/pem/"
	TrustedCABundleHashName    = "logging.openshift.io/hash"
	SecretHashPrefix           = "logging.openshift.io/"
	FluentdTrustedCAName       = "fluentd-trusted-ca-bundle"
	KibanaTrustedCAName        = "kibana-trusted-ca-bundle"
	// internal elasticsearch FQDN to prevent to connect to the global proxy
	ElasticsearchFQDN   = "elasticsearch.openshift-logging.svc"
	ElasticsearchName   = "elasticsearch"
	ElasticsearchPort   = "9200"
	FluentdName         = "fluentd"
	KibanaName          = "kibana"
	KibanaProxyName     = "kibana-proxy"
	CuratorName         = "curator"
	LogStoreURL         = "https://" + ElasticsearchFQDN + ":" + ElasticsearchPort
	MasterCASecretName  = "master-certs"
	CollectorSecretName = "fluentd"

	// Disable gosec linter, complains "possible hard-coded secret"
	CollectorSecretsDir     = "/var/run/ocp-collector/secrets" //nolint:gosec
	KibanaSessionSecretName = "kibana-session-secret"          //nolint:gosec

	LegacySecureforward = "_LEGACY_SECUREFORWARD"
	LegacySyslog        = "_LEGACY_SYSLOG"

	FluentdImageEnvVar = "FLUENTD_IMAGE"
	CertEventName      = "cluster-logging-certs-generate"

	ClusterInfrastructureInstance = "cluster"
)

var ReconcileForGlobalProxyList = []string{FluentdTrustedCAName}
