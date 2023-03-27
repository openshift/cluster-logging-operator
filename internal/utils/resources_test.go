package utils

import (
	"k8s.io/apimachinery/pkg/api/resource"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	v1 "k8s.io/api/core/v1"
)

var (
	requestMemory = resource.MustParse("100Gi")
	requestCPU    = resource.MustParse("500m")
)

var _ = Describe("#AreResourcesSame", func() {
	It("should be true for nil requirements", func() {
		Expect(AreResourcesSame(nil, nil)).To(BeTrue())
	})

	It("should be false when one is nil and the other is not", func() {
		Expect(AreResourcesSame(nil, &v1.ResourceRequirements{})).To(BeFalse())
		Expect(AreResourcesSame(&v1.ResourceRequirements{}, nil)).To(BeFalse())
	})
	It("should be false limits.cpu is different", func() {
		left := &v1.ResourceRequirements{Limits: v1.ResourceList{v1.ResourceCPU: requestCPU}}
		right := &v1.ResourceRequirements{}
		Expect(AreResourcesSame(left, right)).To(BeFalse())
		Expect(AreResourcesSame(right, left)).To(BeFalse())
	})
	It("should be false requests.cpu is different", func() {
		left := &v1.ResourceRequirements{Requests: v1.ResourceList{v1.ResourceCPU: requestCPU}}
		right := &v1.ResourceRequirements{}
		Expect(AreResourcesSame(left, right)).To(BeFalse())
		Expect(AreResourcesSame(right, left)).To(BeFalse())
	})
	It("should be false limits.memory is different", func() {
		left := &v1.ResourceRequirements{Limits: v1.ResourceList{v1.ResourceMemory: requestMemory}}
		right := &v1.ResourceRequirements{}
		Expect(AreResourcesSame(left, right)).To(BeFalse())
		Expect(AreResourcesSame(right, left)).To(BeFalse())
	})
	It("should be false requests.memory is different", func() {
		left := &v1.ResourceRequirements{Requests: v1.ResourceList{v1.ResourceMemory: requestMemory}}
		right := &v1.ResourceRequirements{}
		Expect(AreResourcesSame(left, right)).To(BeFalse())
		Expect(AreResourcesSame(right, left)).To(BeFalse())
	})
	It("should be true when limits and requests are the same", func() {
		left := &v1.ResourceRequirements{Requests: v1.ResourceList{v1.ResourceMemory: requestMemory}}
		right := &v1.ResourceRequirements{Requests: v1.ResourceList{v1.ResourceMemory: requestMemory}}
		Expect(AreResourcesSame(left, right)).To(BeTrue())
		Expect(AreResourcesSame(right, left)).To(BeTrue())
	})
})
