package loki

import (
	"github.com/openshift/cluster-logging-operator/pkg/k8shandler"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func NewQuerierPod(namespace string) *v1.Pod {
	var cmDefaultMode int32 = 0755
	return &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      QuerierName,
			Namespace: namespace,
		},
		Spec: v1.PodSpec{
			Containers: []v1.Container{
				{
					Name:            QuerierName,
					Image:           "alpine:3.11",
					ImagePullPolicy: v1.PullAlways,
					Args: []string{
						"sh",
						"-c",
						"apk add --no-cache curl; while true; do sleep 1; done",
					},
					VolumeMounts: []v1.VolumeMount{
						{
							Name:      "utils",
							MountPath: "/data/",
							ReadOnly:  false,
						},
					},
				},
			},
			Volumes: []v1.Volume{
				{
					Name: "utils",
					VolumeSource: v1.VolumeSource{
						ConfigMap: &v1.ConfigMapVolumeSource{
							LocalObjectReference: v1.LocalObjectReference{
								Name: QuerierName,
							},
							DefaultMode: &cmDefaultMode,
						},
					},
				},
			},
		},
	}
}

func NewQuerierConfigMap(namespace string) *v1.ConfigMap {
	data := map[string]string{
		"loki_util": lokiUtil,
	}
	return k8shandler.NewConfigMap(QuerierName, namespace, data)
}
