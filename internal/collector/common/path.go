package common

import (
	"path/filepath"
	"strings"

	"github.com/openshift/cluster-logging-operator/internal/constants"
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

// OtelConfigMapMountPath is the spec.configmaps mountpath that makes the
// OpenTelemetry operator mount a configmap's entries at OtelConfigMapPath:
// the operator joins /var/conf, the mountpath and its own "configmap-<name>"
// volume directory, so an empty path segment resolves to just that directory
const OtelConfigMapMountPath = "."

// OtelConfigMapPath is the in-pod path of a configmap entry mounted through
// the OpenTelemetry operator's spec.configmaps. The volume directory name
// replicates the operator's naming.ConfigMapExtra (see
// otelConfigMapVolumeName). It is the single source of truth shared by the
// otelc config generation (which references these paths) and the
// OpenTelemetryCollector CR creation (which mounts the configmaps there)
func OtelConfigMapPath(name string, file string) string {
	return filepath.Join("/var/conf", OtelConfigMapVolumeName(name), file)
}

// OtelConfigMapVolumeName replicates the OpenTelemetry operator's
// naming.ConfigMapExtra: "configmap-<name>" truncated to 63 characters and
// sanitized to a DNS name, where any character that is not [a-z0-9] becomes
// '-' in the middle of the name or 'a' at its edges. Configmap names such as
// "openshift-service-ca.crt" therefore mount under
// "configmap-openshift-service-ca-crt"
func OtelConfigMapVolumeName(name string) string {
	volumeName := "configmap-" + name
	if len(volumeName) > 63 {
		volumeName = volumeName[:63]
	}
	var sanitized []rune
	lower := strings.ToLower(volumeName)
	for i, x := range lower {
		if (x >= 'a' && x <= 'z') || (x >= '0' && x <= '9') {
			sanitized = append(sanitized, x)
		} else if i == 0 || i == len(lower)-1 {
			sanitized = append(sanitized, 'a')
		} else {
			sanitized = append(sanitized, '-')
		}
	}
	return string(sanitized)
}
