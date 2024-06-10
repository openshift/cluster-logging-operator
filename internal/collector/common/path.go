package common

import (
	"path/filepath"
)

const (
	basePath = "/var/run/ocp-collector"
)

// SecretPath is the path for any secret visible to the collector
func SecretPath(secretName string, file string) string {
	return filepath.Join(basePath, "secrets", secretName, file)
}

// SecretBasePath is the path for any secret visible to the collector
func SecretBasePath(secretName string) string {
	return filepath.Join(basePath, "secrets", secretName)
}

// ConfigMapPath is the path for any configmap visible to the collector
func ConfigMapPath(name string, file string) string {
	return filepath.Join(basePath, "config", name, file)
}

// ConfigMapBasePath is the path for any configmap visible to the collector
func ConfigMapBasePath(name string) string {
	return filepath.Join(basePath, "config", name)
}

// ServiceAccountBasePath is the base path for any serviceaccount token projection visible to the collector
func ServiceAccountBasePath(name string) string {
	return filepath.Join(basePath, "serviceaccount", name)
}
