package constants

const (
	SingletonName = "instance"
	OpenshiftNS   = "openshift-logging"
	// global proxy / trusted ca bundle consts
	ProxyName                  = "cluster"
	TrustedCABundleKey         = "ca-bundle.crt"
	InjectTrustedCABundleLabel = "config.openshift.io/inject-trusted-cabundle"
	TrustedCABundleMountFile   = "tls-ca-bundle.pem"
	TrustedCABundleMountDir    = "/etc/pki/ca-trust/extracted/pem/"
	TrustedCABundleHashName    = "logging.openshift.io/hash"
	SecretHashPrefix           = "logging.openshift.io/"
	FluentdTrustedCAName       = "fluentd-trusted-ca-bundle"
	KibanaTrustedCAName        = "kibana-trusted-ca-bundle"
	// internal elasticsearch FQDN to prevent to connect to the global proxy
	ElasticsearchFQDN   = "elasticsearch.openshift-logging.svc.cluster.local"
	ElasticsearchName   = "elasticsearch"
	ElasticsearchPort   = "9200"
	FluentdName         = "fluentd"
	KibanaName          = "kibana"
	KibanaProxyName     = "kibana-proxy"
	CuratorName         = "curator"
	LogStoreURL         = "https://" + ElasticsearchFQDN + ":" + ElasticsearchPort
	MasterCASecretName  = "master-certs"
	CollectorSecretName = "fluentd"

	/* #nosec - suppressing rule G101 */
	KibanaSessionSecretName = "kibana-session-secret"

	FluentdImageEnvVar = "FLUENTD_IMAGE"
)

var ReconcileForGlobalProxyList = []string{FluentdTrustedCAName}
