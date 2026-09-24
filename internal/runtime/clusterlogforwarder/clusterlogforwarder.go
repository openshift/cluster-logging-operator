package clusterlogforwarder

import (
	"context"
	"fmt"
	"strings"

	obsv1 "github.com/openshift/cluster-logging-operator/api/observability/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// ServiceAccountRef is a namespace + name pair identifying a collector ServiceAccount.
type ServiceAccountRef struct {
	Namespace string
	Name      string
}

// String returns the ConfigMap data key for this ServiceAccount reference.
// The format "sa_<namespace>_<name>" is collision-free because namespaces
// (DNS-1123 label) and ServiceAccount names (DNS-1123 subdomain) forbid '_',
// and ConfigMap keys may not contain '/'.
func (r ServiceAccountRef) String() string {
	return fmt.Sprintf("sa_%s_%s", r.Namespace, r.Name)
}

// ListServiceAccounts lists all ClusterLogForwarders and returns the
// unique set of ServiceAccount references they declare.
func ListServiceAccounts(ctx context.Context, k8sClient client.Reader) ([]ServiceAccountRef, error) {
	clfList := &obsv1.ClusterLogForwarderList{}
	if err := k8sClient.List(ctx, clfList); err != nil {
		return nil, fmt.Errorf("list ClusterLogForwarders: %w", err)
	}
	seen := map[ServiceAccountRef]struct{}{}
	var refs []ServiceAccountRef
	for i := range clfList.Items {
		sa := strings.TrimSpace(clfList.Items[i].Spec.ServiceAccount.Name)
		if sa == "" {
			continue
		}
		ref := ServiceAccountRef{Namespace: clfList.Items[i].Namespace, Name: sa}
		if _, ok := seen[ref]; !ok {
			seen[ref] = struct{}{}
			refs = append(refs, ref)
		}
	}
	return refs, nil
}
