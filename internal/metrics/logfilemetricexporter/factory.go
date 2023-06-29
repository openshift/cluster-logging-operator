package logfilemetricexporter

import (
	"strings"

	"github.com/openshift/cluster-logging-operator/internal/runtime"

	"github.com/openshift/cluster-logging-operator/internal/tls"
	apps "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"

	configv1 "github.com/openshift/api/config/v1"
	loggingv1 "github.com/openshift/cluster-logging-operator/apis/logging/v1"
	loggingv1a1 "github.com/openshift/cluster-logging-operator/apis/logging/v1alpha1"
	"github.com/openshift/cluster-logging-operator/internal/collector"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	coreFactory "github.com/openshift/cluster-logging-operator/internal/factory"
	"github.com/openshift/cluster-logging-operator/internal/utils"
)

const (
	clusterLoggingPriorityClassName = "system-node-critical"
	ExporterPort                    = int32(2112)
	ExporterPortName                = "exporter-port"
	logContainers                   = "varlogcontainers"
	logContainersValue              = "/var/log/containers"
	exporterMetricsVolumeName       = "lfme-metrics"
	ExporterMetricsSecretName       = "lfme-secret"
	logPods                         = "varlogpods"
	logPodsValue                    = "/var/log/pods"
	metricsVolumePath               = "/etc/logfilemetricexporter/metrics"
)

// resourceRequirements returns the resource requirements for a given metric-exporter implementation
// or it's default if none are specified
func resourceRequirements(exporter loggingv1a1.LogFileMetricExporter) v1.ResourceRequirements {
	if exporter.Spec.Resources == nil {
		return v1.ResourceRequirements{}
	}
	return *exporter.Spec.Resources
}

func nodeSelector(exporter loggingv1a1.LogFileMetricExporter) map[string]string {
	return exporter.Spec.NodeSelector
}

func tolerations(exporter loggingv1a1.LogFileMetricExporter) []v1.Toleration {
	if exporter.Spec.Tolerations == nil {
		return constants.DefaultTolerations()
	}

	// Add default tolerations if tolerations spec'd
	// Spec'd tolerations take precedence
	finalTolerations := make([]v1.Toleration, len(exporter.Spec.Tolerations))
	copy(finalTolerations, exporter.Spec.Tolerations)

	tolerationMap := make(map[string]bool)
	for _, tol := range exporter.Spec.Tolerations {
		tolerationMap[tol.Key] = true
	}

	for _, defaultTol := range constants.DefaultTolerations() {
		if exists := tolerationMap[defaultTol.Key]; !exists {
			finalTolerations = append(finalTolerations, defaultTol)
		}
	}

	return finalTolerations
}

func NewDaemonSet(exporter loggingv1a1.LogFileMetricExporter, namespace, name string, collectionType loggingv1.LogCollectionType, tlsProfileSpec configv1.TLSProfileSpec, visitors ...func(o runtime.Object)) *apps.DaemonSet {
	podSpec := NewPodSpec(exporter, tlsProfileSpec)
	ds := coreFactory.NewDaemonSet(name, namespace, constants.LogfilesmetricexporterName, constants.LogfilesmetricexporterName, string(collectionType), *podSpec, visitors...)
	return ds
}

func NewPodSpec(exporter loggingv1a1.LogFileMetricExporter, tlsProfileSpec configv1.TLSProfileSpec) *v1.PodSpec {

	podSpec := &v1.PodSpec{
		NodeSelector:                  utils.EnsureLinuxNodeSelector(nodeSelector(exporter)),
		PriorityClassName:             clusterLoggingPriorityClassName,
		ServiceAccountName:            constants.CollectorServiceAccountName,
		TerminationGracePeriodSeconds: utils.GetInt64(10),
		Tolerations:                   tolerations(exporter),
		Volumes: []v1.Volume{
			{Name: logContainers, VolumeSource: v1.VolumeSource{HostPath: &v1.HostPathVolumeSource{Path: logContainersValue}}},
			{Name: logPods, VolumeSource: v1.VolumeSource{HostPath: &v1.HostPathVolumeSource{Path: logPodsValue}}},
			{Name: exporterMetricsVolumeName, VolumeSource: v1.VolumeSource{Secret: &v1.SecretVolumeSource{SecretName: ExporterMetricsSecretName}}},
		},
	}

	exporterContainer := newLogMetricsExporterContainer(exporter, tlsProfileSpec)
	podSpec.Containers = []v1.Container{
		*exporterContainer,
	}

	return podSpec
}

func newLogMetricsExporterContainer(exporter loggingv1a1.LogFileMetricExporter, tlsProfileSpec configv1.TLSProfileSpec) *v1.Container {
	exporterContainer := coreFactory.NewContainer(constants.LogfilesmetricexporterName,
		constants.LogfilesmetricexporterName,
		v1.PullIfNotPresent, resourceRequirements(exporter))

	exporterContainer.Ports = []v1.ContainerPort{
		{
			Name:          ExporterPortName,
			ContainerPort: ExporterPort,
			Protocol:      v1.ProtocolTCP,
		},
	}
	exporterContainer.Command = []string{"/bin/bash"}
	exporterContainer.Args = []string{"-c",
		"/usr/local/bin/log-file-metric-exporter -verbosity=2 -dir=/var/log/pods -http=:2112 -keyFile=/etc/logfilemetricexporter/metrics/tls.key -crtFile=/etc/logfilemetricexporter/metrics/tls.crt -tlsMinVersion=" +
			tls.MinTLSVersion(tlsProfileSpec) + " -cipherSuites=" + strings.Join(tls.TLSCiphers(tlsProfileSpec), ",")}

	exporterContainer.VolumeMounts = []v1.VolumeMount{
		{Name: logContainers, ReadOnly: true, MountPath: logContainersValue},
		{Name: logPods, ReadOnly: true, MountPath: logPodsValue},
		{Name: exporterMetricsVolumeName, ReadOnly: true, MountPath: metricsVolumePath},
	}

	collector.AddSecurityContextTo(&exporterContainer)
	return &exporterContainer
}
