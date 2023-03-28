package logfilemetricexporter

import (
	"context"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	loggingv1 "github.com/openshift/cluster-logging-operator/apis/logging/v1"
	loggingv1alpha1 "github.com/openshift/cluster-logging-operator/apis/logging/v1alpha1"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	"github.com/openshift/cluster-logging-operator/internal/utils"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

var _ = Describe("Reconcile LogFileMetricExporter Daemonset", func() {

	defer GinkgoRecover()

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

		lfmeInstance = &loggingv1alpha1.LogFileMetricExporter{
			ObjectMeta: metav1.ObjectMeta{
				Name:      constants.SingletonName,
				Namespace: constants.WatchNamespace,
			},
		}

		dsOwner = utils.AsOwner(cluster)

		dsKey         = types.NamespacedName{Name: constants.LogfilesmetricexporterName, Namespace: cluster.GetNamespace()}
		dsInstance    = &appsv1.DaemonSet{}
		requestMemory = resource.MustParse("100Gi")
		requestCPU    = resource.MustParse("500m")
	)

	It("should reconcile successfully a daemonset when no specs are specified", func() {

		// Reconcile the exporter daemonset
		Expect(ReconcileDaemonset(*lfmeInstance,
			recorder,
			reqClient,
			constants.WatchNamespace,
			constants.LogfilesmetricexporterName, cluster.Spec.Collection.Type, dsOwner)).To(Succeed())

		// Check if daemonset is available
		Expect(reqClient.Get(context.TODO(), dsKey, dsInstance)).Should(Succeed())
		Expect(dsInstance.Spec.Template.Spec.Containers).To(HaveLen(1))
		Expect(dsInstance.Spec.Template.Spec.Containers[0].Resources.Limits).To(BeNil())
		Expect(dsInstance.Spec.Template.Spec.Containers[0].Resources.Requests).To(BeNil())
	})

	It("should reconcile successfully a daemonset with specificied resources.requests", func() {
		lfmeInstance.Spec = loggingv1alpha1.LogFileMetricExporterSpec{
			Resources: &corev1.ResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceCPU:    requestCPU,
					corev1.ResourceMemory: requestMemory,
				},
			},
		}

		// Reconcile the exporter daemonset
		Expect(ReconcileDaemonset(*lfmeInstance,
			recorder,
			reqClient,
			constants.WatchNamespace,
			constants.LogfilesmetricexporterName, cluster.Spec.Collection.Type, dsOwner)).To(Succeed())

		// Get and check the daemonset
		Expect(reqClient.Get(context.TODO(), dsKey, dsInstance)).Should(Succeed())
		Expect(dsInstance.Spec.Template.Spec.Containers).To(HaveLen(1))
		Expect(dsInstance.Spec.Template.Spec.Containers[0].Resources.Limits).To(BeNil())
		Expect(dsInstance.Spec.Template.Spec.Containers[0].Resources.Requests).To(Not(BeNil()))

		Expect(dsInstance.Spec.Template.Spec.Containers[0].Resources.Requests.Cpu().Cmp(requestCPU)).To(Equal(0))
		Expect(dsInstance.Spec.Template.Spec.Containers[0].Resources.Requests.Memory().Cmp(requestMemory)).To(Equal(0))

	})

	It("should reconcile successfully a daemonset with specificied resources.limits", func() {
		lfmeInstance.Spec = loggingv1alpha1.LogFileMetricExporterSpec{
			Resources: &corev1.ResourceRequirements{
				Limits: corev1.ResourceList{
					corev1.ResourceCPU:    requestCPU,
					corev1.ResourceMemory: requestMemory,
				},
			},
		}

		// Reconcile the exporter daemonset
		Expect(ReconcileDaemonset(*lfmeInstance,
			recorder,
			reqClient,
			constants.WatchNamespace,
			constants.LogfilesmetricexporterName, cluster.Spec.Collection.Type, dsOwner)).To(Succeed())

		// Get and check the daemonset
		Expect(reqClient.Get(context.TODO(), dsKey, dsInstance)).Should(Succeed())
		Expect(dsInstance.Spec.Template.Spec.Containers).To(HaveLen(1))
		Expect(dsInstance.Spec.Template.Spec.Containers[0].Resources.Requests).To(BeNil())
		Expect(dsInstance.Spec.Template.Spec.Containers[0].Resources.Limits).To(Not(BeNil()))

		Expect(dsInstance.Spec.Template.Spec.Containers[0].Resources.Limits.Cpu().Cmp(requestCPU)).To(Equal(0))
		Expect(dsInstance.Spec.Template.Spec.Containers[0].Resources.Limits.Memory().Cmp(requestMemory)).To(Equal(0))

	})

	It("should reconcile successfully a daemonset with specificied resources.limits and resources.requests", func() {

		reqMem1 := resource.MustParse("10Gi")
		reqCPU1 := resource.MustParse("100m")

		lfmeInstance.Spec = loggingv1alpha1.LogFileMetricExporterSpec{
			Resources: &corev1.ResourceRequirements{
				Limits: corev1.ResourceList{
					corev1.ResourceCPU:    requestCPU,
					corev1.ResourceMemory: requestMemory,
				},
				Requests: corev1.ResourceList{
					corev1.ResourceCPU:    reqCPU1,
					corev1.ResourceMemory: reqMem1,
				},
			},
		}

		// Reconcile the exporter daemonset
		Expect(ReconcileDaemonset(*lfmeInstance,
			recorder,
			reqClient,
			constants.WatchNamespace,
			constants.LogfilesmetricexporterName, cluster.Spec.Collection.Type, dsOwner)).To(Succeed())

		// Get and check the daemonset
		Expect(reqClient.Get(context.TODO(), dsKey, dsInstance)).Should(Succeed())
		Expect(dsInstance.Spec.Template.Spec.Containers).To(HaveLen(1))
		Expect(dsInstance.Spec.Template.Spec.Containers[0].Resources.Requests).To(Not(BeNil()))
		Expect(dsInstance.Spec.Template.Spec.Containers[0].Resources.Limits).To(Not(BeNil()))

		// Check resource limits
		Expect(dsInstance.Spec.Template.Spec.Containers[0].Resources.Limits.Cpu().Cmp(requestCPU)).To(Equal(0))
		Expect(dsInstance.Spec.Template.Spec.Containers[0].Resources.Limits.Memory().Cmp(requestMemory)).To(Equal(0))

		// Check resource requests
		Expect(dsInstance.Spec.Template.Spec.Containers[0].Resources.Requests.Cpu().Cmp(reqCPU1)).To(Equal(0))
		Expect(dsInstance.Spec.Template.Spec.Containers[0].Resources.Requests.Memory().Cmp(reqMem1)).To(Equal(0))
	})
})
