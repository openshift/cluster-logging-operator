package daemonsets_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	apps "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"

	"github.com/openshift/cluster-logging-operator/internal/utils/comparators/daemonsets"
)

var _ = Describe("daemonset#AreSame", func() {

	var (
		current, desired *apps.DaemonSet
	)

	BeforeEach(func() {
		current = &apps.DaemonSet{
			Spec: apps.DaemonSetSpec{
				Template: v1.PodTemplateSpec{
					Spec: v1.PodSpec{
						Containers: []v1.Container{
							{},
						},
						InitContainers: []v1.Container{
							{},
						},
					},
				},
			},
		}
		desired = current.DeepCopy()

	})

	Context("when evaluating containers", func() {

		It("should recognize the numbers are different", func() {
			container := v1.Container{}
			desired.Spec.Template.Spec.Containers = append(desired.Spec.Template.Spec.Containers, container)
			Expect(daemonsets.AreSame(current, desired)).To(BeFalse())
		})

		It("should recognize different images", func() {
			desired.Spec.Template.Spec.Containers[0].Image = "bar"
			Expect(daemonsets.AreSame(current, desired)).To(BeFalse())
		})
	})

	Context("when evaluating init containers", func() {

		It("should recognize the numbers are different", func() {
			container := v1.Container{}
			desired.Spec.Template.Spec.InitContainers = append(desired.Spec.Template.Spec.InitContainers, container)
			Expect(daemonsets.AreSame(current, desired)).To(BeFalse())
		})

		It("should recognize different images", func() {
			desired.Spec.Template.Spec.InitContainers[0].Image = "bar"
			Expect(daemonsets.AreSame(current, desired)).To(BeFalse())
		})
	})

})
