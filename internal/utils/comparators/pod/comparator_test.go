package pod_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/openshift/cluster-logging-operator/internal/utils/comparators/pod"
	v1 "k8s.io/api/core/v1"
)

var _ = Describe("pod#AreSame", func() {

	var (
		current, desired *v1.PodSpec
	)

	BeforeEach(func() {
		current = &v1.PodSpec{
			Containers: []v1.Container{
				{},
			},
			InitContainers: []v1.Container{
				{},
			},
		}
		desired = current.DeepCopy()
	})

	Context("when evaluating containers", func() {

		It("should recognize the numbers are different", func() {
			container := v1.Container{}
			desired.Containers = append(desired.Containers, container)
			ok, _ := pod.AreSame(current, desired, "")
			Expect(ok).To(BeFalse())
		})

		It("should recognize different images", func() {
			desired.Containers[0].Image = "bar"
			ok, _ := pod.AreSame(current, desired, "")
			Expect(ok).To(BeFalse())
		})
	})

	Context("when evaluating init containers", func() {

		It("should recognize the numbers are different", func() {
			container := v1.Container{}
			desired.InitContainers = append(desired.InitContainers, container)
			ok, _ := pod.AreSame(current, desired, "")
			Expect(ok).To(BeFalse())
		})

		It("should recognize different images", func() {
			desired.InitContainers[0].Image = "bar"
			ok, _ := pod.AreSame(current, desired, "")
			Expect(ok).To(BeFalse())
		})
	})

})
