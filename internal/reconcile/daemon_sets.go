package reconcile

import (
	"context"
	"fmt"

	log "github.com/ViaQ/logerr/v2/log/static"
	"github.com/openshift/cluster-logging-operator/internal/runtime"
	apps "k8s.io/api/apps/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

// DaemonSet reconciles a DaemonSet to the desired spec returning an error
// if there is an issue creating or updating to the desired state
func DaemonSet(k8Client client.Client, desired *apps.DaemonSet) error {
	ds := runtime.NewDaemonSet(desired.Namespace, desired.Name)
	op, err := controllerutil.CreateOrUpdate(context.TODO(), k8Client, ds, func() error {
		// Update the daemonset with our desired state
		ds.Labels = desired.Labels
		ds.Spec = desired.Spec
		ds.OwnerReferences = desired.OwnerReferences
		return nil
	})

	if err == nil {
		log.V(3).Info(fmt.Sprintf("reconciled daemonset - operation: %s", op))
	}
	return err
}
