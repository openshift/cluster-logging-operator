package fluentbit

import (
	"fmt"

	"github.com/ViaQ/logerr/log"
	monitoringv1 "github.com/coreos/prometheus-operator/pkg/apis/monitoring/v1"
	logging "github.com/openshift/cluster-logging-operator/pkg/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/pkg/constants"
	"github.com/openshift/cluster-logging-operator/pkg/factory"
	"github.com/openshift/cluster-logging-operator/pkg/generators/forwarding/fluentbit"
	"github.com/openshift/cluster-logging-operator/pkg/utils"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

//VisitPodSpec allows reuse of exising functions of the original deployment
func VisitPodSpec(podSpec *corev1.PodSpec) *corev1.PodSpec {
	log.V(2).Info("Processing PodSpec to add fluentbit...")
	resourceRequirements := corev1.ResourceRequirements{}
	container := factory.NewContainer(constants.CollectorName, constants.FluentBitName, corev1.PullIfNotPresent, resourceRequirements)
	container.SecurityContext = &corev1.SecurityContext{
		Privileged: utils.GetBool(true),
	}
	container.Env = []corev1.EnvVar{
		{Name: constants.PodIPEnvVar, ValueFrom: &corev1.EnvVarSource{FieldRef: &corev1.ObjectFieldSelector{APIVersion: "v1", FieldPath: "status.podIP"}}},
	}
	container.VolumeMounts = []corev1.VolumeMount{
		{Name: constants.CollectorConfigVolumeName, ReadOnly: true, MountPath: "/etc/fluent-bit"},
		{Name: constants.LogVolumeName, ReadOnly: true, MountPath: constants.VarLogVolumePath},
		{Name: constants.VarLibVolumeName, ReadOnly: false, MountPath: constants.VarLibVolumePath},
	}
	container.Ports = []corev1.ContainerPort{
		{Name: constants.CollectorMetricsPortName, ContainerPort: constants.CollectorMetricsPort, Protocol: corev1.ProtocolTCP},
	}
	podSpec.Containers = append(podSpec.Containers, container)
	podSpec.Volumes = append(podSpec.Volumes,
		corev1.Volume{Name: constants.VarLibVolumeName, VolumeSource: corev1.VolumeSource{HostPath: &corev1.HostPathVolumeSource{Path: constants.VarLibVolumePath}}},
	)
	return podSpec
}

func VisitConfigMap(cm *corev1.ConfigMap, clfSpec logging.ClusterLogForwarderSpec) (*corev1.ConfigMap, error) {
	log.V(2).Info("Processing ConfigMap to add fluentdbit...")
	generator, err := fluentbit.NewConfigGenerator(false, false, false)
	if err != nil {
		return cm, fmt.Errorf("Unable to create collector config generator: %v", err)
	}
	conf, err := generator.Generate(&clfSpec, nil)
	if err != nil {
		return cm, fmt.Errorf("Error while generating fluentbit conf: %v", err)
	}
	cm.Data["fluent-bit.conf"] = conf
	cm.Data["concat-crio.lua"] = fluentbit.ConcatCrioFilter
	cm.Data["parsers.conf"] = fluentbit.Parsers
	return cm, nil
}

func VisitService(service *corev1.Service) *corev1.Service {
	log.V(2).Info("Processing Service to add fluentbit...")
	service.Spec.Ports = append(service.Spec.Ports, corev1.ServicePort{
		Name:       constants.CollectorMetricsPortName,
		Port:       constants.CollectorMetricsPort,
		TargetPort: intstr.FromString(constants.CollectorMetricsPortName),
	})
	return service
}
func VisitServiceMonitor(sm *monitoringv1.ServiceMonitor) *monitoringv1.ServiceMonitor {
	log.V(2).Info("Processing ServiceMonitor to add fluentbit...")
	sm.Spec.Endpoints = append(sm.Spec.Endpoints, monitoringv1.Endpoint{
		Port:   constants.CollectorMetricsPortName,
		Path:   "/api/v1/metrics/prometheus",
		Scheme: "http",
	})
	return sm
}
