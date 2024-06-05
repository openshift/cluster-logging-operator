package security

import (
	"context"

	logging "github.com/openshift/cluster-logging-operator/api/logging/v1"

	"github.com/openshift/cluster-logging-operator/internal/constants"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/output/gcl"
	"github.com/openshift/cluster-logging-operator/internal/validations/clusterlogforwarder/conditions"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	client "sigs.k8s.io/controller-runtime/pkg/client"
)

func MakeOutputSecretMap(k8sClient client.Client, loggingClfOutputs []logging.OutputSpec, namespace string) (map[string]*corev1.Secret, *logging.NamedConditions, error) {
	var err error
	outputSecrets := make(map[string]*corev1.Secret, len(loggingClfOutputs))
	outputStatus := logging.NamedConditions{}
	for _, output := range loggingClfOutputs {
		if output.Secret == nil {
			continue
		}
		var secret *corev1.Secret
		secret, err = GetSecret(k8sClient, output.Secret.Name, namespace)
		if err != nil {
			if errors.IsNotFound(err) {
				outputStatus.Set(output.Name, conditions.CondMissing("secret: %q was not found", output.Secret.Name))
				continue
			}
			return nil, nil, err
		}
		outputSecrets[output.Name] = secret
	}

	if len(outputStatus) == 0 {
		return outputSecrets, nil, nil
	}

	return outputSecrets, &outputStatus, err
}

func GetSecret(k8sClient client.Client, secretName, namespace string) (*corev1.Secret, error) {
	outSecret := &corev1.Secret{}
	key := types.NamespacedName{Name: secretName, Namespace: namespace}
	if err := k8sClient.Get(context.TODO(), key, outSecret); err != nil {
		return nil, err
	}
	return outSecret, nil
}

func HasUsernamePassword(secret *corev1.Secret) bool {
	return HasKeys(secret, constants.ClientUsername, constants.ClientPassword)
}

func HasTLSCertAndKey(secret *corev1.Secret) bool {
	return HasKeys(secret, constants.ClientCertKey, constants.ClientPrivateKey)
}

func HasCABundle(secret *corev1.Secret) bool {
	return HasKeys(secret, constants.TrustedCABundleKey)
}

func HasSharedKey(secret *corev1.Secret) bool {
	return HasKeys(secret, constants.SharedKey)
}

func HasPassphrase(secret *corev1.Secret) bool {
	return HasKeys(secret, constants.Passphrase)
}

func HasBearerTokenFileKey(secret *corev1.Secret) bool {
	return HasKeys(secret, constants.BearerTokenFileKey)
}

func HasAwsRoleArnKey(secret *corev1.Secret) bool {
	return HasKeys(secret, constants.AWSWebIdentityRoleKey)
}

func HasAwsCredentialsKey(secret *corev1.Secret) bool {
	return HasKeys(secret, constants.AWSCredentialsKey)
}

func HasAwsAccessKeyId(secret *corev1.Secret) bool {
	return HasKeys(secret, constants.AWSAccessKeyID)
}

func HasAwsSecretAccessKey(secret *corev1.Secret) bool {
	return HasKeys(secret, constants.AWSSecretAccessKey)
}

func HasGoogleApplicationCredentialsKey(secret *corev1.Secret) bool {
	return HasKeys(secret, gcl.GoogleApplicationCredentialsKey)
}

func HasSplunkHecToken(secret *corev1.Secret) bool {
	return HasKeys(secret, constants.SplunkHECTokenKey)
}

func HasSASLMechanism(secret *corev1.Secret) bool {
	return HasKeys(secret, constants.SASLMechanisms)
}

// GetKey if found return value and ok=true, else ok=false
func GetKey(secret *corev1.Secret, key string) (data []byte, ok bool) {
	if secret == nil {
		return nil, false
	}
	data, ok = secret.Data[key]
	return data, ok
}

// HasKeys true if all keys are present.
func HasKeys(secret *corev1.Secret, keys ...string) bool {
	for _, k := range keys {
		_, ok := GetKey(secret, k)
		if !ok {
			return false
		}
	}
	return true
}

func GetFromSecret(secret *corev1.Secret, name string) string {
	if secret != nil {
		return string(secret.Data[name])
	}
	return ""
}
