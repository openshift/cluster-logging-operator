package clusterinfo

import (
	"context"

	configv1 "github.com/openshift/api/config/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/kubernetes/scheme"
	client "sigs.k8s.io/controller-runtime/pkg/client"
)

func init() {
	runtime.Must(configv1.AddToScheme(scheme.Scheme))
}

// ClusterInfo is global information about where the ClusterLogForwarder is running.
type ClusterInfo struct {
	Version string // Version of the cluster.
	ID      string // Unique identifier of the cluster.
}

// GetClusterInfo gets cluster info for the cluster we are running in.
//
// If is running in a Hosted Control Plane (hypershift management cluster)
// this information describes the *guest* cluster, not the host cluster.
// We assume in this case that CLF is running on behalf of the guest cluster to collect API audit logs.
func Get(ctx context.Context, c client.Reader, namespace string) (*ClusterInfo, error) {
	if hcp := getHCPInfo(ctx, c, namespace); hcp != nil { // Try HCP first
		return hcp, nil
	}
	cv := &configv1.ClusterVersion{}
	if err := c.Get(ctx, client.ObjectKey{Name: "version"}, cv); err != nil {
		return nil, err
	}
	return &ClusterInfo{Version: cv.Spec.DesiredUpdate.Version, ID: string(cv.Spec.ClusterID)}, nil
}

var (
	hcpGVK = schema.GroupVersionKind{
		Group:   "hypershift.openshift.io",
		Version: "v1beta1",
		Kind:    "HostedControlPlane",
	}
)

// getHCPInfo returns ClusterInfo from a hypershift control plane if found, nil otherwise.
func getHCPInfo(ctx context.Context, c client.Reader, namespace string) *ClusterInfo {

	// TODO: Using unstructured objects for HCP because of dependency problems with the hypershift API package.
	// When https://issues.redhat.com/browse/HOSTEDCP-336 is fixed, switch to static approach.
	// Code in ./_clusterinfo_static

	l := &unstructured.UnstructuredList{}
	l.SetGroupVersionKind(hcpGVK)
	err := c.List(ctx, l, client.InNamespace(namespace))
	if err != nil || len(l.Items) != 1 {
		return nil
	}
	spec, _ := l.Items[0].Object["spec"].(map[string]any)
	if spec == nil || spec["release"] == nil || spec["clusterID"] == nil {
		return nil
	}
	return &ClusterInfo{Version: spec["release"].(string), ID: spec["clusterID"].(string)}
}
