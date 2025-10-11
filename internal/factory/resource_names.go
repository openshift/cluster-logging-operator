package factory

import (
	"fmt"

	obsv1 "github.com/openshift/cluster-logging-operator/api/observability/v1"
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
	ForwarderName                    string
	Secrets                          string
	AwsCredentialsFile               string
}

func (f *ForwarderResourceNames) DaemonSetName() string {
	return f.CommonName
}

// GenerateInputServiceName addresses HTTP input service name uniqueness by concatenating the common name with the input service name
func (f *ForwarderResourceNames) GenerateInputServiceName(serviceName string) string {
	return fmt.Sprintf("%s-%s", f.CommonName, serviceName)
}

// ResourceNames is a factory for naming of objects based on ClusterLogForwarder namespace and name
func ResourceNames(clf obsv1.ClusterLogForwarder) *ForwarderResourceNames {
	resBaseName := clf.Name
	return &ForwarderResourceNames{
		CommonName:                       resBaseName,
		SecretMetrics:                    resBaseName + "-metrics",
		ConfigMap:                        resBaseName + "-config",
		MetadataReaderClusterRoleBinding: fmt.Sprintf("cluster-logging-%s-%s-metadata-reader", clf.Namespace, resBaseName),
		ForwarderName:                    clf.Name,
		CaTrustBundle:                    resBaseName + "-trustbundle",
		ServiceAccount:                   clf.Spec.ServiceAccount.Name,
		InternalLogStoreSecret:           clf.Spec.ServiceAccount.Name + "-default",
		ServiceAccountTokenSecret:        clf.Spec.ServiceAccount.Name + "-token",
		Secrets:                          resBaseName + "-secrets",
		AwsCredentialsFile:               resBaseName + "-" + constants.AwsCredentialsConfigMapName,
	}
}
