package reconcile_test

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/openshift/cluster-logging-operator/internal/reconcile"
	"github.com/openshift/cluster-logging-operator/internal/runtime"
	"github.com/openshift/cluster-logging-operator/internal/utils"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8sruntime "k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

var _ = Describe("reconciling DaemonSet ownership", func() {
	const (
		namespace = "openshift-logging"
		name      = "lokistack"
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
		Expect(appsv1.AddToScheme(globalScheme)).To(Succeed())
		return fake.NewClientBuilder().WithScheme(globalScheme).WithObjects(objs...).Build()
	}

	desiredDS := func() *appsv1.DaemonSet {
		ds := runtime.NewDaemonSet(namespace, name)
		ds.Spec.Selector = &metav1.LabelSelector{MatchLabels: map[string]string{"app": "collector"}}
		ds.Spec.Template.Labels = map[string]string{"app": "collector"}
		utils.AddOwnerRefToObject(ds, clfOwner)
		return ds
	}

	It("should refuse to overwrite a DaemonSet owned by another resource", func() {
		existing := runtime.NewDaemonSet(namespace, name)
		existing.Spec.Selector = &metav1.LabelSelector{MatchLabels: map[string]string{"app": "loki"}}
		existing.Spec.Template.Labels = map[string]string{"app": "loki"}
		utils.AddOwnerRefToObject(existing, lokiOwner)
		k8sClient := newClient(existing)

		err := reconcile.DaemonSet(k8sClient, desiredDS())
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("refusing to overwrite"))

		got := &appsv1.DaemonSet{}
		Expect(k8sClient.Get(context.TODO(), client.ObjectKey{Namespace: namespace, Name: name}, got)).To(Succeed())
		Expect(got.Spec.Selector.MatchLabels["app"]).To(Equal("loki"))
		Expect(got.OwnerReferences).To(Equal([]metav1.OwnerReference{lokiOwner}))
	})
})
