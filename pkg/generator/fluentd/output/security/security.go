package security

import (
	"fmt"
	"path/filepath"

	"github.com/openshift/cluster-logging-operator/pkg/constants"
	corev1 "k8s.io/api/core/v1"
)

type TLS bool

type TLSCertKey struct {
	CertPath string
	KeyPath  string
}

type UserNamePass struct {
	UsernamePath string
	PasswordPath string
}

type SharedKey struct {
	Key string
}

type CAFile struct {
	CAFilePath string
}

type Passphrase struct {
	PassphrasePath string
}

var NoSecrets = map[string]*corev1.Secret{}

func HasUsernamePassword(secret *corev1.Secret) bool {
	if secret == nil {
		return false
	}

	if _, ok := secret.Data[constants.ClientUsername]; !ok {
		return false
	}
	if _, ok := secret.Data[constants.ClientPassword]; !ok {
		return false
	}
	return true
}

func HasTLSCertAndKey(secret *corev1.Secret) bool {
	if secret == nil {
		return false
	}

	if _, ok := secret.Data[constants.ClientCertKey]; !ok {
		return false
	}
	if _, ok := secret.Data[constants.ClientPrivateKey]; !ok {
		return false
	}
	return true
}

func HasCABundle(secret *corev1.Secret) bool {
	if secret == nil {
		return false
	}

	if _, ok := secret.Data[constants.TrustedCABundleKey]; !ok {
		return false
	}
	return true
}

func HasSharedKey(secret *corev1.Secret) bool {
	if secret == nil {
		return false
	}

	if _, ok := secret.Data[constants.SharedKey]; !ok {
		return false
	}
	return true
}

func HasPassphrase(secret *corev1.Secret) bool {
	if secret == nil {
		return false
	}

	if _, ok := secret.Data[constants.Passphrase]; !ok {
		return false
	}
	return true
}

func SecretPath(name string, file string) string {
	return fmt.Sprintf("'%s'", filepath.Join("/var/run/ocp-collector/secrets", name, file))
}

func GetFromSecret(secret *corev1.Secret, name string) string {
	if secret != nil {
		return string(secret.Data[name])
	}
	return ""
}
