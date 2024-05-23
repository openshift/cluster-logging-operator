package common

import (
	"fmt"
	"path/filepath"
)

// SecretPath is the path for any secret visible to the collector
func SecretPath(secretName string, file string) string {
	return fmt.Sprintf("%q", filepath.Join("/var/run/ocp-collector/secrets", secretName, file))
}

// SecretBasePath is the path for any secret visible to the collector
func SecretBasePath(secretName string) string {
	return fmt.Sprintf("%q", filepath.Join("/var/run/ocp-collector/secrets", secretName))
}

// ConfigmapPath is the path for any configmap visible to the collector
func ConfigmapPath(name string, file string) string {
	return fmt.Sprintf("%q", filepath.Join("/var/run/ocp-collector/config", name, file))
}

// ConfigmapBasePath is the path for any configmap visible to the collector
func ConfigmapBasePath(name string) string {
	return fmt.Sprintf("%q", filepath.Join("/var/run/ocp-collector/config", name))
}
