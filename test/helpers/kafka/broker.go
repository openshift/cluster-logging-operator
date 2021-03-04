package kafka

import (
	"bytes"
	"fmt"
	"strconv"

	"github.com/openshift/cluster-logging-operator/pkg/factory"
	"github.com/openshift/cluster-logging-operator/pkg/k8shandler"
	"github.com/openshift/cluster-logging-operator/test/helpers/certificate"
	apps "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

const (
	// Kafka deployment definitions
	kafkaBrokerContainerName = "broker"
	kafkaBrokerComponent     = "kafka"
	kafkaBrokerProvider      = "openshift"
	kafkaNodeReader          = "kafka-node-reader"
	kafkaNodeReaderBinding   = "kafka-node-reader-binding"
	kafkaInsidePort          = 9093
	kafkaOutsidePort         = 9094
	kafkaJMXPort             = 5555
)

func NewBrokerStatefuleSet(namespace string) *apps.StatefulSet {
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
				"component": kafkaBrokerComponent,
				"provider":  kafkaBrokerProvider,
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
								v1.ResourceStorage: resource.MustParse("10Gi"),
							},
						},
					},
				},
			},
			Template: v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app":       DeploymentName,
						"component": kafkaBrokerComponent,
						"provider":  kafkaBrokerProvider,
					},
				},
				Spec: v1.PodSpec{
					TerminationGracePeriodSeconds: &termination,
					InitContainers: []v1.Container{
						{
							Name:  "init-config",
							Image: KafkaInitUtilsImage,
							Env: []v1.EnvVar{
								{
									Name: "NODE_NAME",
									ValueFrom: &v1.EnvVarSource{
										FieldRef: &v1.ObjectFieldSelector{
											FieldPath: "spec.nodeName",
										},
									},
								},
								{
									Name: "POD_NAME",
									ValueFrom: &v1.EnvVarSource{
										FieldRef: &v1.ObjectFieldSelector{
											FieldPath: "metadata.name",
										},
									},
								},
								{
									Name: "POD_NAMESPACE",
									ValueFrom: &v1.EnvVarSource{
										FieldRef: &v1.ObjectFieldSelector{
											FieldPath: "metadata.namespace",
										},
									},
								},
								{
									Name:  "ADVERTISE_ADDR",
									Value: fmt.Sprintf("%s.%s.svc.cluster.local", DeploymentName, namespace),
								},
							},
							Command: []string{
								"/bin/bash",
								"/etc/kafka-configmap/init.sh",
							},
							VolumeMounts: []v1.VolumeMount{
								{
									Name:      "brokerconfig",
									MountPath: "/etc/kafka-configmap",
								},
								{
									Name:      "config",
									MountPath: "/etc/kafka",
								},
								{
									Name:      "extensions",
									MountPath: "/opt/kafka/libs/extensions",
								},
							},
						},
					},
					Containers: []v1.Container{
						{
							Name:  kafkaBrokerContainerName,
							Image: KafkaImage,
							Env: []v1.EnvVar{
								{
									Name:  "CLASSPATH",
									Value: "/opt/kafka/libs/extensions/*",
								},
								{
									Name:  "KAFKA_LOG4J_OPTS",
									Value: "-Dlog4j.configuration=file:/etc/kafka/log4j.properties",
								},
								{
									Name:  "JMX_PORT",
									Value: strconv.Itoa(kafkaJMXPort),
								},
							},
							Ports: []v1.ContainerPort{
								{
									Name:          "inside",
									ContainerPort: kafkaInsidePort,
								},
								{
									Name:          "outside",
									ContainerPort: kafkaOutsidePort,
								},
								{
									Name:          "jmx",
									ContainerPort: kafkaJMXPort,
								},
							},
							Command: []string{
								"./bin/kafka-server-start.sh",
								"/etc/kafka/server.properties",
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
									v1.ResourceCPU:    resource.MustParse("250m"),
									v1.ResourceMemory: resource.MustParse("500Mi"),
								},
								Limits: v1.ResourceList{
									v1.ResourceMemory: resource.MustParse("1Gi"),
								},
							},
							ReadinessProbe: &v1.Probe{
								Handler: v1.Handler{
									TCPSocket: &v1.TCPSocketAction{
										Port: intstr.FromInt(kafkaInsidePort),
									},
								},
								TimeoutSeconds: 1,
							},
							VolumeMounts: []v1.VolumeMount{
								{
									Name:      "brokerconfig",
									MountPath: "/etc/kafka-configmap",
								},
								{
									Name:      "brokercerts",
									MountPath: "/etc/kafka-certs",
								},
								{
									Name:      "config",
									MountPath: "/etc/kafka",
								},
								{
									Name:      "brokerlogs",
									MountPath: "/opt/kafka/logs",
								},
								{
									Name:      "extensions",
									MountPath: "/opt/kafka/libs/extensions",
								},
								{
									Name:      "data",
									MountPath: "/var/lib/kafka/data",
								},
							},
						},
					},
					Volumes: []v1.Volume{
						{
							Name: "brokerconfig",
							VolumeSource: v1.VolumeSource{
								ConfigMap: &v1.ConfigMapVolumeSource{
									LocalObjectReference: v1.LocalObjectReference{
										Name: DeploymentName,
									},
								},
							},
						},
						{
							Name: "brokercerts",
							VolumeSource: v1.VolumeSource{
								Secret: &v1.SecretVolumeSource{
									SecretName: DeploymentName,
								},
							},
						},
						{
							Name: "brokerlogs",
							VolumeSource: v1.VolumeSource{
								EmptyDir: &v1.EmptyDirVolumeSource{},
							},
						},
						{
							Name: "config",
							VolumeSource: v1.VolumeSource{
								EmptyDir: &v1.EmptyDirVolumeSource{},
							},
						},
						{
							Name: "extensions",
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

func NewBrokerService(namespace string) *v1.Service {
	ports := []v1.ServicePort{
		{
			Name: "plaintext",
			Port: 9092,
		},
		{
			Name: "ssl",
			Port: 9093,
		},
	}
	return factory.NewService(DeploymentName, namespace, kafkaBrokerComponent, ports)
}

func NewBrokerRBAC(namespace string) (*rbacv1.ClusterRole, *rbacv1.ClusterRoleBinding) {
	cr := &rbacv1.ClusterRole{
		ObjectMeta: metav1.ObjectMeta{
			Name: kafkaNodeReader,
		},
		Rules: []rbacv1.PolicyRule{
			{
				APIGroups: []string{""},
				Resources: []string{"nodes"},
				Verbs:     []string{"get"},
			},
			{
				APIGroups: []string{""},
				Resources: []string{"pods"},
				Verbs:     []string{"get", "create", "update", "patch", "delete"},
			},
		},
	}

	crb := &rbacv1.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name: kafkaNodeReaderBinding,
		},
		Subjects: []rbacv1.Subject{
			{
				Kind:      "ServiceAccount",
				Name:      "default",
				Namespace: namespace,
			},
		},
		RoleRef: rbacv1.RoleRef{
			APIGroup: "rbac.authorization.k8s.io",
			Kind:     "ClusterRole",
			Name:     kafkaNodeReader,
		},
	}

	return cr, crb
}

func NewBrokerConfigMap(namespace string) *v1.ConfigMap {
	data := map[string]string{
		"init.sh":           initKafkaScript,
		"server.properties": serverProperties,
		"client.properties": clientProperties,
		"log4j.properties":  log4jProperties,
	}
	return k8shandler.NewConfigMap(DeploymentName, namespace, data)
}

func NewBrokerSecret(namespace string) *v1.Secret {
	rootCA := certificate.NewCA(nil, "Root CA")
	intermediateCA := certificate.NewCA(rootCA, "Intermediate CA")
	serverCert := certificate.NewCert(intermediateCA, "Server", fmt.Sprintf("%s.%s.svc.cluster.local", DeploymentName, namespace))

	data := map[string][]byte{
		"server.jks":    certificate.JKSKeyStore(serverCert, "server"),
		"ca-bundle.jks": certificate.JKSTrustStore([]*certificate.CertKey{rootCA, intermediateCA}, "ca-bundle"),
		"ca-bundle.crt": bytes.Join([][]byte{rootCA.CertificatePEM(), intermediateCA.CertificatePEM()}, []byte{}),
	}

	secret := k8shandler.NewSecret(
		DeploymentName,
		namespace,
		data,
	)
	return secret
}
