package kafka

import (
	"fmt"

	"github.com/openshift/cluster-logging-operator/pkg/k8shandler"
	apps "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

func NewKafkaConsumerDeployment(namespace, topic string) *apps.Deployment {
	name := ConsumerNameForTopic(topic)
	podSpec := v1.PodSpec{
		InitContainers: []v1.Container{
			{
				Name:  "topic-create",
				Image: KafkaImage,
				Command: []string{
					"./bin/kafka-topics.sh",
					"--zookeeper",
					"zookeeper.openshift-logging.svc.cluster.local:2181",
					"--create",
					"--if-not-exists",
					"--topic",
					topic,
					"--partitions",
					"1",
					"--replication-factor",
					"1",
				},
				Resources: v1.ResourceRequirements{
					Limits: v1.ResourceList{
						v1.ResourceCPU:    resource.MustParse("200m"),
						v1.ResourceMemory: resource.MustParse("100Mi"),
					},
				},
			},
		},
		Containers: []v1.Container{
			{
				Name:  name,
				Image: KafkaImage,
				Command: []string{
					"/bin/bash",
					"-ce",
					fmt.Sprintf(
						`./bin/kafka-console-consumer.sh --bootstrap-server %s --topic %s --from-beginning --consumer.config /etc/kafka-configmap/client.properties > /shared/consumed.logs`,
						ClusterLocalEndpoint(namespace),
						topic,
					),
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
						Name:      "shared",
						MountPath: "/shared",
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
				Name: "shared",
				VolumeSource: v1.VolumeSource{
					EmptyDir: &v1.EmptyDirVolumeSource{},
				},
			},
		},
	}

	return k8shandler.NewDeployment(name, namespace, DeploymentName, name, podSpec)
}
