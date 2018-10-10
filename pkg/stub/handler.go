package stub

import (
	"context"
	"fmt"

	"github.com/openshift/cluster-logging-operator/pkg/apis/logging/v1alpha1"
	"github.com/openshift/cluster-logging-operator/pkg/k8shandler"
	"github.com/operator-framework/operator-sdk/pkg/sdk"
)

func NewHandler() sdk.Handler {
	return &Handler{}
}

type Handler struct {
	// Fill me
}

func (h *Handler) Handle(ctx context.Context, event sdk.Event) error {

	// Ignore the delete event since the garbage collector will clean up all secondary resources for the CR
	// All secondary resources must have the CR set as their OwnerReference for this to be the case
	if event.Deleted {
		return nil
	}

	switch o := event.Object.(type) {
	case *v1alpha1.ClusterLogging:
		return Reconcile(o)
	}

	return nil
}

func Reconcile(logging *v1alpha1.ClusterLogging) (err error) {
	// Reconcile certs
	if err = k8shandler.CreateOrUpdateCertificates(logging); err != nil {
		return fmt.Errorf("Unable to create or update certificates: %v", err)
	}

	// Reconcile Log Store
	if err = k8shandler.CreateOrUpdateLogStore(logging); err != nil {
		return fmt.Errorf("Unable to create or update logstore: %v", err)
	}

	// Reconcile Visualization
	if err = k8shandler.CreateOrUpdateVisualization(logging); err != nil {
		return fmt.Errorf("Unable to create or update visualization: %v", err)
	}

	// Reconcile Curation
	if err = k8shandler.CreateOrUpdateCuration(logging); err != nil {
		return fmt.Errorf("Unable to create or update curation: %v", err)
	}

	// Reconcile Collection
	if err = k8shandler.CreateOrUpdateCollection(logging); err != nil {
		return fmt.Errorf("Unable to create or update collection: %v", err)
	}

	if err = k8shandler.UpdateStatus(logging); err != nil {
		return fmt.Errorf("Unable to update Cluster Logging status: %v", err)
	}

	return nil
}
