package common

import (
	logging "github.com/openshift/cluster-logging-operator/api/logging/v1"
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	"github.com/openshift/cluster-logging-operator/internal/generator/helpers/security"
	corev1 "k8s.io/api/core/v1"
)

func MapOutputTls(loggingTls *logging.OutputTLSSpec, outputSecret *corev1.Secret) *obs.OutputTLSSpec {
	if loggingTls == nil && !security.HasTLSCertAndKey(outputSecret) &&
		!security.HasCABundle(outputSecret) && !security.HasPassphrase(outputSecret) {
		return nil
	}

	obsTls := &obs.OutputTLSSpec{}

	if loggingTls != nil {
		obsTls.InsecureSkipVerify = loggingTls.InsecureSkipVerify
		obsTls.TLSSecurityProfile = loggingTls.TLSSecurityProfile
	}

	if security.HasTLSCertAndKey(outputSecret) {
		obsTls.Certificate = &obs.ValueReference{
			Key:        constants.ClientCertKey,
			SecretName: outputSecret.Name,
		}
		obsTls.Key = &obs.SecretReference{
			Key:        constants.ClientPrivateKey,
			SecretName: outputSecret.Name,
		}
	}
	if security.HasCABundle(outputSecret) {
		obsTls.CA = &obs.ValueReference{
			Key:        constants.TrustedCABundleKey,
			SecretName: outputSecret.Name,
		}
	}
	if security.HasPassphrase(outputSecret) {
		obsTls.KeyPassphrase = &obs.SecretReference{
			Key:        constants.Passphrase,
			SecretName: outputSecret.Name,
		}
	}

	return obsTls
}
