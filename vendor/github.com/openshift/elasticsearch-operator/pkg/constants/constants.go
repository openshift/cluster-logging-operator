package constants

import "github.com/openshift/elasticsearch-operator/pkg/utils"

const (
	ProxyName                  = "cluster"
	TrustedCABundleKey         = "ca-bundle.crt"
	TrustedCABundleMountDir    = "/etc/pki/ca-trust/extracted/pem/"
	TrustedCABundleMountFile   = "tls-ca-bundle.pem"
	InjectTrustedCABundleLabel = "config.openshift.io/inject-trusted-cabundle"
	TrustedCABundleHashName    = "logging.openshift.io/hash"
	KibanaTrustedCAName        = "kibana-trusted-ca-bundle"
	SecretHashPrefix           = "logging.openshift.io/"
	KibanaInstanceName         = "kibana"
	ElasticsearchDefaultImage  = "quay.io/openshift/origin-logging-elasticsearch6"
)

var (
	ReconcileForGlobalProxyList = []string{KibanaTrustedCAName}
	packagedElasticsearchImage  = utils.LookupEnvWithDefault("ELASTICSEARCH_IMAGE", ElasticsearchDefaultImage)
)

func PackagedElasticsearchImage() string {
	return packagedElasticsearchImage
}
