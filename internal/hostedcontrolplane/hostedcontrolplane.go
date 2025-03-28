package hostedcontrolplane

import (
	"context"

	hypershiftv1beta1 "github.com/openshift/hypershift/api/hypershift/v1beta1"
	client "sigs.k8s.io/controller-runtime/pkg/client"
)

// VersionID contains the version and ID of a HostedControlPlane
type VersionID struct {
	Version string // Version of the guest cluster.
	ID      string // Unique identifier of the guest cluster.
}

// GetVersionID returns the guest cluster version and ID for a HCP namepace.
// Returns nil if the namespace is not a valid HCP namespace with a non-empty version and ID.
func GetVersionID(ctx context.Context, c client.Reader, namespace string) *VersionID {
	hcps := &hypershiftv1beta1.HostedControlPlaneList{}
	err := c.List(ctx, hcps, client.InNamespace(namespace))
	if err != nil || len(hcps.Items) != 1 {
		// A valid HCP namespace must contain exactly one HCP instance, anything else is invalid.
		return nil
	}
	id := hcps.Items[0].Spec.ClusterID
	version := hcps.Items[0].Status.VersionStatus.Desired.Version
	if version == "" {
		// Check deprecated top-level version field, see:
		// https://github.com/openshift/hypershift/blob/9ccf6274be95242f7b17623e2a40724b9a6a5595/api/hypershift/v1beta1/hosted_controlplane.go#L260		version = hcps.Items[0].Status.Version //nolint:staticcheck
		version = hcps.Items[0].Status.Version //nolint:staticcheck
	}
	if id == "" || version == "" {
		return nil
	}
	return &VersionID{Version: version, ID: id}
}
