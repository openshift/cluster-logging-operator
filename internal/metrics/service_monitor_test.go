package metrics

import (
	"context"
	"fmt"

	. "github.com/onsi/ginkgo/v2"
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

	It("should reconcile a full profile ServiceMonitor with only the rename rule", func() {
		Expect(ReconcileServiceMonitor(
			reqClient,
			constants.OpenshiftNS,
			serviceName,
			serviceName,
			owner,
			selector,
			portName,
			FullRelabelConfigs,
			constants.MetricsCollectionProfileFull,
		)).To(Succeed())

		Expect(reqClient.Get(context.TODO(), serviceMonitorKey, smInstance)).Should(Succeed())

		Expect(smInstance.Name).To(Equal(serviceName))
		Expect(smInstance.Labels[constants.LabelMetricsCollectionProfile]).To(Equal(constants.MetricsCollectionProfileFull))

		expJobLabel := fmt.Sprintf("monitor-%s", serviceName)
		Expect(smInstance.Spec.JobLabel).To(Equal(expJobLabel))
		Expect(smInstance.Spec.Endpoints).ToNot(BeEmpty())
		Expect(smInstance.Spec.Endpoints[0].Port).To(Equal(portName))

		svcURL := fmt.Sprintf("%s.openshift-logging.svc", serviceName)
		Expect(smInstance.Spec.Endpoints[0].TLSConfig.SafeTLSConfig.ServerName).
			To(Equal(svcURL))

		Expect(smInstance.Spec.Endpoints[0].BearerTokenFile).
			To(Equal("/var/run/secrets/kubernetes.io/serviceaccount/token"))

		By("verifying MetricRelabelConfigs contains only the rename rule")
		relabelConfigs := smInstance.Spec.Endpoints[0].MetricRelabelConfigs
		Expect(relabelConfigs).To(HaveLen(1))
		Expect(relabelConfigs[0].Regex).To(Equal("(.*)-(.*)"))
		Expect(relabelConfigs[0].TargetLabel).To(Equal("__name__"))
	})

	It("should reconcile a minimal profile ServiceMonitor with collector relabel configs", func() {
		minimalName := constants.MetricsCollectionProfileMinimal + "-" + serviceName
		Expect(ReconcileServiceMonitor(
			reqClient,
			constants.OpenshiftNS,
			minimalName,
			serviceName,
			owner,
			selector,
			portName,
			CollectorMinimalRelabelConfigs,
			constants.MetricsCollectionProfileMinimal,
		)).To(Succeed())

		sm := &monitoringv1.ServiceMonitor{}
		Expect(reqClient.Get(context.TODO(), types.NamespacedName{Name: minimalName, Namespace: constants.OpenshiftNS}, sm)).Should(Succeed())

		Expect(sm.Labels[constants.LabelMetricsCollectionProfile]).To(Equal(constants.MetricsCollectionProfileMinimal))

		By("verifying TLS ServerName uses the service name, not the ServiceMonitor name")
		svcURL := fmt.Sprintf("%s.openshift-logging.svc", serviceName)
		Expect(sm.Spec.Endpoints[0].TLSConfig.SafeTLSConfig.ServerName).To(Equal(svcURL))

		By("verifying MetricRelabelConfigs contains rename + keep + drop")
		relabelConfigs := sm.Spec.Endpoints[0].MetricRelabelConfigs
		Expect(relabelConfigs).To(HaveLen(3))

		Expect(relabelConfigs[0].Regex).To(Equal("(.*)-(.*)"))
		Expect(string(relabelConfigs[1].Action)).To(Equal("keep"))
		Expect(relabelConfigs[1].SourceLabels).To(Equal([]monitoringv1.LabelName{"__name__"}))
		Expect(relabelConfigs[1].Regex).To(Equal(CollectorMinimalRelabelConfigs[1].Regex))
		Expect(string(relabelConfigs[2].Action)).To(Equal("drop"))
		Expect(relabelConfigs[2].SourceLabels).To(Equal([]monitoringv1.LabelName{"component_kind", "__name__"}))
	})

	It("should update an existing ServiceMonitor on re-reconciliation", func() {
		By("creating with full profile")
		Expect(ReconcileServiceMonitor(
			reqClient,
			constants.OpenshiftNS,
			"update-test",
			"update-test",
			owner,
			selector,
			portName,
			FullRelabelConfigs,
			constants.MetricsCollectionProfileFull,
		)).To(Succeed())

		sm := &monitoringv1.ServiceMonitor{}
		key := types.NamespacedName{Name: "update-test", Namespace: constants.OpenshiftNS}
		Expect(reqClient.Get(context.TODO(), key, sm)).Should(Succeed())
		Expect(sm.Labels[constants.LabelMetricsCollectionProfile]).To(Equal(constants.MetricsCollectionProfileFull))
		Expect(sm.Spec.Endpoints[0].MetricRelabelConfigs).To(HaveLen(1))

		By("re-reconciling with minimal profile and different relabel configs")
		Expect(ReconcileServiceMonitor(
			reqClient,
			constants.OpenshiftNS,
			"update-test",
			"update-test",
			owner,
			selector,
			portName,
			CollectorMinimalRelabelConfigs,
			constants.MetricsCollectionProfileMinimal,
		)).To(Succeed())

		updated := &monitoringv1.ServiceMonitor{}
		Expect(reqClient.Get(context.TODO(), key, updated)).Should(Succeed())
		Expect(updated.Labels[constants.LabelMetricsCollectionProfile]).To(Equal(constants.MetricsCollectionProfileMinimal))
		Expect(updated.Spec.Endpoints[0].MetricRelabelConfigs).To(HaveLen(3))
	})
})
