package vector

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/openshift/cluster-logging-operator/internal/factory"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

var _ = Describe("CollectorVisitor", func() {
	var resNames = &factory.ForwarderResourceNames{ForwarderName: "test", ConfigMap: "test-config"}

	envValue := func(container *corev1.Container, name string) (string, bool) {
		for _, e := range container.Env {
			if e.Name == name {
				return e.Value, true
			}
		}
		return "", false
	}

	Context("VECTOR_THREADS", func() {
		It("is set from the CPU limit when one is present", func() {
			container := &corev1.Container{
				Resources: corev1.ResourceRequirements{
					Limits: corev1.ResourceList{corev1.ResourceCPU: resource.MustParse("6")},
				},
			}

			CollectorVisitor(container, &corev1.PodSpec{}, resNames, "test-ns", "warn")

			value, ok := envValue(container, "VECTOR_THREADS")
			Expect(ok).To(BeTrue(), "expected VECTOR_THREADS to be set when a CPU limit is present")
			Expect(value).To(Equal("6"))
		})

		It("is omitted when no CPU limit is set", func() {
			container := &corev1.Container{
				Resources: corev1.ResourceRequirements{
					Requests: corev1.ResourceList{corev1.ResourceCPU: resource.MustParse("500m")},
				},
			}

			CollectorVisitor(container, &corev1.PodSpec{}, resNames, "test-ns", "warn")

			_, ok := envValue(container, "VECTOR_THREADS")
			Expect(ok).To(BeFalse(), "expected VECTOR_THREADS to be omitted when no CPU limit is set")
		})
	})
})
