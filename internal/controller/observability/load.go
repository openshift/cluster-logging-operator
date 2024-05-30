package observability

import (
	"context"

	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	"github.com/openshift/cluster-logging-operator/internal/runtime"
	obsruntime "github.com/openshift/cluster-logging-operator/internal/runtime/observability"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// FetchClusterLogForwarder returns a copy of the ClusterLogForwarder
func FetchClusterLogForwarder(k8sClient client.Client, namespace, name string) (*obs.ClusterLogForwarder, error) {
	key := types.NamespacedName{Name: name, Namespace: namespace}
	proto := obsruntime.NewClusterLogForwarder(namespace, name, runtime.Initialize)
	if err := k8sClient.Get(context.TODO(), key, proto); err != nil {
		return proto, err
	}
	// Do not modify cached copy
	return proto.DeepCopy(), nil
}
