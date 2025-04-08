package initialize

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	"github.com/openshift/cluster-logging-operator/internal/utils"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

var _ = Describe("#Resources", func() {
	var (
		forwarder    obs.ClusterLogForwarder
		expResources = &v1.ResourceRequirements{
			Limits: v1.ResourceList{
				v1.ResourceMemory: resource.MustParse("2048Mi"),
				v1.ResourceCPU:    resource.MustParse("6000m"),
			},
			Requests: v1.ResourceList{
				v1.ResourceMemory: resource.MustParse("64Mi"),
				v1.ResourceCPU:    resource.MustParse("500m"),
			},
		}
	)

	BeforeEach(func() {
		forwarder = obs.ClusterLogForwarder{
			Spec: obs.ClusterLogForwarderSpec{},
		}
	})

	It("should apply the default resources when collector is not defined", func() {
		forwarder := Resources(forwarder, utils.NoOptions)
		Expect(forwarder.Spec.Collector.Resources).To(Equal(expResources))
	})
	It("should apply the default resources when resources are not defined", func() {
		forwarder.Spec.Collector = &obs.CollectorSpec{}
		forwarder := Resources(forwarder, utils.NoOptions)
		Expect(forwarder.Spec.Collector.Resources).To(Equal(expResources))
	})
	It("should apply the spec'd resources when defined", func() {
		resources := &v1.ResourceRequirements{
			Limits: v1.ResourceList{
				v1.ResourceMemory: resource.MustParse("2048Mi"),
				v1.ResourceCPU:    resource.MustParse("6000m"),
			},
			Requests: v1.ResourceList{
				v1.ResourceMemory: resource.MustParse("64Mi"),
				v1.ResourceCPU:    resource.MustParse("500m"),
			},
		}
		forwarder.Spec.Collector = &obs.CollectorSpec{
			Resources: resources,
		}
		forwarder := Resources(forwarder, utils.NoOptions)
		Expect(forwarder.Spec.Collector.Resources).To(Equal(resources))
	})

})
