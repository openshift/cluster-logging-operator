package collector

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	"github.com/openshift/cluster-logging-operator/internal/factory"
	"github.com/openshift/cluster-logging-operator/internal/runtime"
	"github.com/openshift/cluster-logging-operator/internal/utils"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

var _ = Describe("RemoveLegacyCollectorConfigMap", func() {
	const (
		namespace     = "openshift-logging"
		forwarderName = "lokistack"
	)

	var (
		forwarder *obs.ClusterLogForwarder
		owner     metav1.OwnerReference
	)

	BeforeEach(func() {
		forwarder = &obs.ClusterLogForwarder{
			ObjectMeta: metav1.ObjectMeta{
				Name:      forwarderName,
				Namespace: namespace,
				UID:       "clf-uid",
			},
		}
		owner = utils.AsOwner(forwarder)
	})

	It("should delete the legacy ConfigMap when owned by the forwarder", func() {
		legacy := runtime.NewConfigMap(namespace, factory.LegacyCollectorConfigMapName(forwarderName), nil)
		utils.AddOwnerRefToObject(legacy, owner)
		k8sClient := fake.NewFakeClient(legacy)

		Expect(RemoveLegacyCollectorConfigMap(k8sClient, namespace, forwarderName, owner)).To(Succeed())

		err := k8sClient.Get(context.TODO(), types.NamespacedName{Namespace: namespace, Name: legacy.Name}, &corev1.ConfigMap{})
		Expect(errors.IsNotFound(err)).To(BeTrue())
	})

	It("should not delete a ConfigMap owned by another resource", func() {
		otherOwner := metav1.OwnerReference{
			APIVersion: "loki.grafana.com/v1",
			Kind:       "LokiStack",
			Name:       forwarderName,
			UID:        "loki-uid",
			Controller: utils.GetPtr(true),
		}
		legacy := runtime.NewConfigMap(namespace, factory.LegacyCollectorConfigMapName(forwarderName), nil)
		utils.AddOwnerRefToObject(legacy, otherOwner)
		k8sClient := fake.NewFakeClient(legacy)

		Expect(RemoveLegacyCollectorConfigMap(k8sClient, namespace, forwarderName, owner)).To(Succeed())

		got := &corev1.ConfigMap{}
		Expect(k8sClient.Get(context.TODO(), types.NamespacedName{Namespace: namespace, Name: legacy.Name}, got)).To(Succeed())
		Expect(got.Name).To(Equal(legacy.Name))
	})

	It("should succeed when the legacy ConfigMap does not exist", func() {
		k8sClient := fake.NewFakeClient()
		Expect(RemoveLegacyCollectorConfigMap(k8sClient, namespace, forwarderName, owner)).To(Succeed())
	})
})
