package version

import (
	"context"

	configv1 "github.com/openshift/api/config/v1"
	"github.com/openshift/cluster-logging-operator/internal/hostedcontrolplane"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var (
	clusterVersion string
	clusterID      string
)

// ClusterVersion retrieves the ClusterVersion spec
func ClusterVersion(k8client client.Reader) (string, string, error) {
	if clusterVersion == "" && clusterID == "" {
		proto := &configv1.ClusterVersion{}
		key := client.ObjectKey{Name: "version"}
		if err := k8client.Get(context.TODO(), key, proto); err != nil {
			return "", "", err
		}
		clusterVersion = proto.Status.Desired.Version
		clusterID = string(proto.Spec.ClusterID)
	}
	return clusterVersion, clusterID, nil
}

// HostedClusterVersion retrieves the version info of the hosted cluster
func HostedClusterVersion(ctx context.Context, k8client client.Reader, namespace string) (version, id string) {
	// If reconciling in a hosted control plane namespace, use the guest cluster version/id
	// provided by the hostedcontrolplane resource.
	if info := hostedcontrolplane.GetVersionID(ctx, k8client, namespace); info != nil {
		return info.Version, info.ID
	}
	return "", ""
}
