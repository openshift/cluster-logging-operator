// Copyright 2020 Red Hat, Inc
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
// http://www.apache.org/licenses/LICENSE-2.0
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package edge

import (
	"github.com/ViaQ/logerr/log"
	monitoringv1 "github.com/coreos/prometheus-operator/pkg/apis/monitoring/v1"
	configv1 "github.com/openshift/api/config/v1"
	"github.com/openshift/cluster-logging-operator/pkg/constants"
	"github.com/openshift/cluster-logging-operator/pkg/factory"
	"github.com/openshift/cluster-logging-operator/pkg/generators/forwarding/fluentbit"
	topologyapi "github.com/openshift/cluster-logging-operator/pkg/k8shandler/logforwardingtopology"
	"github.com/openshift/cluster-logging-operator/pkg/utils"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

//EdgeTopologyEnhanced defines a topology with edge normalization that relies
//upon multiple containers to collect logs
type EdgeTopologyEnhanced struct {
	ReconcileCollector      func(proxyConfig *configv1.Proxy, topology topologyapi.LogForwarderTopology) (err error)
	GenerateCollectorConfig func() (string, error)
	UndeployCollector       func() error
}

//Reconcile enhanced edge topology to to spec
func (topology EdgeTopologyEnhanced) Reconcile(proxyConfig *configv1.Proxy) (err error) {
	return topology.ReconcileCollector(proxyConfig, topology)
}

//ProcessPodSpec allows reuse of exising functions of the original deployment
func (topology EdgeTopologyEnhanced) ProcessPodSpec(podSpec *corev1.PodSpec) *corev1.PodSpec {
	log.V(2).Info("EdgeTopologyEnhanced Processing PodSpec..")
	resourceRequirements := corev1.ResourceRequirements{}
	container := factory.NewContainer(constants.CollectorName, constants.FluentBitImage, corev1.PullIfNotPresent, resourceRequirements)
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

func (topology EdgeTopologyEnhanced) ProcessConfigMap(cm *corev1.ConfigMap) *corev1.ConfigMap {
	log.V(2).Info("EdgeTopologyEnhanced Processing ConfigMap..")
	conf, err := topology.GenerateCollectorConfig()
	if err != nil {
		log.V(0).Error(err, "Error while generating fluentbit conf")
	}
	cm.Data["fluent-bit.conf"] = conf
	cm.Data["concat-crio.lua"] = fluentbit.ConcatCrioFilter
	cm.Data["parsers.conf"] = fluentbit.Parsers
	return cm
}

func (topology EdgeTopologyEnhanced) ProcessService(service *corev1.Service) *corev1.Service {
	log.V(2).Info("EdgeTopologyEnhanced Processing Service..")
	service.Spec.Ports = append(service.Spec.Ports, corev1.ServicePort{
		Name:       constants.CollectorMetricsPortName,
		Port:       constants.CollectorMetricsPort,
		TargetPort: intstr.FromString(constants.CollectorMetricsPortName),
	})
	return service
}
func (topology EdgeTopologyEnhanced) ProcessServiceMonitor(sm *monitoringv1.ServiceMonitor) *monitoringv1.ServiceMonitor {
	log.V(2).Info("EdgeTopologyEnhanced Processing ServiceMonitor..")
	sm.Spec.Endpoints = append(sm.Spec.Endpoints, monitoringv1.Endpoint{
		Port:   constants.CollectorMetricsPortName,
		Path:   "/api/v1/metrics/prometheus",
		Scheme: "http",
	})
	return sm
}

func (topology EdgeTopologyEnhanced) Undeploy() error {
	return topology.UndeployCollector()
}

//Name of the collector topology
func (topology EdgeTopologyEnhanced) Name() string {
	return topologyapi.LogForwardingEnhancedEdgeNormalizationTopology
}
