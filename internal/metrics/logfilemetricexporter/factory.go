package logfilemetricexporter

import (
	"strings"

	"github.com/openshift/cluster-logging-operator/internal/runtime"

	"github.com/openshift/cluster-logging-operator/internal/tls"
	apps "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	configv1 "github.com/openshift/api/config/v1"
	loggingv1a1 "github.com/openshift/cluster-logging-operator/api/logging/v1alpha1"
	"github.com/openshift/cluster-logging-operator/internal/auth"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	coreFactory "github.com/openshift/cluster-logging-operator/internal/factory"
	"github.com/openshift/cluster-logging-operator/internal/utils"
)

const (
	clusterLoggingPriorityClassName = "system-node-critical"
	exporterPort                    = int32(2112)
	logContainers                   = "varlogcontainers"
	logContainersValue              = "/var/log/containers"
	exporterMetricsVolumeName       = "lfme-metrics"
	ExporterMetricsSecretName       = "lfme-secret"
	logPods                         = "varlogpods"
	logPodsValue                    = "/var/log/pods"
	metricsVolumePath               = "/etc/logfilemetricexporter/metrics"

	// lfmeRunAsUser is the fixed non-root UID the exporter runs as. The exporter reads the
	// hostPath log directories via group 0 (the default GID granted by OpenShift), which
	// satisfies the 0750 root:root permissions on /var/log/pods without joining extra groups.
	lfmeRunAsUser int64 = 1000
	// selinuxTypeLogWriter (container_logwriter_t) is an MCS-constrained container domain that
	// grants read plus the inotify "watch"/"watch_reads" permissions on container_log_t, which
	// the exporter requires to watch /var/log/pods. It is far more restrictive than the
	// super-privileged spc_t; the otherwise-preferable container_logreader_t domain is not
	// usable because it denies the inotify "watch" permission.
	selinuxTypeLogWriter = "container_logwriter_t"
)

var (
	lfmeMaxUnavailable = intstr.Parse("100%")
)

// resourceRequirements returns the resource requirements for a given metric-exporter implementation
// or it's default if none are specified
func resourceRequirements(exporter loggingv1a1.LogFileMetricExporter) *v1.ResourceRequirements {
	if exporter.Spec.Resources == nil {
		return &v1.ResourceRequirements{}
	}
	return exporter.Spec.Resources
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

func NewDaemonSet(exporter loggingv1a1.LogFileMetricExporter, namespace, name string, tlsProfileSpec configv1.TLSProfileSpec, visitors ...func(o runtime.Object)) *apps.DaemonSet {
	podSpec := NewPodSpec(exporter, tlsProfileSpec)
	ds := coreFactory.NewDaemonSet(namespace, name, exporter.Name, constants.LogfilesmetricexporterName, constants.LogfilesmetricexporterName, lfmeMaxUnavailable, *podSpec, visitors...)
	return ds
}

func NewPodSpec(exporter loggingv1a1.LogFileMetricExporter, tlsProfileSpec configv1.TLSProfileSpec) *v1.PodSpec {

	podSpec := &v1.PodSpec{
		NodeSelector:                  utils.EnsureLinuxNodeSelector(nodeSelector(exporter)),
		PriorityClassName:             clusterLoggingPriorityClassName,
		ServiceAccountName:            constants.LogfilesmetricexporterName,
		TerminationGracePeriodSeconds: utils.GetPtr[int64](10),
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
	exporterContainer := runtime.NewContainer(constants.LogfilesmetricexporterName,
		utils.GetComponentImage(constants.LogfilesmetricexporterName),
		v1.PullIfNotPresent, resourceRequirements(exporter))

	exporterContainer.Ports = []v1.ContainerPort{
		{
			Name:          constants.MetricsPortName,
			ContainerPort: constants.LogfilesmetricexporterPort,
			Protocol:      v1.ProtocolTCP,
		},
	}
	exporterContainer.Command = []string{"/bin/bash"}
	exporterContainer.Args = []string{"-c",
		"/usr/local/bin/log-file-metric-exporter -verbosity=2 -dir=/var/log/pods -http=:2112 -keyFile=/etc/logfilemetricexporter/metrics/tls.key -crtFile=/etc/logfilemetricexporter/metrics/tls.crt -secureMetrics -tlsMinVersion=" +
			tls.MinTLSVersion(tlsProfileSpec) + " -cipherSuites=" + strings.Join(tls.TLSCiphers(tlsProfileSpec), ",") +
			" -groups=" + strings.Join(tls.TLSGroups(tlsProfileSpec), ",")}

	exporterContainer.VolumeMounts = []v1.VolumeMount{
		{Name: logContainers, ReadOnly: true, MountPath: logContainersValue},
		{Name: logPods, ReadOnly: true, MountPath: logPodsValue},
		{Name: exporterMetricsVolumeName, ReadOnly: true, MountPath: metricsVolumePath},
	}

	exporterContainer.SecurityContext = securityContext()
	return exporterContainer
}

// securityContext returns the minimal security context required by the log-file-metric-exporter.
// The exporter runs as a fixed non-root UID with all capabilities dropped, a read-only root
// filesystem, no privilege escalation and the default seccomp profile. It runs under the
// MCS-constrained container_logwriter_t SELinux domain, which grants read plus the inotify
// watch the exporter needs on the host log tree while remaining far more restrictive than spc_t.
func securityContext() *v1.SecurityContext {
	return &v1.SecurityContext{
		Capabilities: &v1.Capabilities{
			Drop: auth.RequiredDropCapabilities,
		},
		SELinuxOptions: &v1.SELinuxOptions{
			Type: selinuxTypeLogWriter,
		},
		RunAsUser:                utils.GetPtr(lfmeRunAsUser),
		RunAsNonRoot:             utils.GetPtr(true),
		ReadOnlyRootFilesystem:   utils.GetPtr(true),
		AllowPrivilegeEscalation: utils.GetPtr(false),
		SeccompProfile: &v1.SeccompProfile{
			Type: v1.SeccompProfileTypeRuntimeDefault,
		},
	}
}
