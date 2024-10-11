package metrics

import (
	"context"
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

var _ = Describe("Reconcile ServiceMonitor", func() {

	defer GinkgoRecover()

	_ = monitoringv1.AddToScheme(scheme.Scheme)

	var (

		// Adding ns and label to account for addSecurityLabelsToNamespace() added in LOG-2620
		namespace = &corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Labels: map[string]string{"test": "true"},
				Name:   constants.OpenshiftNS,
			},
		}

		reqClient = fake.NewFakeClient(
			namespace,
		)
		owner       = metav1.OwnerReference{}
		portName    = "test-port"
		serviceName = "test-service"
		selector    = map[string]string{
			constants.LabelK8sComponent: "test-component",
		}

		serviceMonitorKey = types.NamespacedName{Name: serviceName, Namespace: namespace.Name}
		smInstance        = &monitoringv1.ServiceMonitor{}
	)

	It("should successfully reconcile the ServiceMonitor", func() {
		// Reconcile the exporter daemonset
		Expect(ReconcileServiceMonitor(
			reqClient,
			constants.OpenshiftNS,
			serviceName,
			owner,
			selector,
			portName,
		)).To(Succeed())

		// Get and check the ServiceMonitor
		Expect(reqClient.Get(context.TODO(), serviceMonitorKey, smInstance)).Should(Succeed())

		Expect(smInstance.Name).To(Equal(serviceName))

		expJobLabel := fmt.Sprintf("monitor-%s", serviceName)
		Expect(smInstance.Spec.JobLabel).To(Equal(expJobLabel))
		Expect(smInstance.Spec.Endpoints).ToNot(BeEmpty())
		Expect(smInstance.Spec.Endpoints[0].Port).To(Equal(portName))

		svcURL := fmt.Sprintf("%s.openshift-logging.svc", serviceName)
		Expect(smInstance.Spec.Endpoints[0].TLSConfig.SafeTLSConfig.ServerName).
			To(Equal(svcURL))
	})
})
