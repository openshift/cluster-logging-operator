package hostedcontrolplane

import (
	"context"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	configv1 "github.com/openshift/api/config/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

var _ = Describe("[internal][hostedcontrolplane]", func() {
	runtime.Must(configv1.AddToScheme(scheme.Scheme))
	version, id := "v1.2.3", "1234-abcd"

	Describe("GetVersionID", func() {
		It("gets version and ID for a HCP namespace", func() {
			hcp := &unstructured.Unstructured{}
			hcp.Object = map[string]any{
				"spec": map[string]any{
					"clusterID": id,
				},
				"status": map[string]any{
					"versionStatus": map[string]any{
						"desired": map[string]any{ //configv1.Release{
							"version": version,
						},
					},
				},
			}
			hcp.SetGroupVersionKind(hcpGVK)
			hcp.SetName("foobar")
			hcp.SetNamespace("testing")
			c := fake.NewFakeClient(hcp)
			info := GetVersionID(context.Background(), c, hcp.GetNamespace())
			Expect(info).NotTo(BeNil())
			Expect(info).To(Equal(&VersionID{Version: version, ID: id}))
		})

		It("gets version and ID for a HCP namespace using deprecated version field", func() {
			hcp := &unstructured.Unstructured{}
			hcp.Object = map[string]any{
				"spec": map[string]any{
					"clusterID": id,
				},
				"status": map[string]any{
					"version": version,
				},
			}
			hcp.SetGroupVersionKind(hcpGVK)
			hcp.SetName("foobar")
			hcp.SetNamespace("testing")
			c := fake.NewFakeClient(hcp)
			info := GetVersionID(context.Background(), c, hcp.GetNamespace())
			Expect(info).NotTo(BeNil())
			Expect(info).To(Equal(&VersionID{Version: version, ID: id}))
		})

		It("returns nil for non-HCP namespace", func() {
			c := fake.NewFakeClient()
			info := GetVersionID(context.Background(), c, "testing")
			Expect(info).To(BeNil())
		})

		It("returns nil for an invalid HCP object (lacking version and ID)", func() {
			hcp := &unstructured.Unstructured{}
			hcp.Object = map[string]any{
				"spec": map[string]any{
					"clusterID": map[string]any{"foo": "bar"}, // Invalid ID
				},
				"status": map[string]any{}, // Missing version
			}
			hcp.SetGroupVersionKind(hcpGVK)
			hcp.SetName("foobar")
			hcp.SetNamespace("testing")
			c := fake.NewFakeClient(hcp)
			info := GetVersionID(context.Background(), c, hcp.GetNamespace())
			Expect(info).To(BeNil())
		})

	})
})
