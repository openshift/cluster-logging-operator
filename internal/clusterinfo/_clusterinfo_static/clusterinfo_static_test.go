package clusterinfo

import (
	"context"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	configv1 "github.com/openshift/api/config/v1"
	hypershiftv1beta1 "github.com/openshift/hypershift/api/hypershift/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func init() {
	hypershiftv1beta1.AddToScheme(scheme.Scheme)
}

var _ = Describe("[internal][clusterinfo]", func() {
	// All clusters have a ClusterVersion
	clusterVersion := &configv1.ClusterVersion{
		ObjectMeta: metav1.ObjectMeta{Name: "version"},
		Spec: configv1.ClusterVersionSpec{
			ClusterID:     "clusterVersion-id",
			DesiredUpdate: &configv1.Update{Version: "clusterVersion-version"},
		}}

	// Each HCP management namespace has a HostedControlPlane
	hcp := &hypershiftv1beta1.HostedControlPlane{
		ObjectMeta: metav1.ObjectMeta{Namespace: "testing", Name: "foobar"},
		Spec: hypershiftv1beta1.HostedControlPlaneSpec{
			ClusterID: "hypershift-id",
		},
		Status: hypershiftv1beta1.HostedControlPlaneStatus{
			VersionStatus: &hypershiftv1beta1.ClusterVersionStatus{
				Desired: configv1.Release{
					Version: "hypershift-version",
				},
			},
		},
	}

	DescribeTable("GetClusterInfo", func(id, version string, objects ...client.Object) {
		c := fake.NewClientBuilder().WithObjects(objects...).Build()
		info, err := Get(context.Background(), c, hcp.Namespace)
		Expect(err).To(Succeed())
		Expect(info).To(Equal(ClusterInfo{Version: version, ID: id}))
	},
		Entry("Standalone cluster", "clusterVersion-id", "clusterVersion-version", clusterVersion),
		Entry("Hypershift cluster", "hypershift-id", "hypershift-version", clusterVersion, hcp),
	)
})
