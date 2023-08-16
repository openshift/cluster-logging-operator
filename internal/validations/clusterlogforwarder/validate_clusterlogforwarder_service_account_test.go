package clusterlogforwarder

import (
	"context"
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	loggingv1 "github.com/openshift/cluster-logging-operator/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	authorizationapi "k8s.io/api/authorization/v1"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

var _ = Describe("[internal][validations] validate clusterlogforwarder permissions", func() {
	var (
		k8sClient      client.Client
		customClf      loggingv1.ClusterLogForwarder
		extras         map[string]bool
		providedSAName = "test-serviceAccount"
	)

	BeforeEach(func() {
		customClf = loggingv1.ClusterLogForwarder{
			ObjectMeta: v1.ObjectMeta{
				Name:      "test-clf",
				Namespace: constants.OpenshiftNS,
			},
			Spec: loggingv1.ClusterLogForwarderSpec{
				ServiceAccountName: providedSAName,
			},
		}
		extras = map[string]bool{}
	})

	Context("service account existence", func() {

		It("should fail when no service account found", func() {
			customClf.Spec.ServiceAccountName = providedSAName
			k8sClient = fake.NewClientBuilder().Build()
			err, _ := ValidateServiceAccount(customClf, k8sClient, extras)
			Expect(err).To(MatchError(MatchRegexp("service account not found: .+")))
		})

		It("should succeed when service account is found", func() {

			clfServiceAccount := &corev1.ServiceAccount{
				ObjectMeta: v1.ObjectMeta{
					Name:      providedSAName,
					Namespace: customClf.Namespace,
				},
			}
			k8sClient = fake.NewClientBuilder().WithObjects(clfServiceAccount).Build()
			//customClf.Spec.ServiceAccountName = providedSAName

			serviceAccount, err := getServiceAccount(customClf.Spec.ServiceAccountName, customClf.Namespace, k8sClient)
			Expect(serviceAccount).ToNot(BeNil())
			Expect(err).To(BeNil())
		})

		It("should return an error if custom clusterlogforwarder does not include a service account name", func() {
			customClf.Spec.ServiceAccountName = ""
			Expect(ValidateServiceAccount(customClf, k8sClient, extras)).ToNot(Succeed())
		})
	})

	Context("when evaluating inputs", func() {
		var (
			clfServiceAccount = &corev1.ServiceAccount{
				ObjectMeta: v1.ObjectMeta{
					Name:      "test-serviceAccount",
					Namespace: constants.OpenshiftNS,
				},
			}
		)
		BeforeEach(func() {
			k8sClient = &mockSARClient{
				fake.NewClientBuilder().WithObjects(clfServiceAccount).Build(),
			}
		})
		It("should pass validation for application logs", func() {
			inputName := "some-custom-namespace"
			customClf.Spec = loggingv1.ClusterLogForwarderSpec{
				ServiceAccountName: clfServiceAccount.Name,
				Pipelines: []loggingv1.PipelineSpec{
					{
						Name: "pipeline1",
						InputRefs: []string{
							loggingv1.InputNameApplication,
							inputName,
						},
					},
				},
			}
			Expect(ValidateServiceAccount(customClf, k8sClient, extras)).To(Succeed())
		})
		It("should pass validation when service account can collect specified inputs", func() {
			inputName := "some-custom-namespace"
			customClf.Spec = loggingv1.ClusterLogForwarderSpec{
				ServiceAccountName: clfServiceAccount.Name,
				Inputs: []loggingv1.InputSpec{
					{
						Name: "my-custom-input",
						Application: &loggingv1.Application{
							Namespaces: []string{inputName},
						},
					},
				},
				Pipelines: []loggingv1.PipelineSpec{
					{
						Name: "pipeline1",
						InputRefs: []string{
							loggingv1.InputNameApplication,
							inputName,
						},
					},
					{
						Name: "pipeline2",
						InputRefs: []string{
							loggingv1.InputNameApplication,
						},
					},
				},
			}
			Expect(ValidateServiceAccount(customClf, k8sClient, extras)).To(Succeed())
		})

		It("should return validation error if service account cannot collect specified inputs", func() {
			customClf.Spec = loggingv1.ClusterLogForwarderSpec{
				ServiceAccountName: clfServiceAccount.Name,

				Pipelines: []loggingv1.PipelineSpec{
					{
						Name: "pipeline1",
						InputRefs: []string{
							loggingv1.InputNameApplication,
							loggingv1.InputNameAudit,
						},
					},
					{
						Name: "pipeline2",
						InputRefs: []string{
							loggingv1.InputNameApplication,
						},
					},
				},
			}
			Expect(ValidateServiceAccount(customClf, k8sClient, extras)).To(Not(Succeed()))
		})
	})

	It("should not validate clusterlogforwarder named 'instance' in the namespace 'openshift-logging'", func() {
		singletonClf := loggingv1.ClusterLogForwarder{
			ObjectMeta: v1.ObjectMeta{
				Name:      constants.SingletonName,
				Namespace: constants.OpenshiftNS,
			},
			Spec: loggingv1.ClusterLogForwarderSpec{
				ServiceAccountName: "test-sa",
			},
		}
		Expect(ValidateServiceAccount(singletonClf, k8sClient, extras)).To(Succeed())
	})

	It("should fail validation when namespace is openshift-logging and SA name is 'logcollector' because this is reserved for the legacy usecase", func() {
		singletonClf := loggingv1.ClusterLogForwarder{
			ObjectMeta: v1.ObjectMeta{
				Name:      "anything",
				Namespace: constants.OpenshiftNS,
			},
			Spec: loggingv1.ClusterLogForwarderSpec{
				ServiceAccountName: constants.CollectorServiceAccountName,
			},
		}
		Expect(ValidateServiceAccount(singletonClf, k8sClient, extras)).To(MatchError(MatchRegexp("reserved serviceaccount")))
	})

})

// Mocking a subject access review
type mockSARClient struct {
	client.Client
}

func (c *mockSARClient) Create(ctx context.Context, obj client.Object, opts ...client.CreateOption) error {
	sar, ok := obj.(*authorizationapi.SubjectAccessReview)
	if !ok {
		return fmt.Errorf("unexpected object type: %T", obj)
	}
	inputName := sar.Spec.ResourceAttributes.Name
	if inputName == loggingv1.InputNameApplication || inputName == loggingv1.InputNameInfrastructure {
		sar.Status.Allowed = true
	}
	return nil
}
