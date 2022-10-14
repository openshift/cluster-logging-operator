package services_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/openshift/cluster-logging-operator/internal/utils/comparators/services"
)

var _ = Describe("services#AreSame", func() {

	var (
		current, desired *v1.Service
	)

	BeforeEach(func() {
		current = &v1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Labels: map[string]string{},
			},
			Spec: v1.ServiceSpec{
				Selector: map[string]string{},
				Ports:    []v1.ServicePort{},
			},
		}
		desired = current.DeepCopy()

	})

	Context("when evaluating labels", func() {
		It("should recognize they are the same", func() {
			ok, _ := services.AreSame(current, desired)
			Expect(ok).To(BeTrue())
		})

		It("should recognize they are different", func() {
			desired.Labels["foo"] = "bar"
			ok, _ := services.AreSame(current, desired)
			Expect(ok).To(BeFalse())
		})
	})
	Context("when evaluating selectors", func() {
		It("should recognize they are the same", func() {
			ok, _ := services.AreSame(current, desired)
			Expect(ok).To(BeTrue())
		})

		It("should recognize they are different", func() {
			desired.Spec.Selector["foo"] = "bar"
			ok, _ := services.AreSame(current, desired)
			Expect(ok).To(BeFalse())
		})
	})
	Context("when evaluating ServicePorts", func() {
		It("should recognize they are the same", func() {
			ok, _ := services.AreSame(current, desired)
			Expect(ok).To(BeTrue())
		})

		It("should recognize they are different lengths", func() {
			desired.Spec.Ports = append(desired.Spec.Ports, v1.ServicePort{})
			ok, _ := services.AreSame(current, desired)
			Expect(ok).To(BeFalse())
		})

		It("should recognize they are different content", func() {
			current.Spec.Ports = append(desired.Spec.Ports, v1.ServicePort{Name: "bar", Port: 1051})
			desired.Spec.Ports = append(desired.Spec.Ports, v1.ServicePort{Name: "bar", Port: 1050})
			ok, _ := services.AreSame(current, desired)
			Expect(ok).To(BeFalse())
		})
	})
})
