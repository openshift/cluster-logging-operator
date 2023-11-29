package clusterinfo

import (
	"context"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	configv1 "github.com/openshift/api/config/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func init() {
	runtime.Must(configv1.AddToScheme(scheme.Scheme))
}

var _ = Describe("[internal][clusterinfo]", func() {

	// All clusters have a ClusterVersion
	cv := &configv1.ClusterVersion{
		ObjectMeta: metav1.ObjectMeta{Name: "version"},
		Spec: configv1.ClusterVersionSpec{
			ClusterID:     "clusterVersion-id",
			DesiredUpdate: &configv1.Update{Version: "clusterVersion-version"},
		}}

	// Each HCP management namespace has a HostedControlPlane
	hcp := &unstructured.Unstructured{}
	hcp.Object = map[string]any{
		"spec": map[string]any{
			"clusterID": "hypershift-id",
			"release":   "hypershift-version",
		},
		"status": map[string]any{
			"versionStatus": map[string]any{
				"desired": map[string]any{ //configv1.Release{
					"version": "hypershift-version",
				},
			},
		},
	}
	hcp.SetGroupVersionKind(hcpGVK)
	hcp.SetName("foobar")
	hcp.SetNamespace("testing")

	DescribeTable("Get", func(id, version string, objs ...client.Object) {
		c := fake.NewClientBuilder().WithObjects(objs...).Build()
		info, err := Get(context.Background(), c, hcp.GetNamespace())
		Expect(err).To(Succeed())
		Expect(info).To(Equal(&ClusterInfo{Version: version, ID: id}))
	},
		Entry("Uses hostedControlPlane in hypershift cluster", "hypershift-id", "hypershift-version", cv, hcp),
		Entry("uses clusterversion in standalone cluster", "clusterVersion-id", "clusterVersion-version", cv),
	)
})
