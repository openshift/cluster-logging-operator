package clusterinfo

import (
	"context"
	configv1 "github.com/openshift/api/config/v1"
	hypershiftv1beta1 "github.com/openshift/hypershift/api/hypershift/v1beta1"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func init() {
	hypershiftv1beta1.AddToScheme(scheme.Scheme)
	configv1.AddToScheme(scheme.Scheme)
}

func init() {
	runtime.Must(hypershiftv1beta1.AddToScheme(scheme.Scheme))
	runtime.Must(configv1.AddToScheme(scheme.Scheme))
}

// ClusterInfo is global information about where the ClusterLogForwarder is running.
type ClusterInfo struct {
	Version string // Version of the cluster.
	ID      string // Unique identifier of the cluster.
}

// GetClusterInfo gets cluster info for the cluster we are running in.
//
// If namespace contains a HostedControlPlane then return info for the *guest* cluster, not the host cluster.
// We assume in this case that CLF is running on behalf of the guest cluster to collect API audit logs.
func Get(ctx context.Context, c client.Reader, namespace string) (ClusterInfo, error) {
	// Use HCP info if exactly one HCP is present in the namespace.
	hcps := &hypershiftv1beta1.HostedControlPlaneList{}
	err := c.List(context.Background(), hcps, client.InNamespace(namespace))
	if err == nil && len(hcps.Items) == 1 {
		return ClusterInfo{
			Version: hcps.Items[0].Status.VersionStatus.Desired.Version,
			ID:      hcps.Items[0].Spec.ClusterID,
		}, nil
	}
	// Use standalone ClusterVersion info.
	cv := &configv1.ClusterVersion{}
	if err := c.Get(ctx, client.ObjectKey{Name: "version"}, cv); err != nil {
		return ClusterInfo{}, err
	}
	return ClusterInfo{
		Version: cv.Spec.DesiredUpdate.Version,
		ID:      string(cv.Spec.ClusterID),
	}, nil
}
