package observability

import (
	"context"
	log "github.com/ViaQ/logerr/v2/log/static"
	obsv1 "github.com/openshift/cluster-logging-operator/api/observability/v1"
	"github.com/openshift/cluster-logging-operator/internal/runtime"
	obsruntime "github.com/openshift/cluster-logging-operator/internal/runtime/observability"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// FetchClusterLogForwarder, migrate and validate
func FetchClusterLogForwarder(k8sClient client.Client, namespace, name string) (forwarder obsv1.ClusterLogForwarder, err error, status *obsv1.ClusterLogForwarderStatus) {
	key := types.NamespacedName{Name: name, Namespace: namespace}
	proto := obsruntime.NewClusterLogForwarder(namespace, name, runtime.Initialize)
	if err = k8sClient.Get(context.TODO(), key, proto); err != nil {
		if !apierrors.IsNotFound(err) {
			log.Error(err, "Encountered unexpected error getting", "forwarder", key)
			return obsv1.ClusterLogForwarder{}, err, nil
		}
	}

	// Do not modify cached copy
	forwarder = *proto.DeepCopy()
	// TODO Integrate migrate if needed
	// TODO Integrate validate if needed
	return forwarder, nil, status
}
