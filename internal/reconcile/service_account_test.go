package reconcile_test

import (
	"context"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/openshift/cluster-logging-operator/internal/reconcile"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

var _ = Describe("ServiceAccount reconcile", func() {
	It("should propagate finalizers from desired to reconciled ServiceAccount", func() {
		c := fake.NewFakeClient()
		desired := &corev1.ServiceAccount{
			ObjectMeta: metav1.ObjectMeta{
				Name:       "logfilesmetricexporter",
				Namespace:  "openshift-logging",
				Finalizers: []string{metav1.FinalizerDeleteDependents},
			},
		}

		_, err := reconcile.ServiceAccount(c, desired)
		Expect(err).To(Succeed())

		got := &corev1.ServiceAccount{}
		key := client.ObjectKey{Name: "logfilesmetricexporter", Namespace: "openshift-logging"}
		Expect(c.Get(context.TODO(), key, got)).To(Succeed())

		Expect(got.Finalizers).To(HaveLen(1))
		Expect(got.Finalizers[0]).To(Equal(metav1.FinalizerDeleteDependents))
	})
})
