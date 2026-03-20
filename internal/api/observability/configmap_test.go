package observability_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	internalobs "github.com/openshift/cluster-logging-operator/internal/api/observability"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("ConfigMaps ", func() {

	It("should return different hashes when configmap content changes", func() {
			original := internalobs.ConfigMaps{
				"ca-test": &corev1.ConfigMap{
					ObjectMeta: metav1.ObjectMeta{Name: "ca-test"},
					Data:       map[string]string{"service-ca.crt": "-----BEGIN CERTIFICATE-----\nORIGINAL\n-----END CERTIFICATE-----"},
				},
			}
			modified := internalobs.ConfigMaps{
				"ca-test": &corev1.ConfigMap{
					ObjectMeta: metav1.ObjectMeta{Name: "ca-test"},
					Data:       map[string]string{"service-ca.crt": "-----BEGIN CERTIFICATE-----\nUPDATED\n-----END CERTIFICATE-----"},
				},
			}
			Expect(original.Hash64a()).ToNot(Equal(modified.Hash64a()))
		})

		It("should return the same hash for identical content", func() {
			cm1 := internalobs.ConfigMaps{
				"ca-test": &corev1.ConfigMap{
					ObjectMeta: metav1.ObjectMeta{Name: "ca-test"},
					Data:       map[string]string{"service-ca.crt": "same-content"},
				},
			}
			cm2 := internalobs.ConfigMaps{
				"ca-test": &corev1.ConfigMap{
					ObjectMeta: metav1.ObjectMeta{Name: "ca-test"},
					Data:       map[string]string{"service-ca.crt": "same-content"},
				},
			}
			Expect(cm1.Hash64a()).To(Equal(cm2.Hash64a()))
		})

		It("should return a consistent hash for empty configmaps", func() {
			empty := internalobs.ConfigMaps{}
			Expect(empty.Hash64a()).ToNot(BeEmpty())
	})
})
