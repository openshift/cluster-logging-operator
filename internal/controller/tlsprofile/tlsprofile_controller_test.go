package tlsprofile_test

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	configv1 "github.com/openshift/api/config/v1"
	"github.com/openshift/cluster-logging-operator/internal/controller/tlsprofile"
	"github.com/openshift/cluster-logging-operator/internal/tls"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

var _ = Describe("TLSProfileReconciler", func() {
	var (
		reconciler *tlsprofile.TLSProfileReconciler
		k8sClient  client.Client
		scheme     *runtime.Scheme
	)

	BeforeEach(func() {
		scheme = runtime.NewScheme()
		Expect(configv1.AddToScheme(scheme)).To(Succeed())

		k8sClient = fake.NewClientBuilder().
			WithScheme(scheme).
			Build()

		reconciler = &tlsprofile.TLSProfileReconciler{
			Client:         k8sClient,
			InitialProfile: nil,
		}
	})

	Context("when APIServer is not the cluster APIServer", func() {
		It("should not process the request", func() {
			apiServer := &configv1.APIServer{
				ObjectMeta: metav1.ObjectMeta{
					Name: "not-cluster",
				},
			}
			Expect(k8sClient.Create(context.TODO(), apiServer)).To(Succeed())

			req := ctrl.Request{
				NamespacedName: types.NamespacedName{
					Name: "not-cluster",
				},
			}

			result, err := reconciler.Reconcile(context.TODO(), req)
			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(Equal(ctrl.Result{}))
		})
	})

	Context("when APIServer does not exist", func() {
		It("should return without error", func() {
			req := ctrl.Request{
				NamespacedName: types.NamespacedName{
					Name: tls.APIServerName,
				},
			}

			result, err := reconciler.Reconcile(context.TODO(), req)
			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(Equal(ctrl.Result{}))
		})
	})

	Context("when TLS profile has not changed", func() {
		It("should not exit", func() {
			initialProfile := &configv1.TLSSecurityProfile{
				Type: configv1.TLSProfileIntermediateType,
			}

			reconciler.InitialProfile = initialProfile

			apiServer := &configv1.APIServer{
				ObjectMeta: metav1.ObjectMeta{
					Name: tls.APIServerName,
				},
				Spec: configv1.APIServerSpec{
					TLSSecurityProfile: initialProfile,
				},
			}
			Expect(k8sClient.Create(context.TODO(), apiServer)).To(Succeed())

			req := ctrl.Request{
				NamespacedName: types.NamespacedName{
					Name: tls.APIServerName,
				},
			}

			result, err := reconciler.Reconcile(context.TODO(), req)
			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(Equal(ctrl.Result{}))
		})
	})
})
