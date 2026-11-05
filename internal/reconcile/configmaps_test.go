package reconcile_test

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/openshift/cluster-logging-operator/internal/reconcile"
	"github.com/openshift/cluster-logging-operator/internal/runtime"
	"github.com/openshift/cluster-logging-operator/internal/utils"
	"github.com/openshift/cluster-logging-operator/internal/utils/comparators"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8sruntime "k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

var _ = Describe("reconciling ConfigMap", func() {

	const (
		namespace = "openshift-logging"
		name      = "lokistack-config"
	)

	var (
		clfOwner = metav1.OwnerReference{
			APIVersion: "observability.openshift.io/v1",
			Kind:       "ClusterLogForwarder",
			Name:       "lokistack",
			UID:        "clf-uid",
			Controller: utils.GetPtr(true),
		}
		lokiOwner = metav1.OwnerReference{
			APIVersion: "loki.grafana.com/v1",
			Kind:       "LokiStack",
			Name:       "lokistack",
			UID:        "loki-uid",
			Controller: utils.GetPtr(true),
		}
	)

	newClient := func(objs ...client.Object) client.Client {
		globalScheme := k8sruntime.NewScheme()
		Expect(scheme.AddToScheme(globalScheme)).To(Succeed())
		return fake.NewClientBuilder().WithScheme(globalScheme).WithObjects(objs...).Build()
	}

	getConfigMap := func(k8sClient client.Client) *corev1.ConfigMap {
		result := &corev1.ConfigMap{}
		Expect(k8sClient.Get(context.TODO(), client.ObjectKey{Namespace: namespace, Name: name}, result)).To(Succeed())
		return result
	}

	desiredCollectorCM := func() *corev1.ConfigMap {
		cm := runtime.NewConfigMap(namespace, name, map[string]string{
			"vector.toml": "data_dir = \"/var/lib/vector\"",
		})
		utils.AddOwnerRefToObject(cm, clfOwner)
		return cm
	}

	It("should create the ConfigMap when it does not exist", func() {
		k8sClient := newClient()
		desired := desiredCollectorCM()

		Expect(reconcile.Configmap(k8sClient, k8sClient, desired, comparators.CompareLabels)).To(Succeed())

		result := getConfigMap(k8sClient)
		Expect(result.Data).To(HaveKey("vector.toml"))
		Expect(result.OwnerReferences).To(Equal([]metav1.OwnerReference{clfOwner}))
	})

	It("should update the ConfigMap when owned by the ClusterLogForwarder", func() {
		existing := runtime.NewConfigMap(namespace, name, map[string]string{
			"vector.toml": "old",
		})
		utils.AddOwnerRefToObject(existing, clfOwner)
		k8sClient := newClient(existing)

		desired := desiredCollectorCM()
		Expect(reconcile.Configmap(k8sClient, k8sClient, desired, comparators.CompareLabels)).To(Succeed())

		result := getConfigMap(k8sClient)
		Expect(result.Data["vector.toml"]).To(Equal("data_dir = \"/var/lib/vector\""))
		Expect(result.OwnerReferences).To(Equal([]metav1.OwnerReference{clfOwner}))
	})

	It("should refuse to overwrite a ConfigMap owned by another resource", func() {
		existing := runtime.NewConfigMap(namespace, name, map[string]string{
			"config.yaml": "auth_enabled: false",
		})
		utils.AddOwnerRefToObject(existing, lokiOwner)
		k8sClient := newClient(existing)

		desired := desiredCollectorCM()
		err := reconcile.Configmap(k8sClient, k8sClient, desired, comparators.CompareLabels)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("refusing to overwrite"))

		result := getConfigMap(k8sClient)
		Expect(result.Data).To(HaveKey("config.yaml"))
		Expect(result.Data).NotTo(HaveKey("vector.toml"))
		Expect(result.OwnerReferences).To(Equal([]metav1.OwnerReference{lokiOwner}))
	})

	It("should refuse an identical ConfigMap owned by another resource", func() {
		existing := runtime.NewConfigMap(namespace, name, map[string]string{
			"vector.toml": "data_dir = \"/var/lib/vector\"",
		})
		utils.AddOwnerRefToObject(existing, lokiOwner)
		k8sClient := newClient(existing)

		err := reconcile.Configmap(k8sClient, k8sClient, desiredCollectorCM(), comparators.CompareLabels)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("refusing to overwrite"))
	})

	It("should refuse to take ownership of an existing ConfigMap with no owner", func() {
		existing := runtime.NewConfigMap(namespace, name, map[string]string{
			"config.yaml": "auth_enabled: false",
		})
		k8sClient := newClient(existing)

		desired := desiredCollectorCM()
		err := reconcile.Configmap(k8sClient, k8sClient, desired, comparators.CompareLabels)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("refusing to overwrite"))

		result := getConfigMap(k8sClient)
		Expect(result.Data).To(HaveKey("config.yaml"))
		Expect(result.OwnerReferences).To(BeEmpty())
	})
})
