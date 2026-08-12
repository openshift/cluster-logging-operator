package admission

import (
	"context"
	"fmt"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	admissionregistrationv1 "k8s.io/api/admissionregistration/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	apiruntime "k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/discovery"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestAdmission(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "[internal][admission] Suite")
}

var _ = Describe("SA usage ValidatingAdmissionPolicy", func() {
	var (
		ctx        context.Context
		fakeClient client.Client
	)

	BeforeEach(func() {
		ctx = context.Background()
		scheme := apiruntime.NewScheme()
		Expect(admissionregistrationv1.AddToScheme(scheme)).To(Succeed())
		fakeClient = fake.NewClientBuilder().WithScheme(scheme).Build()
	})

	It("creates the policy and binding from embedded manifests", func() {
		Expect(ReconcileSAUsageAuthorization(ctx, fakeClient)).To(Succeed())

		policy := &admissionregistrationv1.ValidatingAdmissionPolicy{}
		Expect(fakeClient.Get(ctx, client.ObjectKey{Name: SAUsagePolicyName}, policy)).To(Succeed())
		Expect(policy.Spec.FailurePolicy).ToNot(BeNil())
		Expect(*policy.Spec.FailurePolicy).To(Equal(admissionregistrationv1.Fail))
		Expect(policy.Spec.Validations).To(HaveLen(1))
		Expect(policy.Spec.Validations[0].Message).To(ContainSubstring("not authorized to use the referenced ServiceAccount"))

		binding := &admissionregistrationv1.ValidatingAdmissionPolicyBinding{}
		Expect(fakeClient.Get(ctx, client.ObjectKey{Name: SAUsagePolicyBindingName}, binding)).To(Succeed())
		Expect(binding.Spec.PolicyName).To(Equal(SAUsagePolicyName))
		Expect(binding.Spec.ValidationActions).To(ContainElement(admissionregistrationv1.Deny))
	})

	It("updates an existing policy when the spec changes", func() {
		Expect(ReconcileSAUsageAuthorization(ctx, fakeClient)).To(Succeed())

		policy := &admissionregistrationv1.ValidatingAdmissionPolicy{}
		Expect(fakeClient.Get(ctx, client.ObjectKey{Name: SAUsagePolicyName}, policy)).To(Succeed())
		policy.Spec.FailurePolicy = nil
		Expect(fakeClient.Update(ctx, policy)).To(Succeed())

		Expect(ReconcileSAUsageAuthorization(ctx, fakeClient)).To(Succeed())

		Expect(fakeClient.Get(ctx, client.ObjectKey{Name: SAUsagePolicyName}, policy)).To(Succeed())
		Expect(policy.Spec.FailurePolicy).ToNot(BeNil())
		Expect(*policy.Spec.FailurePolicy).To(Equal(admissionregistrationv1.Fail))
	})

	It("decodes the embedded manifests with expected metadata", func() {
		policy, err := desiredSAUsagePolicy()
		Expect(err).ToNot(HaveOccurred())
		Expect(policy.Name).To(Equal(SAUsagePolicyName))

		binding, err := desiredSAUsagePolicyBinding()
		Expect(err).ToNot(HaveOccurred())
		Expect(binding.Name).To(Equal(SAUsagePolicyBindingName))
		Expect(binding.Spec.PolicyName).To(Equal(SAUsagePolicyName))
	})

	It("recognizes unsupported admission policy API errors", func() {
		Expect(isUnsupportedAdmissionPolicyAPI(&meta.NoKindMatchError{
			GroupKind: schema.GroupKind{Group: admissionregistrationv1.GroupName, Kind: "ValidatingAdmissionPolicy"},
		})).To(BeTrue())
		Expect(isUnsupportedAdmissionPolicyAPI(apiruntime.NewNotRegisteredErrForKind(
			"test", schema.GroupVersionKind{Group: admissionregistrationv1.GroupName, Version: "v1", Kind: "ValidatingAdmissionPolicy"},
		))).To(BeTrue())
		Expect(isUnsupportedAdmissionPolicyAPI(&discovery.ErrGroupDiscoveryFailed{
			Groups: map[schema.GroupVersion]error{
				{Group: admissionregistrationv1.GroupName, Version: "v1"}: fmt.Errorf("discovery failed"),
			},
		})).To(BeTrue())
		Expect(isUnsupportedAdmissionPolicyAPI(fmt.Errorf("forbidden"))).To(BeFalse())
	})
})
