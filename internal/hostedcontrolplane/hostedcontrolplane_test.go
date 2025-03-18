package hostedcontrolplane

import (
	"context"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	configv1 "github.com/openshift/api/config/v1"
	hypershiftv1beta1 "github.com/openshift/hypershift/api/hypershift/v1beta1"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

var _ = Describe("[internal][hostedcontrolplane]", func() {
	runtime.Must(configv1.AddToScheme(scheme.Scheme))
	runtime.Must(hypershiftv1beta1.AddToScheme(scheme.Scheme))
	var (
		hcp                  *hypershiftv1beta1.HostedControlPlane
		version              = "v1.2.3"
		id                   = "1234-abcd"
		hcpName              = "foobar"
		hcpNamespace         = "testing"
		hcpDeprecatedVersion = "v1.2.4"
	)

	BeforeEach(func() {
		hcp = &hypershiftv1beta1.HostedControlPlane{
			Spec: hypershiftv1beta1.HostedControlPlaneSpec{
				ClusterID: id,
			},
			Status: hypershiftv1beta1.HostedControlPlaneStatus{
				VersionStatus: &hypershiftv1beta1.ClusterVersionStatus{
					Desired: configv1.Release{
						Version: version,
					},
				},
			},
		}
		hcp.SetName(hcpName)
		hcp.SetNamespace(hcpNamespace)
	})

	Describe("GetVersionID", func() {
		It("gets version and ID for a HCP namespace", func() {
			c := fake.NewFakeClient(hcp)
			info := GetVersionID(context.Background(), c, hcp.GetNamespace())
			Expect(info).NotTo(BeNil())
			Expect(info).To(Equal(&VersionID{Version: version, ID: id}))
		})

		It("gets version and ID for a HCP namespace using deprecated version field", func() {
			hcp.Status.Version = hcpDeprecatedVersion //nolint:staticcheck
			c := fake.NewFakeClient(hcp)
			info := GetVersionID(context.Background(), c, hcp.GetNamespace())
			Expect(info).NotTo(BeNil())
			Expect(info).To(Equal(&VersionID{Version: version, ID: id}))
		})

		It("returns nil for non-HCP namespace", func() {
			c := fake.NewFakeClient()
			info := GetVersionID(context.Background(), c, "normalNs")
			Expect(info).To(BeNil())
		})

		It("returns nil for an invalid HCP object (lacking version and ID)", func() {
			hcp.Spec.ClusterID = ""
			hcp.Status.VersionStatus.Desired.Version = ""
			c := fake.NewFakeClient(hcp)
			info := GetVersionID(context.Background(), c, hcp.GetNamespace())
			Expect(info).To(BeNil())
		})
	})
})
