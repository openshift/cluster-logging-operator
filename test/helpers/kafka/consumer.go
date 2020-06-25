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
				Image: "solsson/kafka-cli@sha256:9fa3306e9f5d18283d10e01f7c115d8321eedc682f262aff784bd0126e1f2221",
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
				Image: "solsson/kafka:2.4.1",
				Command: []string{
					"/bin/bash",
					"-ce",
					fmt.Sprintf(
						`./bin/kafka-console-consumer.sh --bootstrap-server %s --topic %s --from-beginning | tee /shared/consumed.logs`,
						ClusterLocalEndpoint(namespace),
						topic,
					),
				},
				VolumeMounts: []v1.VolumeMount{
					{
						Name:      "shared",
						MountPath: "/shared",
					},
				},
			},
		},
		Volumes: []v1.Volume{
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
