package clusterlogforwarder

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	loggingv1 "github.com/openshift/cluster-logging-operator/api/logging/v1"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	"github.com/openshift/cluster-logging-operator/internal/runtime"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

var _ = Describe("ValidateForwarderForConversion", func() {
	var (
		clf       *loggingv1.ClusterLogForwarder
		k8sClient client.Client
		esSecret  = &corev1.Secret{
			ObjectMeta: v1.ObjectMeta{
				Name:      "es-secret",
				Namespace: constants.OpenshiftNS,
			},
			Data: map[string][]byte{"user": []byte("username")},
		}
	)
	BeforeEach(func() {
		k8sClient = fake.NewClientBuilder().WithObjects(esSecret).Build()
		clf = runtime.NewClusterLogForwarder(constants.OpenshiftNS, constants.SingletonName)
	})
	It("should return no secrets, status, or errors if clusterlogforwarder is nil", func() {
		secrets, status, err := ValidateClusterLogForwarderForConversion(nil, k8sClient)
		Expect(secrets).To(BeNil())
		Expect(status).To(BeNil())
		Expect(err).To(BeNil())
	})
	It("should return error if clf references fluentDForward as an output", func() {
		clf.Spec.Outputs = []loggingv1.OutputSpec{
			{
				Name: "fluentdFor",
				Type: loggingv1.OutputTypeFluentdForward,
			},
		}
		secrets, status, err := ValidateClusterLogForwarderForConversion(clf, k8sClient)

		Expect(secrets).To(BeNil())
		cond := status.Conditions[0]
		Expect(cond.Message).To(Equal("cannot migrate CLF because fluentDForward is referenced as an output."))
		Expect(err).ToNot(BeNil())
	})
	It("should return error if output has missing secrets", func() {
		clf.Spec.Outputs = []loggingv1.OutputSpec{
			{
				Name: "es",
				Type: loggingv1.OutputTypeElasticsearch,
				Secret: &loggingv1.OutputSecretSpec{
					Name: "missing",
				},
			},
		}
		secrets, status, err := ValidateClusterLogForwarderForConversion(clf, k8sClient)

		Expect(secrets).To(BeNil())
		cond := status.Conditions[0]
		Expect(cond.Message).To(Equal("outputs have defined secrets that are missing"))
		Expect(err).ToNot(BeNil())
	})
	It("should return map of secrets if valid CLF", func() {
		clf.Spec.Outputs = []loggingv1.OutputSpec{
			{
				Name: "es",
				Type: loggingv1.OutputTypeElasticsearch,
				Secret: &loggingv1.OutputSecretSpec{
					Name: esSecret.Name,
				},
			},
		}
		secrets, status, err := ValidateClusterLogForwarderForConversion(clf, k8sClient)

		Expect(err).To(BeNil())
		Expect(status.Conditions).To(BeNil())
		Expect(secrets).To(ContainElement(esSecret))
	})
})
