package factory

import (
	"fmt"
	logging "github.com/openshift/cluster-logging-operator/apis/logging/v1"

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

func (f *ForwarderResourceNames) DaemonSetName() string {
	return f.CommonName
}

// GenerateResourceNames is a factory for naming of objects based on ClusterLogForwarder namespace and name
func GenerateResourceNames(clf logging.ClusterLogForwarder) *ForwarderResourceNames {
	resBaseName := clf.Name
	if clf.Namespace == constants.OpenshiftNS && clf.Name == constants.SingletonName {
		resBaseName = constants.CollectorName
	}

	forwarderResNames := &ForwarderResourceNames{
		CommonName:                       resBaseName,
		SecretMetrics:                    resBaseName + "-metrics",
		ConfigMap:                        resBaseName + "-config",
		MetadataReaderClusterRoleBinding: fmt.Sprintf("cluster-logging-%s-%s-metadata-reader", clf.Namespace, resBaseName),
	}

	if clf.Namespace == constants.OpenshiftNS && clf.Name == constants.SingletonName {
		forwarderResNames.CaTrustBundle = constants.CollectorTrustedCAName
		forwarderResNames.ServiceAccount = constants.CollectorServiceAccountName
		forwarderResNames.InternalLogStoreSecret = resBaseName
		forwarderResNames.ServiceAccountTokenSecret = constants.LogCollectorToken
	} else {
		forwarderResNames.CaTrustBundle = resBaseName + "-trustbundle"
		forwarderResNames.ServiceAccount = clf.Spec.ServiceAccountName
		forwarderResNames.InternalLogStoreSecret = clf.Spec.ServiceAccountName + "-default"
		forwarderResNames.ServiceAccountTokenSecret = clf.Spec.ServiceAccountName + "-token"
	}

	return forwarderResNames
}
