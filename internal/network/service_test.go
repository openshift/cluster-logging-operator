package network

import (
	"context"
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	"github.com/openshift/cluster-logging-operator/internal/runtime"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

var _ = Describe("Reconcile Service", func() {

	defer GinkgoRecover()

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
		portName      = "test-port"
		port          = int32(1337)
		certSecret    = "test-secret"
		componentName = "some-component"
		serviceName   = "test-service"

		commonLabels = func(o runtime.Object) {
			runtime.SetCommonLabels(o, "vector", "test", componentName)
		}

		owner = metav1.OwnerReference{}

		serviceKey      = types.NamespacedName{Name: serviceName, Namespace: namespace.Name}
		serviceInstance = &corev1.Service{}
	)

	It("should successfully reconcile the service", func() {
		// Reconcile the service
		Expect(ReconcileService(
			reqClient,
			constants.OpenshiftNS,
			serviceName,
			serviceName,
			componentName,
			portName,
			certSecret,
			port,
			owner,
			commonLabels)).To(Succeed())

		// Get and check the service
		Expect(reqClient.Get(context.TODO(), serviceKey, serviceInstance)).Should(Succeed())

		Expect(serviceInstance.Name).To(Equal(serviceName))
		Expect(serviceInstance.Spec.Ports).ToNot(BeEmpty(), "Exp. to have spec.Ports")

		Expect(serviceInstance.Spec.Ports[0].Port).
			To(Equal(port), fmt.Sprintf("Exp service port of: %v", port))

		Expect(serviceInstance.Annotations[constants.AnnotationServingCertSecretName]).
			To(Equal(certSecret))
	})

})
