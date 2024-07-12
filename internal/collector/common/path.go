package common

import (
	"github.com/openshift/cluster-logging-operator/internal/constants"
	"path/filepath"
)

// SecretPath is the path for any secret visible to the collector
func SecretPath(secretName string, file string) string {
	return filepath.Join(constants.CollectorSecretsDir, secretName, file)
}

// SecretBasePath is the path for any secret visible to the collector
func SecretBasePath(secretName string) string {
	return filepath.Join(constants.CollectorSecretsDir, secretName)
}

// ConfigMapPath is the path for any configmap visible to the collector
func ConfigMapPath(name string, file string) string {
	return filepath.Join(constants.ConfigMapBaseDir, name, file)
}

// ConfigMapBasePath is the path for any configmap visible to the collector
func ConfigMapBasePath(name string) string {
	return filepath.Join(constants.ConfigMapBaseDir, name)
}

// ServiceAccountBasePath is the base path for any serviceaccount token projection visible to the collector
func ServiceAccountBasePath(name string) string {
	return filepath.Join(constants.ServiceAccountSecretPath, name)
}
