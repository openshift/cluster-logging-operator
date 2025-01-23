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

// Deployment reconciles a Deployment to the desired spec returning an error
// if there is an issue creating or updating to the desired state
func Deployment(k8Client client.Client, desired *apps.Deployment) error {
	dpl := runtime.NewDeployment(desired.Namespace, desired.Name)
	op, err := controllerutil.CreateOrUpdate(context.TODO(), k8Client, dpl, func() error {
		// Update the deployment with our desired state
		dpl.Labels = desired.Labels
		dpl.Spec = desired.Spec
		dpl.OwnerReferences = desired.OwnerReferences
		return nil
	})

	if err == nil {
		log.V(3).Info(fmt.Sprintf("reconciled deployment - operation: %s", op))
	}
	return err
}
