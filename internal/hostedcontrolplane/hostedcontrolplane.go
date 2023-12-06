package hostedcontrolplane

import (
	"context"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
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
	// TODO: Using unstructured objects for HCP because of dependency problems with the hypershift API package.
	// See https://issues.redhat.com/browse/HOSTEDCP-336
	l := &unstructured.UnstructuredList{}
	l.SetGroupVersionKind(hcpGVK)
	err := c.List(ctx, l, client.InNamespace(namespace))
	if err != nil || len(l.Items) != 1 {
		// A valid HCP namespace must contain exactly one HCP instance, anything else is invalid.
		return nil
	}
	id := dig(l.Items[0].Object, "spec", "clusterID")
	version := dig(l.Items[0].Object, "status", "versionStatus", "desired", "version")
	if version == "" {
		// Check deprecated top-level version field, see:
		// https://github.com/openshift/hypershift/blob//api/hypershift/v1beta1/hosted_controlplane.go#L241
		version = dig(l.Items[0].Object, "status", "version")
	}
	if id == "" || version == "" {
		return nil
	}
	return &VersionID{Version: version, ID: id}
}

var hcpGVK = schema.GroupVersionKind{
	Group:   "hypershift.openshift.io",
	Version: "v1beta1",
	Kind:    "HostedControlPlane",
}

// dig out a string value from nested map[string]any
func dig(v any, keys ...string) string {
	for _, k := range keys {
		m, _ := v.(map[string]any)
		v = m[k] // Index of nil map returns nil.
	}
	s, _ := v.(string)
	return s
}
