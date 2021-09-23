package constants

import "github.com/openshift/elasticsearch-operator/internal/utils"

const (
	ProxyName                   = "cluster"
	TrustedCABundleKey          = "ca-bundle.crt"
	TrustedCABundleMountDir     = "/etc/pki/ca-trust/extracted/pem/"
	TrustedCABundleMountFile    = "tls-ca-bundle.pem"
	InjectTrustedCABundleLabel  = "config.openshift.io/inject-trusted-cabundle"
	TrustedCABundleHashName     = "logging.openshift.io/hash"
	KibanaTrustedCAName         = "kibana-trusted-ca-bundle"
	SecretHashPrefix            = "logging.openshift.io/"
	ElasticsearchDefaultImage   = "quay.io/openshift/origin-logging-elasticsearch6"
	ProxyDefaultImage           = "quay.io/openshift/origin-elasticsearch-proxy:latest"
	TheoreticalShardMaxSizeInMB = 40960

	// OcpTemplatePrefix is the prefix all operator generated templates
	OcpTemplatePrefix = "ocp-gen"
)

var (
	ReconcileForGlobalProxyList = []string{KibanaTrustedCAName}
	packagedElasticsearchImage  = utils.LookupEnvWithDefault("ELASTICSEARCH_IMAGE", ElasticsearchDefaultImage)
	ExpectedSecretKeys          = []string{
		"admin-ca",
		"admin-cert",
		"admin-key",
		"elasticsearch.crt",
		"elasticsearch.key",
		"logging-es.crt",
		"logging-es.key"}
)

func PackagedElasticsearchImage() string {
	return packagedElasticsearchImage
}
