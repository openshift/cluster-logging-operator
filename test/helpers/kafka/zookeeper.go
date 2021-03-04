package kafka

import (
	"github.com/openshift/cluster-logging-operator/pkg/factory"
	"github.com/openshift/cluster-logging-operator/pkg/k8shandler"
	apps "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	zookeeperDeploymentName     = "zookeeper"
	zookeeperComponent          = "zookeeper"
	zookeeperProvider           = "openshift"
	zookeeperClientPort         = 2181
	zookeeperPeerPort           = 2888
	zookeeperLeaderElectionPort = 3888
)

func NewZookeeperStatefuleSet(namespace string) *apps.StatefulSet {
	var (
		replicas    int32 = 1
		termination int64 = 10
	)

	return &apps.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      zookeeperDeploymentName,
			Namespace: namespace,
			Labels: map[string]string{
				"app":       zookeeperDeploymentName,
				"component": zookeeperComponent,
				"provider":  zookeeperProvider,
			},
		},
		Spec: apps.StatefulSetSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": zookeeperDeploymentName,
				},
			},
			Replicas:    &replicas,
			ServiceName: zookeeperDeploymentName,
			UpdateStrategy: apps.StatefulSetUpdateStrategy{
				Type: apps.RollingUpdateStatefulSetStrategyType,
			},
			PodManagementPolicy: apps.PodManagementPolicyType("Parallel"),
			VolumeClaimTemplates: []v1.PersistentVolumeClaim{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "data",
						Namespace: namespace,
					},
					Spec: v1.PersistentVolumeClaimSpec{
						AccessModes: []v1.PersistentVolumeAccessMode{
							v1.ReadWriteOnce,
						},
						Resources: v1.ResourceRequirements{
							Requests: v1.ResourceList{
								v1.ResourceStorage: resource.MustParse("1Gi"),
							},
						},
					},
				},
			},
			Template: v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app":       zookeeperDeploymentName,
						"component": zookeeperComponent,
						"provider":  zookeeperProvider,
					},
				},
				Spec: v1.PodSpec{
					TerminationGracePeriodSeconds: &termination,
					InitContainers: []v1.Container{
						{
							Name:  "init-config",
							Image: KafkaInitUtilsImage,
							Command: []string{
								"/bin/bash",
								"/etc/kafka-configmap/init.sh",
							},
							VolumeMounts: []v1.VolumeMount{
								{
									Name:      "configmap",
									MountPath: "/etc/kafka-configmap",
								},
								{
									Name:      "config",
									MountPath: "/etc/kafka",
								},
								{
									Name:      "data",
									MountPath: "/var/lib/zookeeper",
								},
							},
						},
					},
					Containers: []v1.Container{
						{
							Name:  zookeeperDeploymentName,
							Image: KafkaImage,
							Env: []v1.EnvVar{

								{
									Name:  "KAFKA_LOG4J_OPTS",
									Value: "-Dlog4j.configuration=file:/etc/kafka/log4j.properties",
								},
							},
							Ports: []v1.ContainerPort{
								{
									Name:          "client",
									ContainerPort: zookeeperClientPort,
								},
								{
									Name:          "peer",
									ContainerPort: zookeeperPeerPort,
								},
								{
									Name:          "leader-election",
									ContainerPort: zookeeperLeaderElectionPort,
								},
							},
							Command: []string{
								"./bin/zookeeper-server-start.sh",
								"/etc/kafka/zookeeper.properties",
							},
							Lifecycle: &v1.Lifecycle{
								PreStop: &v1.Handler{
									Exec: &v1.ExecAction{
										Command: []string{
											"sh",
											"-ce",
											"kill -s TERM 1; while $(kill -0 1 2>/dev/null); do sleep 1; done",
										},
									},
								},
							},
							Resources: v1.ResourceRequirements{
								Requests: v1.ResourceList{
									v1.ResourceCPU:    resource.MustParse("10m"),
									v1.ResourceMemory: resource.MustParse("100Mi"),
								},
								Limits: v1.ResourceList{
									v1.ResourceMemory: resource.MustParse("120Mi"),
								},
							},
							VolumeMounts: []v1.VolumeMount{
								{
									Name:      "config",
									MountPath: "/etc/kafka",
								},
								{
									Name:      "zookeeperlogs",
									MountPath: "/opt/kafka/logs",
								},
								{
									Name:      "data",
									MountPath: "/var/lib/zookeeper",
								},
							},
						},
					},
					Volumes: []v1.Volume{
						{
							Name: "configmap",
							VolumeSource: v1.VolumeSource{
								ConfigMap: &v1.ConfigMapVolumeSource{
									LocalObjectReference: v1.LocalObjectReference{
										Name: zookeeperDeploymentName,
									},
								},
							},
						},
						{
							Name: "config",
							VolumeSource: v1.VolumeSource{
								EmptyDir: &v1.EmptyDirVolumeSource{},
							},
						},
						{
							Name: "zookeeperlogs",
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

func NewZookeeperService(namespace string) *v1.Service {
	ports := []v1.ServicePort{
		{
			Name: "client",
			Port: zookeeperClientPort,
		},
		{
			Name: "peer",
			Port: zookeeperPeerPort,
		},
		{
			Name: "leader-election",
			Port: zookeeperLeaderElectionPort,
		},
	}
	return factory.NewService(zookeeperDeploymentName, namespace, zookeeperComponent, ports)
}

func NewZookeeperConfigMap(namespace string) *v1.ConfigMap {
	data := map[string]string{
		"init.sh":              initZookeeperScript,
		"zookeeper.properties": zookeeperProperties,
		"log4j.properties":     zookeeperLog4JProperties,
	}
	return k8shandler.NewConfigMap(zookeeperDeploymentName, namespace, data)
}
