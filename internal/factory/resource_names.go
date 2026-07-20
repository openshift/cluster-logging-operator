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
	MetricsAuthClusterRoleBinding    string
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

// LegacyCollectorConfigMapName returns the collector ConfigMap name used before
// the clf- prefix was introduced (LOG-9591).
func LegacyCollectorConfigMapName(forwarderName string) string {
	return forwarderName + "-config"
}

// ResourceNames is a factory for naming of objects based on ClusterLogForwarder namespace and name
func ResourceNames(clf obsv1.ClusterLogForwarder) *ForwarderResourceNames {
	resBaseName := clf.Name
	return &ForwarderResourceNames{
		CommonName:                       resBaseName,
		SecretMetrics:                    resBaseName + "-metrics",
		// Prefix with "clf-" so the ConfigMap does not collide with LokiStack's "{name}-config"
		ConfigMap:                        fmt.Sprintf("clf-%s-config", resBaseName),
		MetadataReaderClusterRoleBinding: fmt.Sprintf("cluster-logging-%s-%s-metadata-reader", clf.Namespace, resBaseName),
		MetricsAuthClusterRoleBinding:    fmt.Sprintf("cluster-logging-%s-%s-metrics-auth", clf.Namespace, resBaseName),
		ForwarderName:                    clf.Name,
		CaTrustBundle:                    resBaseName + "-trustbundle",
		ServiceAccount:                   clf.Spec.ServiceAccount.Name,
		InternalLogStoreSecret:           clf.Spec.ServiceAccount.Name + "-default",
		ServiceAccountTokenSecret:        clf.Spec.ServiceAccount.Name + "-token",
		Secrets:                          resBaseName + "-secrets",
		AwsCredentialsFile:               resBaseName + "-" + constants.AwsCredentialsConfigMapName,
	}
}
