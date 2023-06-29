package secrets_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/openshift/cluster-logging-operator/internal/utils/comparators/secrets"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("secrets#AreSame", func() {

	var (
		current, desired *v1.Secret
	)

	BeforeEach(func() {
		current = &v1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Annotations: map[string]string{"one": "two"},
			},
			Data: map[string][]byte{"foo": []byte("bar")},
		}
		desired = current.DeepCopy()

	})

	Context("when no comparison options are provided", func() {
		It("should recognize when they are the same", func() {
			Expect(secrets.AreSame(current, desired)).To(BeTrue())
		})
		It("should recognize when only the data is different", func() {
			current.Data["xyz"] = []byte("abc")
			Expect(secrets.AreSame(current, desired)).To(BeFalse())
		})
	})
	Context("when optionally comparing labels and annotations", func() {
		BeforeEach(func() {
			current.Labels = map[string]string{"foo": "bar"}
			desired = current.DeepCopy()
		})
		It("should recognize when they are the same", func() {
			Expect(secrets.AreSame(current, desired, secrets.CompareLabels, secrets.CompareAnnotations)).To(BeTrue())
		})
		It("should recognize when the labels are different", func() {
			current.Labels["foo"] = "abc"
			Expect(secrets.AreSame(current, desired, secrets.CompareLabels, secrets.CompareAnnotations)).To(BeFalse())
		})
		It("should recognize when the annotations are different", func() {
			current.Annotations["foo"] = "abc"
			Expect(secrets.AreSame(current, desired, secrets.CompareLabels, secrets.CompareAnnotations)).To(BeFalse())
		})
	})
})
