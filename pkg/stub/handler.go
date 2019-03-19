package stub

import (
	"context"
	"fmt"

	logging "github.com/openshift/cluster-logging-operator/pkg/apis/logging/v1alpha1"
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
	case *logging.ClusterLogging:
		return Reconcile(o)
	}

	return nil
}

// Reconcile deploys or updates cluster logging to match its spec
func Reconcile(spec *logging.ClusterLogging) (err error) {
	cluster := k8shandler.NewClusterLogging(spec)
	// Reconcile certs
	if cluster.Exists() {
		if cluster.Spec.ManagementState == logging.ManagementStateManaged {
			if err = cluster.CreateOrUpdateCertificates(); err != nil {
				return fmt.Errorf("Unable to create or update certificates for %q: %v", cluster.Name, err)
			}
		}
	}

	// Reconcile Log Store
	if cluster.Exists() {
		if cluster.Spec.ManagementState == logging.ManagementStateManaged {
			if err = cluster.CreateOrUpdateLogStore(); err != nil {
				return fmt.Errorf("Unable to create or update logstore for %q: %v", cluster.Name, err)
			}
		}
	}

	// Reconcile Visualization
	if cluster.Exists() {
		if cluster.Spec.ManagementState == logging.ManagementStateManaged {
			if err = cluster.CreateOrUpdateVisualization(); err != nil {
				return fmt.Errorf("Unable to create or update visualization for %q: %v", cluster.Name, err)
			}
		}
	}

	// Reconcile Curation
	if cluster.Exists() {
		if cluster.Spec.ManagementState == logging.ManagementStateManaged {
			if err = cluster.CreateOrUpdateCuration(); err != nil {
				return fmt.Errorf("Unable to create or update curation for %q: %v", cluster.Name, err)
			}
		}
	}

	// Reconcile Collection
	if cluster.Exists() {
		if cluster.Spec.ManagementState == logging.ManagementStateManaged {
			if err = cluster.CreateOrUpdateCollection(); err != nil {
				return fmt.Errorf("Unable to create or update collection for %q: %v", cluster.Name, err)
			}
		}
	}

	return nil
}
