package loki

import (
	"github.com/openshift/cluster-logging-operator/pkg/k8shandler"
	apps "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

const (
	lokiComponent = "loki"
	lokiProvider  = "openshift"
)

func NewStatefulSet(namespace string) *apps.StatefulSet {
	var (
		replicas    int32 = 1
		termination int64 = 30
	)
	return &apps.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      DeploymentName,
			Namespace: namespace,
			Labels: map[string]string{
				"app":       DeploymentName,
				"component": lokiComponent,
				"provider":  lokiProvider,
			},
		},
		Spec: apps.StatefulSetSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": DeploymentName,
				},
			},
			Replicas:    &replicas,
			ServiceName: DeploymentName,
			UpdateStrategy: apps.StatefulSetUpdateStrategy{
				Type: apps.RollingUpdateStatefulSetStrategyType,
			},
			PodManagementPolicy: apps.PodManagementPolicyType("OrderedReady"),
			Template: v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app":       DeploymentName,
						"component": lokiComponent,
						"provider":  lokiProvider,
					},
				},
				Spec: v1.PodSpec{
					TerminationGracePeriodSeconds: &termination,
					Containers: []v1.Container{
						{
							Name:  "loki",
							Image: "grafana/loki:v1.3.0",
							Args: []string{
								"-config.file=/etc/loki/loki.yaml",
								"-log.level=debug",
							},
							Ports: []v1.ContainerPort{
								{
									Name:          "http-metrics",
									ContainerPort: ListenerPort,
									Protocol:      v1.ProtocolTCP,
								},
							},
							LivenessProbe: &v1.Probe{
								Handler: v1.Handler{
									HTTPGet: &v1.HTTPGetAction{
										Path: "/ready",
										Port: intstr.FromInt(ListenerPort),
									},
								},
								InitialDelaySeconds: 15,
							},
							ReadinessProbe: &v1.Probe{
								Handler: v1.Handler{
									HTTPGet: &v1.HTTPGetAction{
										Path: "/ready",
										Port: intstr.FromInt(ListenerPort),
									},
								},
								InitialDelaySeconds: 15,
							},
							Resources: v1.ResourceRequirements{
								Requests: v1.ResourceList{
									v1.ResourceCPU:    resource.MustParse("500m"),
									v1.ResourceMemory: resource.MustParse("1Gi"),
								},
								Limits: v1.ResourceList{
									v1.ResourceCPU:    resource.MustParse("1000m"),
									v1.ResourceMemory: resource.MustParse("2Gi"),
								},
							},
							VolumeMounts: []v1.VolumeMount{
								{
									Name:      "config",
									MountPath: "/etc/loki",
								},
								{
									Name:      "storage",
									MountPath: "/data",
								},
							},
						},
					},
					Volumes: []v1.Volume{
						{
							Name: "config",
							VolumeSource: v1.VolumeSource{
								ConfigMap: &v1.ConfigMapVolumeSource{
									LocalObjectReference: v1.LocalObjectReference{
										Name: DeploymentName,
									},
								},
							},
						},
						{
							Name: "storage",
							VolumeSource: v1.VolumeSource{
								EmptyDir: &v1.EmptyDirVolumeSource{},
							},
						},
					},
				},
			},
		},
	}
}

func NewService(namespace string) *v1.Service {
	ports := []v1.ServicePort{
		{
			Name:       "server",
			Port:       ListenerPort,
			Protocol:   v1.ProtocolTCP,
			TargetPort: intstr.FromString("http-metrics"),
		},
	}
	return k8shandler.NewService(DeploymentName, namespace, lokiComponent, ports)
}

func NewConfigMap(namespace string) *v1.ConfigMap {
	data := map[string]string{
		"loki.yaml": lokiYaml,
	}
	return k8shandler.NewConfigMap(DeploymentName, namespace, data)
}
