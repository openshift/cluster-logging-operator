package metrics

import (
	"context"
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	loggingv1 "github.com/openshift/cluster-logging-operator/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	"github.com/openshift/cluster-logging-operator/internal/utils"
	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

var _ = Describe("Reconcile ServiceMonitor", func() {

	defer GinkgoRecover()

	_ = monitoringv1.AddToScheme(scheme.Scheme)

	var (
		cluster = &loggingv1.ClusterLogging{
			ObjectMeta: metav1.ObjectMeta{
				Name:      constants.SingletonName,
				Namespace: constants.OpenshiftNS,
			},
			Spec: loggingv1.ClusterLoggingSpec{
				ManagementState: loggingv1.ManagementStateManaged,
				LogStore: &loggingv1.LogStoreSpec{
					Type: loggingv1.LogStoreTypeElasticsearch,
				},
				Collection: &loggingv1.CollectionSpec{
					Type:          loggingv1.LogCollectionTypeFluentd,
					CollectorSpec: loggingv1.CollectorSpec{},
				},
			},
		}

		// Adding ns and label to account for addSecurityLabelsToNamespace() added in LOG-2620
		namespace = &corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Labels: map[string]string{"test": "true"},
				Name:   cluster.Namespace,
			},
		}

		reqClient = fake.NewFakeClient( //nolint
			cluster,
			namespace,
		)
		recorder = record.NewFakeRecorder(100)
		owner    = utils.AsOwner(cluster)
		portName = "test-port"

		serviceMonitorKey = types.NamespacedName{Name: constants.LogfilesmetricexporterName, Namespace: cluster.GetNamespace()}
		smInstance        = &monitoringv1.ServiceMonitor{}
	)

	It("should successfully reconcile the ServiceMonitor for LogFileMetricExporter", func() {
		// Reconcile the exporter daemonset
		Expect(ReconcileServiceMonitor(recorder,
			reqClient,
			cluster.GetNamespace(),
			constants.LogfilesmetricexporterName,
			portName,
			owner)).To(Succeed())

		// Get and check the ServiceMonitor
		Expect(reqClient.Get(context.TODO(), serviceMonitorKey, smInstance)).Should(Succeed())

		Expect(smInstance.Name).To(Equal(constants.LogfilesmetricexporterName))

		expJobLabel := fmt.Sprintf("monitor-%s", constants.LogfilesmetricexporterName)
		Expect(smInstance.Spec.JobLabel).To(Equal(expJobLabel))
		Expect(smInstance.Spec.Endpoints).ToNot(BeEmpty())
		Expect(smInstance.Spec.Endpoints[0].Port).To(Equal(portName))

		svcURL := fmt.Sprintf("%s.openshift-logging.svc", constants.LogfilesmetricexporterName)
		Expect(smInstance.Spec.Endpoints[0].TLSConfig.SafeTLSConfig.ServerName).
			To(Equal(svcURL))
	})
})
