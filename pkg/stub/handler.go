package stub

import (
	"context"
	"fmt"

	logging "github.com/openshift/cluster-logging-operator/pkg/apis/logging/v1alpha1"
	"github.com/openshift/cluster-logging-operator/pkg/k8shandler"
	"github.com/openshift/cluster-logging-operator/pkg/utils"
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

func Reconcile(cluster *logging.ClusterLogging) (err error) {
	exists := true

	// Reconcile certs
	if exists, cluster = utils.DoesClusterLoggingExist(cluster); exists {
		if cluster.Spec.ManagementState == logging.ManagementStateManaged {
			if err = k8shandler.CreateOrUpdateCertificates(cluster); err != nil {
				return fmt.Errorf("Unable to create or update certificates: %v", err)
			}
		}
	}

	// Reconcile Log Store
	if exists, cluster = utils.DoesClusterLoggingExist(cluster); exists {
		if cluster.Spec.ManagementState == logging.ManagementStateManaged {
			if err = k8shandler.CreateOrUpdateLogStore(cluster); err != nil {
				return fmt.Errorf("Unable to create or update logstore: %v", err)
			}
		}
	}

	// Reconcile Visualization
	if exists, cluster = utils.DoesClusterLoggingExist(cluster); exists {
		if cluster.Spec.ManagementState == logging.ManagementStateManaged {
			if err = k8shandler.CreateOrUpdateVisualization(cluster); err != nil {
				return fmt.Errorf("Unable to create or update visualization: %v", err)
			}
		}
	}

	// Reconcile Curation
	if exists, cluster = utils.DoesClusterLoggingExist(cluster); exists {
		if cluster.Spec.ManagementState == logging.ManagementStateManaged {
			if err = k8shandler.CreateOrUpdateCuration(cluster); err != nil {
				return fmt.Errorf("Unable to create or update curation: %v", err)
			}
		}
	}

	// Reconcile Collection
	if exists, cluster = utils.DoesClusterLoggingExist(cluster); exists {
		if cluster.Spec.ManagementState == logging.ManagementStateManaged {
			if err = k8shandler.CreateOrUpdateCollection(cluster); err != nil {
				return fmt.Errorf("Unable to create or update collection: %v", err)
			}
		}
	}

	return nil
}
