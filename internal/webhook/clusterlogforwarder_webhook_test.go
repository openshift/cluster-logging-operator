package webhook

import (
	"context"
	"fmt"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	admissionv1 "k8s.io/api/admission/v1"
	authenticationv1 "k8s.io/api/authentication/v1"
	authorizationapi "k8s.io/api/authorization/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

func TestWebhook(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "[internal][webhook] Suite")
}

func contextWithUserInfo(info authenticationv1.UserInfo) context.Context {
	req := admission.Request{
		AdmissionRequest: admissionv1.AdmissionRequest{
			UserInfo: info,
		},
	}
	return admission.NewContextWithRequest(context.TODO(), req)
}

var _ = Describe("ClusterLogForwarder Webhook", func() {

	Context("ValidatingWebhook (Validator)", func() {
		var (
			validator *ClusterLogForwarderValidator
			clf       *obs.ClusterLogForwarder
			mockSAR   *mockSARClient
		)

		BeforeEach(func() {
			mockSAR = &mockSARClient{Client: fake.NewFakeClient(), allowed: true}
			validator = &ClusterLogForwarderValidator{Client: mockSAR}
			clf = &obs.ClusterLogForwarder{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-clf",
					Namespace: "openshift-logging",
				},
				Spec: obs.ClusterLogForwarderSpec{
					ServiceAccount: obs.ServiceAccount{
						Name: "log-collector",
					},
				},
			}
		})

		It("should allow CLF when user has permission to use the SA", func() {
			ctx := contextWithUserInfo(authenticationv1.UserInfo{Username: "admin-user"})
			warnings, err := validator.ValidateCreate(ctx, clf)
			Expect(err).ToNot(HaveOccurred())
			Expect(warnings).To(BeNil())
		})

		It("should reject CLF when user does not have permission to use the SA", func() {
			mockSAR.allowed = false
			ctx := contextWithUserInfo(authenticationv1.UserInfo{Username: "attacker"})
			warnings, err := validator.ValidateCreate(ctx, clf)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("not authorized to use service account"))
			Expect(warnings).To(BeNil())
		})

		It("should allow CLF without a service account name", func() {
			clf.Spec.ServiceAccount.Name = ""
			ctx := contextWithUserInfo(authenticationv1.UserInfo{Username: "bob"})
			warnings, err := validator.ValidateCreate(ctx, clf)
			Expect(err).ToNot(HaveOccurred())
			Expect(warnings).To(BeNil())
		})

		It("should allow delete without checks", func() {
			ctx := contextWithUserInfo(authenticationv1.UserInfo{Username: "bob"})
			warnings, err := validator.ValidateDelete(ctx, clf)
			Expect(err).ToNot(HaveOccurred())
			Expect(warnings).To(BeNil())
		})

		It("should validate on update the same as create", func() {
			mockSAR.allowed = false
			ctx := contextWithUserInfo(authenticationv1.UserInfo{Username: "attacker"})
			oldClf := clf.DeepCopy()
			warnings, err := validator.ValidateUpdate(ctx, oldClf, clf)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("not authorized to use service account"))
			Expect(warnings).To(BeNil())
		})

		It("should pass user groups and UID in the SAR", func() {
			ctx := contextWithUserInfo(authenticationv1.UserInfo{
				Username: "bob",
				Groups:   []string{"system:authenticated", "developers"},
				UID:      "uid-456",
			})
			warnings, err := validator.ValidateCreate(ctx, clf)
			Expect(err).ToNot(HaveOccurred())
			Expect(warnings).To(BeNil())
			Expect(mockSAR.lastSAR.Spec.User).To(Equal("bob"))
			Expect(mockSAR.lastSAR.Spec.Groups).To(Equal([]string{"system:authenticated", "developers"}))
			Expect(mockSAR.lastSAR.Spec.UID).To(Equal("uid-456"))
		})

		It("should return error when SAR creation fails", func() {
			mockSAR.err = fmt.Errorf("API server unavailable")
			ctx := contextWithUserInfo(authenticationv1.UserInfo{Username: "bob"})
			warnings, err := validator.ValidateCreate(ctx, clf)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("failed to check service account usage permission"))
			Expect(warnings).To(BeNil())
		})
	})
})

type mockSARClient struct {
	client.Client
	allowed bool
	err     error
	lastSAR *authorizationapi.SubjectAccessReview
}

func (c *mockSARClient) Create(ctx context.Context, obj client.Object, opts ...client.CreateOption) error {
	sar, ok := obj.(*authorizationapi.SubjectAccessReview)
	if !ok {
		return fmt.Errorf("unexpected object type: %T", obj)
	}
	Expect(sar.Spec.ResourceAttributes).ToNot(BeNil())
	Expect(sar.Spec.ResourceAttributes.Resource).To(Equal("serviceaccounts"))
	Expect(sar.Spec.ResourceAttributes.Verb).To(Equal("use"))
	c.lastSAR = sar
	if c.err != nil {
		return c.err
	}
	sar.Status.Allowed = c.allowed
	return nil
}
