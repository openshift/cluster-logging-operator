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
	"k8s.io/apimachinery/pkg/util/sets"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

var _ = Describe("[internal][validations] validate clusterlogforwarder service account", func() {
	var (
		k8sClient client.Client
		customClf loggingv1.ClusterLogForwarder
		extras    map[string]bool
	)

	BeforeEach(func() {
		customClf = loggingv1.ClusterLogForwarder{
			ObjectMeta: v1.ObjectMeta{
				Name:      "test-clf",
				Namespace: constants.OpenshiftNS,
			},
		}
		extras = map[string]bool{}
	})

	Context("service account existence", func() {
		It("should fail when no service account found", func() {
			k8sClient = fake.NewClientBuilder().Build()
			serviceAccount, err := getServiceAccount(customClf.Spec.ServiceAccountName, customClf.Namespace, k8sClient)
			Expect(serviceAccount).To(BeNil())
			Expect(err).To(MatchError(MatchRegexp("service account: .+ not found")))
		})

		It("should succeed when service account is found", func() {
			providedSAName := "test-serviceAccount"
			clfServiceAccount := &corev1.ServiceAccount{
				ObjectMeta: v1.ObjectMeta{
					Name:      providedSAName,
					Namespace: customClf.Namespace,
				},
			}
			k8sClient = fake.NewClientBuilder().WithObjects(clfServiceAccount).Build()
			customClf.Spec.ServiceAccountName = providedSAName

			serviceAccount, err := getServiceAccount(customClf.Spec.ServiceAccountName, customClf.Namespace, k8sClient)
			Expect(serviceAccount).ToNot(BeNil())
			Expect(err).To(BeNil())
		})
	})

	Context("validate permissions", func() {
		Context("gather clf inputs", func() {
			var expectedSet sets.String

			var clf = loggingv1.ClusterLogForwarder{
				ObjectMeta: v1.ObjectMeta{
					Name:      "custom-clf",
					Namespace: constants.OpenshiftNS,
				},
			}
			BeforeEach(func() {
				clf.Spec = loggingv1.ClusterLogForwarderSpec{}
			})

			It("should gather all inputs from clf.spec.pipelines", func() {
				clf.Spec = loggingv1.ClusterLogForwarderSpec{
					Pipelines: []loggingv1.PipelineSpec{
						{
							Name: "pipeline1",
							InputRefs: []string{
								loggingv1.InputNameAudit,
								loggingv1.InputNameApplication,
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
				expectedSet = sets.NewString(loggingv1.InputNameAudit, loggingv1.InputNameApplication)
				inputs := gatherPipelineInputs(clf)
				Expect(inputs).To(HaveLen(2))
				Expect(inputs).To(Equal(expectedSet))
			})
		})
		Context("subjectAccessReview", func() {
			var (
				clfServiceAccount = &corev1.ServiceAccount{
					ObjectMeta: v1.ObjectMeta{
						Name:      "test-serviceAccount",
						Namespace: constants.OpenshiftNS,
					},
				}
				mockSARClient = &mockSARClient{}
			)
			It("should pass validation when service account can collect specified inputs", func() {
				Expect(validateServiceAccountPermissions(mockSARClient, sets.NewString(loggingv1.InputNameApplication, loggingv1.InputNameInfrastructure), clfServiceAccount, constants.OpenshiftNS)).To(Succeed())
			})

			It("should return validation error if service account cannot collect specified inputs", func() {
				Expect(validateServiceAccountPermissions(mockSARClient, sets.NewString(loggingv1.InputNameApplication, loggingv1.InputNameAudit), clfServiceAccount, constants.OpenshiftNS)).ToNot(Succeed())
			})
		})
	})

	It("should not validate clusterlogforwarder named 'instance'", func() {
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

	It("should return an error if custom clusterlogforwarder does not include a service account name", func() {
		Expect(ValidateServiceAccount(customClf, k8sClient, extras)).ToNot(Succeed())
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
