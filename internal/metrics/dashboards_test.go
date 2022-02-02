package metrics

import (
	"context"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

var _ = Describe("ReconcileDashboards", func() {
	var (
		fakeClient client.Client

		GetDashboard = func() *corev1.ConfigMap {
			key := client.ObjectKeyFromObject(newDashboardConfigMap())
			actual := &corev1.ConfigMap{}
			Expect(fakeClient.Get(context.TODO(), key, actual)).To(Succeed(), "Exp the configmap to exist")
			actual.ResourceVersion = ""
			return actual
		}

		setup = func(cm *corev1.ConfigMap) {
			if cm != nil {
				fakeClient = fake.NewClientBuilder().WithObjects(cm).Build()
			}
		}
		exp     = newDashboardConfigMap()
		initial *corev1.ConfigMap
	)

	BeforeEach(func() {
		fakeClient = fake.NewClientBuilder().Build()
		initial = newDashboardConfigMap()
	})

	Context("when the configmap does not exist", func() {
		BeforeEach(func() {
			setup(nil)
		})
		It("should create a new dashboard configmap", func() {
			Expect(ReconcileDashboards(fakeClient, fakeClient)).To(Succeed())
			Expect(GetDashboard()).To(Equal(exp))
		})
	})

	Context("when the configmap does exist", func() {

		It("should update the configmap when the dashboard is different", func() {
			initial := newDashboardConfigMap()
			initial.Labels[constants.TrustedCABundleHashName] = "abc"
			setup(initial)
			Expect(ReconcileDashboards(fakeClient, fakeClient)).To(Succeed())
			Expect(GetDashboard()).To(Equal(exp), "Exp the configmap to be updated")
		})

		It("should do nothing to the configmap when the dashboard is the same", func() {
			setup(initial)
			Expect(ReconcileDashboards(fakeClient, fakeClient)).To(Succeed())
			Expect(GetDashboard()).To(Equal(exp))
		})
	})

})
