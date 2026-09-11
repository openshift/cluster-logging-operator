package admission

import (
	"fmt"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	internalreconcile "github.com/openshift/cluster-logging-operator/internal/reconcile"
	admissionregistrationv1 "k8s.io/api/admissionregistration/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	apiruntime "k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/discovery"
)

func TestAdmission(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "[internal][admission] Suite")
}

var _ = Describe("admission policy helpers", func() {
	It("recognizes unsupported admission policy API errors", func() {
		Expect(internalreconcile.IsUnsupportedAdmissionPolicyAPI(&meta.NoKindMatchError{
			GroupKind: schema.GroupKind{Group: admissionregistrationv1.GroupName, Kind: "ValidatingAdmissionPolicy"},
		})).To(BeTrue())
		Expect(internalreconcile.IsUnsupportedAdmissionPolicyAPI(apiruntime.NewNotRegisteredErrForKind(
			"test", schema.GroupVersionKind{Group: admissionregistrationv1.GroupName, Version: "v1", Kind: "ValidatingAdmissionPolicy"},
		))).To(BeTrue())
		Expect(internalreconcile.IsUnsupportedAdmissionPolicyAPI(&discovery.ErrGroupDiscoveryFailed{
			Groups: map[schema.GroupVersion]error{
				{Group: admissionregistrationv1.GroupName, Version: "v1"}: fmt.Errorf("discovery failed"),
			},
		})).To(BeTrue())
		Expect(internalreconcile.IsUnsupportedAdmissionPolicyAPI(fmt.Errorf("forbidden"))).To(BeFalse())
	})
})
