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
