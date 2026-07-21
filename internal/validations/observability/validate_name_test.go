package observability

import (
	internalcontext "github.com/openshift/cluster-logging-operator/internal/api/context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	. "github.com/openshift/cluster-logging-operator/test/matchers"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

var _ = Describe("[internal][validations] validate clusterlogforwarder name", func() {
	const (
		testNamespace = "openshift-logging"
		testName      = "lokistack"
	)

	var (
		clf       *obs.ClusterLogForwarder
		lokiStack *unstructured.Unstructured
	)

	BeforeEach(func() {
		clf = &obs.ClusterLogForwarder{
			ObjectMeta: metav1.ObjectMeta{
				Name:      testName,
				Namespace: testNamespace,
			},
		}
		lokiStack = &unstructured.Unstructured{}
		lokiStack.SetGroupVersionKind(lokiStackGVK)
		lokiStack.SetName(testName)
		lokiStack.SetNamespace(testNamespace)
	})

	Context("#validateName", func() {
		It("should pass validation when no LokiStack exists with the same name", func() {
			k8sClient := fake.NewClientBuilder().Build()
			validateName(internalcontext.ForwarderContext{
				Client:    k8sClient,
				Reader:    k8sClient,
				Forwarder: clf,
			})
			Expect(clf.Status.Conditions).To(BeEmpty())
		})

		It("should fail validation when a LokiStack exists with the same name and namespace", func() {
			k8sClient := fake.NewClientBuilder().WithObjects(lokiStack).Build()
			validateName(internalcontext.ForwarderContext{
				Client:    k8sClient,
				Reader:    k8sClient,
				Forwarder: clf,
			})
			Expect(clf.Status.Conditions).To(HaveCondition(
				obs.ConditionTypeName,
				false,
				obs.ReasonNameConflict,
				`.*conflicts with LokiStack "lokistack".*both use ConfigMap "lokistack-config".*`,
			))
		})

		It("should pass validation when a LokiStack exists with the same name in a different namespace", func() {
			lokiStack.SetNamespace("openshift-logging-storage")
			k8sClient := fake.NewClientBuilder().WithObjects(lokiStack).Build()
			validateName(internalcontext.ForwarderContext{
				Client:    k8sClient,
				Reader:    k8sClient,
				Forwarder: clf,
			})
			Expect(clf.Status.Conditions).To(BeEmpty())
		})

		It("should pass validation when a LokiStack exists with a different name in the same namespace", func() {
			lokiStack.SetName("other-lokistack")
			k8sClient := fake.NewClientBuilder().WithObjects(lokiStack).Build()
			validateName(internalcontext.ForwarderContext{
				Client:    k8sClient,
				Reader:    k8sClient,
				Forwarder: clf,
			})
			Expect(clf.Status.Conditions).To(BeEmpty())
		})
	})
})
