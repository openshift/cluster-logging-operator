package vector

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

var _ = Describe("vectorThreads", func() {
	DescribeTable("deriving the worker-thread count from the CPU limit",
		func(resources corev1.ResourceRequirements, wantValue string, wantOk bool) {
			value, ok := vectorThreads(resources)
			Expect(ok).To(Equal(wantOk))
			if wantOk {
				Expect(value).To(Equal(wantValue))
			}
		},
		Entry("returns nothing when no resources are set",
			corev1.ResourceRequirements{}, "", false),
		Entry("returns nothing when only requests are set",
			corev1.ResourceRequirements{Requests: corev1.ResourceList{corev1.ResourceCPU: resource.MustParse("500m")}}, "", false),
		Entry("returns nothing when limits are set but no CPU limit",
			corev1.ResourceRequirements{Limits: corev1.ResourceList{corev1.ResourceMemory: resource.MustParse("2Gi")}}, "", false),
		Entry("uses an integer CPU limit of 6 cores",
			corev1.ResourceRequirements{Limits: corev1.ResourceList{corev1.ResourceCPU: resource.MustParse("6")}}, "6", true),
		Entry("uses a millicpu limit of 6000m",
			corev1.ResourceRequirements{Limits: corev1.ResourceList{corev1.ResourceCPU: resource.MustParse("6000m")}}, "6", true),
		Entry("floors a fractional limit to whole cores",
			corev1.ResourceRequirements{Limits: corev1.ResourceList{corev1.ResourceCPU: resource.MustParse("1500m")}}, "1", true),
		Entry("clamps a sub-core limit to a minimum of 1",
			corev1.ResourceRequirements{Limits: corev1.ResourceList{corev1.ResourceCPU: resource.MustParse("500m")}}, "1", true),
	)
})
