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
	ElasticsearchDefaultImage   = "quay.io/openshift-logging/elasticsearch6"
	ProxyDefaultImage           = "quay.io/openshift-logging/elasticsearch-proxy:latest"
	CuratorDefaultImage         = "quay.io/openshift-logging/curator5"
	TheoreticalShardMaxSizeInMB = 40960

	// OcpTemplatePrefix is the prefix all operator generated templates
	OcpTemplatePrefix = "ocp-gen"

	SecurityIndex = ".security"

	EOCertManagementLabel = "logging.openshift.io/elasticsearch-cert-management"
	EOComponentCertPrefix = "logging.openshift.io/elasticsearch-cert."
)

var (
	ReconcileForGlobalProxyList = []string{KibanaTrustedCAName}
	packagedCuratorImage        = utils.LookupEnvWithDefault("CURATOR_IMAGE", CuratorDefaultImage)
	ExpectedSecretKeys          = []string{
		"admin-ca",
		"admin-cert",
		"admin-key",
		"elasticsearch.crt",
		"elasticsearch.key",
		"logging-es.crt",
		"logging-es.key",
	}
)

func PackagedCuratorImage() string {
	return packagedCuratorImage
}
