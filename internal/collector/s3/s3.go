package s3

import (
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	"github.com/openshift/cluster-logging-operator/internal/api/observability"
	"github.com/openshift/cluster-logging-operator/internal/collector/aws"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// ReconcileAWSCredentialsConfigMap reconciles a configmap with credential profile(s) for S3 output(s).
// This function delegates to the shared AWS credentials module.
func ReconcileAWSCredentialsConfigMap(k8sClient client.Client, reader client.Reader, namespace, name string, outputs []obs.OutputSpec, secrets observability.Secrets, configMaps map[string]*corev1.ConfigMap, owner metav1.OwnerReference) (*corev1.ConfigMap, error) {
	return aws.ReconcileAWSCredentialsConfigMap(k8sClient, reader, namespace, name, outputs, secrets, configMaps, owner)
}

// GenerateS3CredentialProfiles generates AWS CLI profiles for S3 outputs.
// This function delegates to the shared AWS credentials module.
func GenerateS3CredentialProfiles(outputs []obs.OutputSpec, secrets observability.Secrets) (string, error) {
	return aws.GenerateAWSCredentialProfiles(outputs, secrets)
}