package factory

import (
	"fmt"

	"github.com/openshift/cluster-logging-operator/internal/constants"
)

type ForwarderResourceNames struct {
	CommonName                       string
	SecretMetrics                    string
	ConfigMap                        string
	MetadataReaderClusterRoleBinding string
	CaTrustBundle                    string
	ServiceAccount                   string
	InternalLogStoreSecret           string
	ServiceAccountTokenSecret        string
}

func GenerateResourceNames(name, namespace string) *ForwarderResourceNames {
	resBaseName := name
	if name == constants.SingletonName {
		resBaseName = constants.CollectorName
	}

	forwarderResNames := &ForwarderResourceNames{
		CommonName:                       resBaseName,
		SecretMetrics:                    resBaseName + "-metrics",
		ConfigMap:                        resBaseName + "-config",
		MetadataReaderClusterRoleBinding: fmt.Sprintf("cluster-logging-%s-%s-metadata-reader", namespace, resBaseName),
	}

	if name == constants.SingletonName {
		forwarderResNames.CaTrustBundle = constants.CollectorTrustedCAName
		forwarderResNames.ServiceAccount = constants.CollectorServiceAccountName
		forwarderResNames.InternalLogStoreSecret = resBaseName
		forwarderResNames.ServiceAccountTokenSecret = constants.LogCollectorToken
	} else {
		forwarderResNames.CaTrustBundle = resBaseName + "-trustbundle"
		forwarderResNames.ServiceAccount = resBaseName
		forwarderResNames.InternalLogStoreSecret = resBaseName + "-default"
		forwarderResNames.ServiceAccountTokenSecret = resBaseName + "-token"
	}

	return forwarderResNames
}
