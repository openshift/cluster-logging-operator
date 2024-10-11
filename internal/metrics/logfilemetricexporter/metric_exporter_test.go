package logfilemetricexporter

import (
	"context"
	"fmt"

	. "github.com/onsi/ginkgo"
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

		reqClient = fake.NewFakeClient(
			namespace,
		)

		lfmeInstance = &loggingv1alpha1.LogFileMetricExporter{}

		// Daemonset
		dsKey      = types.NamespacedName{Name: constants.LogfilesmetricexporterName, Namespace: namespace.Name}
		dsInstance = &appsv1.DaemonSet{}
		reqMem1    = resource.MustParse("50Gi")
		reqCPU1    = resource.MustParse("300m")
		reqMem2    = resource.MustParse("15Gi")
		reqCPU2    = resource.MustParse("100m")

		// Service
		serviceKey      = types.NamespacedName{Name: constants.LogfilesmetricexporterName, Namespace: namespace.Name}
		serviceInstance = &corev1.Service{}

		// Service Monitor
		serviceMonitorKey = types.NamespacedName{Name: constants.LogfilesmetricexporterName, Namespace: namespace.Name}
		smInstance        = &monitoringv1.ServiceMonitor{}
	)

	It("Should reconcile successfully a daemonset, service, and service monitor", func() {

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

		// Reconcile the LogFileMetricExporter
		Expect(Reconcile(lfmeInstance, reqClient, utils.AsOwner(lfmeInstance))).To(Succeed())

		// Daemonset
		// Get and check the daemonset
		Expect(reqClient.Get(context.TODO(), dsKey, dsInstance)).Should(Succeed())
		Expect(dsInstance.Spec.Template.Spec.Containers).To(HaveLen(1))
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
		Expect(smInstance.Spec.Endpoints[0].Port).To(Equal(exporterPortName))

		svcURL := fmt.Sprintf("%s.openshift-logging.svc", constants.LogfilesmetricexporterName)
		Expect(smInstance.Spec.Endpoints[0].TLSConfig.SafeTLSConfig.ServerName).
			To(Equal(svcURL))
	})
})
