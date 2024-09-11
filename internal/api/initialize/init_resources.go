package initialize

import (
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	"github.com/openshift/cluster-logging-operator/internal/utils"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

var (

	//https://vector.dev/docs/reference/configuration/sources/kubernetes_logs/#resource-limits

	DefaultRequestMemory = resource.MustParse("64Mi")
	DefaultRequestCpu    = resource.MustParse("500m")
	DefaultLimitMemory   = resource.MustParse("2048Mi")
	DefaultLimitCpu      = resource.MustParse("6000m")
)

func Resources(forwarder obs.ClusterLogForwarder, options utils.Options) obs.ClusterLogForwarder {
	if forwarder.Spec.Collector == nil {
		forwarder.Spec.Collector = &obs.CollectorSpec{}
	}
	if forwarder.Spec.Collector.Resources == nil {
		forwarder.Spec.Collector.Resources = &corev1.ResourceRequirements{
			Limits: corev1.ResourceList{
				corev1.ResourceMemory: DefaultLimitMemory,
				corev1.ResourceCPU:    DefaultLimitCpu,
			},
			Requests: corev1.ResourceList{
				corev1.ResourceMemory: DefaultRequestMemory,
				corev1.ResourceCPU:    DefaultRequestCpu,
			},
		}
	}
	return forwarder
}
