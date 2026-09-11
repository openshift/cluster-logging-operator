package observability

import (
	"context"
	"fmt"
	"strings"

	obsv1 "github.com/openshift/cluster-logging-operator/api/observability/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// ServiceAccountRef is a namespace + name pair identifying a collector ServiceAccount.
type ServiceAccountRef struct {
	Namespace string
	Name      string
}

// CollectorServiceAccounts lists all ClusterLogForwarders and returns the
// unique set of ServiceAccount references they declare.
func CollectorServiceAccounts(ctx context.Context, k8sClient client.Reader) ([]ServiceAccountRef, error) {
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

// Initializer is a function that knows how to initialize a kubernetes runtime object
type Initializer func(o runtime.Object, namespace, name string, visitors ...func(o runtime.Object))

// NewClusterLogForwarder returns a ClusterLogForwarder with name and namespace.
func NewClusterLogForwarder(namespace, name string, initialize Initializer, visitors ...func(clf *obsv1.ClusterLogForwarder)) *obsv1.ClusterLogForwarder {
	clf := &obsv1.ClusterLogForwarder{}
	initialize(clf, namespace, name)
	for _, v := range visitors {
		v(clf)
	}
	clf.Spec.ManagementState = obsv1.ManagementStateManaged
	return clf
}
