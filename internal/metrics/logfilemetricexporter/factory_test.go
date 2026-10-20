package logfilemetricexporter

import (
	"context"
	"fmt"

	"sigs.k8s.io/controller-runtime/pkg/client"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	securityv1 "github.com/openshift/api/security/v1"
	loggingv1alpha1 "github.com/openshift/cluster-logging-operator/api/logging/v1alpha1"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	"github.com/openshift/cluster-logging-operator/internal/runtime"
	"github.com/openshift/cluster-logging-operator/internal/utils"
	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

var _ = Describe("Reconcile LogFileMetricExporter", func() {

	defer GinkgoRecover()
	_ = monitoringv1.AddToScheme(scheme.Scheme)
	_ = securityv1.AddToScheme(scheme.Scheme)

	var (
		// Adding ns and label to account for addSecurityLabelsToNamespace() added in LOG-2620
		namespace = &corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Labels: map[string]string{"test": "true"},
				Name:   constants.OpenshiftNS,
			},
		}

		reqClient = fake.NewFakeClient(namespace)
		reader    = reqClient.(client.Reader)

		lfmeInstance *loggingv1alpha1.LogFileMetricExporter

		// Daemonset
		dsKey      = types.NamespacedName{Name: constants.LogfilesmetricexporterName, Namespace: namespace.Name}
		dsInstance *appsv1.DaemonSet
		reqMem1    = resource.MustParse("50Gi")
		reqCPU1    = resource.MustParse("300m")
		reqMem2    = resource.MustParse("15Gi")
		reqCPU2    = resource.MustParse("100m")

		// Service
		serviceKey      = types.NamespacedName{Name: constants.LogfilesmetricexporterName, Namespace: namespace.Name}
		serviceInstance *corev1.Service

		// Service Monitor
		serviceMonitorKey = types.NamespacedName{Name: constants.LogfilesmetricexporterName, Namespace: namespace.Name}
		smInstance        *monitoringv1.ServiceMonitor
	)

	BeforeEach(func() {
		// Reset instances before each test
		lfmeInstance = &loggingv1alpha1.LogFileMetricExporter{}
		dsInstance = &appsv1.DaemonSet{}
		serviceInstance = &corev1.Service{}
		smInstance = &monitoringv1.ServiceMonitor{}

		runtime.Initialize(lfmeInstance, constants.OpenshiftNS, constants.SingletonName)

		lfmeInstance.Spec = loggingv1alpha1.LogFileMetricExporterSpec{
			Resources: &corev1.ResourceRequirements{
				Limits: corev1.ResourceList{
					corev1.ResourceCPU:    reqCPU2,
					corev1.ResourceMemory: reqMem2,
				},
				Requests: corev1.ResourceList{
					corev1.ResourceCPU:    reqCPU1,
					corev1.ResourceMemory: reqMem1,
				},
			},
		}
	})

	Context("when OpenShift version is 4.21 or older", func() {
		DescribeTable("SecurityContext configuration based on OpenShift version",
			func(version string, expectedUser int64, expectedNonRoot bool) {
				Expect(Reconcile(lfmeInstance, reqClient, reader, utils.AsOwner(lfmeInstance), version)).To(Succeed())

				Expect(reqClient.Get(context.TODO(), dsKey, dsInstance)).Should(Succeed())
				Expect(dsInstance.Spec.Template.Spec.Containers).To(HaveLen(1))

				sc := dsInstance.Spec.Template.Spec.Containers[0].SecurityContext
				Expect(sc).To(Not(BeNil()))
				Expect(*sc.RunAsUser).To(Equal(expectedUser))
				Expect(*sc.RunAsNonRoot).To(Equal(expectedNonRoot))
			},
			Entry("OpenShift 4.19 (older, needs root workaround)", "4.19.0", int64(0), false),
			Entry("OpenShift 4.20 (older, needs root workaround) with pre-release tags", "4.20.0-alfa123", int64(0), false),
			Entry("OpenShift 4.22 (newer, secure non-root)", "4.22.2", int64(1000), true),
			Entry("OpenShift 5.0 (newer, secure non-root)", "5.0.2", int64(1000), true),
			Entry("Invalid version string (fallback to secure defaults)", "invalid-version", int64(1000), true),
		)
	})

	Context("when OpenShift version is 4.22 or newer", func() {
		It("Should reconcile successfully and use secure non-root defaults", func() {
			Expect(Reconcile(lfmeInstance, reqClient, reader, utils.AsOwner(lfmeInstance), "4.22.0")).To(Succeed())

			// Daemonset
			// Get and check the daemonset
			Expect(reqClient.Get(context.TODO(), dsKey, dsInstance)).Should(Succeed())
			Expect(dsInstance.Spec.Template.Spec.Containers).To(HaveLen(1))
			Expect(dsInstance.Spec.Template.Spec.Containers[0].Resources.Requests).To(Not(BeNil()))
			Expect(dsInstance.Spec.Template.Spec.Containers[0].Resources.Limits).To(Not(BeNil()))
		// Daemonset
		// Get and check the daemonset
		Expect(reqClient.Get(context.TODO(), dsKey, dsInstance)).Should(Succeed())
		Expect(dsInstance.Spec.Template.Spec.Containers).To(HaveLen(1))
		Expect(dsInstance.Spec.Template.Spec.Containers[0].Args).To(HaveLen(2))
		Expect(dsInstance.Spec.Template.Spec.Containers[0].Args[1]).To(ContainSubstring("-secureMetrics=false"))
		Expect(dsInstance.Spec.Template.Spec.Containers[0].Resources.Requests).To(Not(BeNil()))
		Expect(dsInstance.Spec.Template.Spec.Containers[0].Resources.Limits).To(Not(BeNil()))

			// Check resource limits
			Expect(dsInstance.Spec.Template.Spec.Containers[0].Resources.Limits.Cpu().Cmp(reqCPU2)).To(Equal(0))
			Expect(dsInstance.Spec.Template.Spec.Containers[0].Resources.Limits.Memory().Cmp(reqMem2)).To(Equal(0))

			// Check request limits
			Expect(dsInstance.Spec.Template.Spec.Containers[0].Resources.Requests.Cpu().Cmp(reqCPU1)).To(Equal(0))
			Expect(dsInstance.Spec.Template.Spec.Containers[0].Resources.Requests.Memory().Cmp(reqMem1)).To(Equal(0))

			// Service
			// Get and check the service
			Expect(reqClient.Get(context.TODO(), serviceKey, serviceInstance)).Should(Succeed())

			Expect(serviceInstance.Name).To(Equal(constants.LogfilesmetricexporterName))
			Expect(serviceInstance.Spec.Ports).ToNot(BeEmpty(), "Exp. to have spec.Ports")

			Expect(serviceInstance.Spec.Ports[0].Port).
				To(Equal(exporterPort), fmt.Sprintf("Exp service port of: %v", exporterPort))

			Expect(serviceInstance.Annotations[constants.AnnotationServingCertSecretName]).
				To(Equal(ExporterMetricsSecretName))

			// ServiceMonitor
			// Get and check the ServiceMonitor
			Expect(reqClient.Get(context.TODO(), serviceMonitorKey, smInstance)).Should(Succeed())

			Expect(smInstance.Name).To(Equal(constants.LogfilesmetricexporterName))

			expJobLabel := fmt.Sprintf("monitor-%s", constants.LogfilesmetricexporterName)
			Expect(smInstance.Spec.JobLabel).To(Equal(expJobLabel))
			Expect(smInstance.Spec.Endpoints).ToNot(BeEmpty())
			Expect(smInstance.Spec.Endpoints[0].Port).To(Equal(constants.MetricsPortName))

			svcURL := fmt.Sprintf("%s.openshift-logging.svc", constants.LogfilesmetricexporterName)
			Expect(smInstance.Spec.Endpoints[0].TLSConfig.SafeTLSConfig.ServerName).
				To(Equal(svcURL))
		})
	})
})
