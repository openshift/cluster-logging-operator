package stub

import (
	"context"
	"fmt"

	"github.com/openshift/cluster-logging-operator/pkg/apis/logging/v1alpha1"
	"github.com/openshift/cluster-logging-operator/pkg/k8shandler"
	"github.com/openshift/cluster-logging-operator/pkg/utils"
	"github.com/operator-framework/operator-sdk/pkg/sdk"
	"github.com/sirupsen/logrus"
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
	exists := utils.DoesClusterLoggingExist(logging)
	if !exists {
		return nil
	}

	//TODO add check for the explicit names we support APP,INFRA
	cluster := k8shandler.NewClusterLogging(logging)
	// Reconcile certs
	if err = k8shandler.CreateOrUpdateCertificates(logging); err != nil {
		return fmt.Errorf("Unable to create or update certificates: %v", err)
	}

	// Reconcile Stacks
	for _, stack := range logging.Spec.Stacks {
		if err := ReconcileStack(cluster, &stack); err != nil {
			return fmt.Errorf("Unable to create stack '%v': %v", stack.Name, err)
		}
	}
	// Reconcile Visualization Singletons
	oauthSecret := utils.GetRandomWord(64)
	if err = k8shandler.CreateOrUpdateKibanaSecret(cluster, oauthSecret); err != nil {
		return
	}
	if err = k8shandler.CreateOrUpdateOauthClient(cluster, string(oauthSecret)); err != nil {
		return
	}
	if err = k8shandler.CreateSharedConfig(cluster); err != nil {
		return
	}

	// Reconcile Collection
	if err = k8shandler.CreateOrUpdateCollection(cluster); err != nil {
		return fmt.Errorf("Unable to create or update collection: %v", err)
	}

	return nil
}
func ReconcileStack(cluster *k8shandler.ClusterLogging, stack *v1alpha1.StackSpec) (err error) {
	if stack.Type != v1alpha1.StackTypeElastic {
		logrus.Debugf("Skipping unrecognized stack type '%v'", stack.Type)
		return nil
	}
	// Reconcile Log Store
	if err = cluster.CreateOrUpdateLogStore(stack); err != nil {
		return fmt.Errorf("Unable to create or update logstore: %v", err)
	}

	// Reconcile Visualization
	if err = cluster.CreateOrUpdateVisualization(stack); err != nil {
		return fmt.Errorf("Unable to create or update visualization: %v", err)
	}

	// Reconcile Curation
	if err = cluster.CreateOrUpdateCuration(stack); err != nil {
		return fmt.Errorf("Unable to create or update curation: %v", err)
	}

	return nil
}
