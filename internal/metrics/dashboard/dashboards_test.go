package dashboard

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

var _ = Describe("ReconcileForDashboards", func() {
	var (
		fakeClient client.Client
		scheme     *runtime.Scheme

		GetDashboard = func() *corev1.ConfigMap {
			key := client.ObjectKeyFromObject(newDashboardConfigMap())
			actual := &corev1.ConfigMap{}
			Expect(fakeClient.Get(context.TODO(), key, actual)).To(Succeed(), "Exp the configmap to exist")
			actual.ResourceVersion = ""
			actual.TypeMeta = corev1.ConfigMap{}.TypeMeta
			return actual
		}

		setup = func(cm *corev1.ConfigMap) {
			if cm != nil {
				fakeClient = fake.NewClientBuilder().WithScheme(scheme).WithObjects(cm).Build()
			} else {
				fakeClient = fake.NewClientBuilder().WithScheme(scheme).Build()
			}
		}
		exp     = newDashboardConfigMap()
		initial *corev1.ConfigMap
	)

	BeforeEach(func() {
		scheme = runtime.NewScheme()
		Expect(corev1.AddToScheme(scheme)).To(Succeed())
		fakeClient = fake.NewClientBuilder().WithScheme(scheme).Build()
		initial = newDashboardConfigMap()
		exp = newDashboardConfigMap()
		exp.TypeMeta = corev1.ConfigMap{}.TypeMeta
	})

	Context("when the configmap does not exist", func() {
		BeforeEach(func() {
			setup(nil)
		})
		It("should create a new dashboard configmap", func() {
			Expect(ReconcileForDashboards(fakeClient, fakeClient)).To(Succeed())
			Expect(GetDashboard()).To(Equal(exp))
		})
	})

	Context("when the configmap does exist", func() {

		It("should update the configmap when the dashboard is different", func() {
			initial := newDashboardConfigMap()
			initial.Labels[DashboardHashName] = "abc"
			setup(initial)
			Expect(ReconcileForDashboards(fakeClient, fakeClient)).To(Succeed())
			Expect(GetDashboard()).To(Equal(exp), "Exp the configmap to be updated")
		})

		It("should do nothing to the configmap when the dashboard is the same", func() {
			setup(initial)
			Expect(ReconcileForDashboards(fakeClient, fakeClient)).To(Succeed())
			Expect(GetDashboard()).To(Equal(exp))
		})
	})

})
