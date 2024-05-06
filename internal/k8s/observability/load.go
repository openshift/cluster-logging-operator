package observability

import (
	"context"

	log "github.com/ViaQ/logerr/v2/log/static"
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	obsmigrate "github.com/openshift/cluster-logging-operator/internal/migrations/observability"
	"github.com/openshift/cluster-logging-operator/internal/runtime"
	obsruntime "github.com/openshift/cluster-logging-operator/internal/runtime/observability"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// FetchClusterLogForwarder, migrate and validate
func FetchClusterLogForwarder(k8sClient client.Client, namespace, name string) (forwarder obs.ClusterLogForwarder, err error, status *obs.ClusterLogForwarderStatus) {
	key := types.NamespacedName{Name: name, Namespace: namespace}
	proto := obsruntime.NewClusterLogForwarder(namespace, name, runtime.Initialize)
	if err = k8sClient.Get(context.TODO(), key, proto); err != nil {
		if !apierrors.IsNotFound(err) {
			log.Error(err, "Encountered unexpected error getting", "forwarder", key)
			return obs.ClusterLogForwarder{}, err, nil
		}
	}
	// Do not modify cached copy
	forwarder = *proto.DeepCopy()
	var migrationConds []metav1.Condition
	forwarder.Spec, migrationConds = obsmigrate.MigrateClusterLogForwarder(forwarder.Spec)

	forwarder.Status.Conditions = append(forwarder.Status.Conditions, migrationConds...)
	// TODO Integrate validate if needed
	return forwarder, nil, status
}
