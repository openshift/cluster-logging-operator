package stub

import (
	"context"
	"fmt"

	"github.com/ViaQ/elasticsearch-operator/pkg/apis/elasticsearch/v1alpha1"
	"github.com/ViaQ/elasticsearch-operator/pkg/k8shandler"
	"github.com/sirupsen/logrus"

	"github.com/operator-framework/operator-sdk/pkg/sdk"
)

func NewHandler() sdk.Handler {
	return &Handler{}
}

type Handler struct {
	// Fill me
}

func (h *Handler) Handle(ctx context.Context, event sdk.Event) error {
	if event.Deleted {
		return nil
	}

	switch o := event.Object.(type) {
	case *v1alpha1.Elasticsearch:
		return Reconcile(o)
	}
	return nil
}

// Reconcile reconciles the cluster's state to the spec specified
func Reconcile(es *v1alpha1.Elasticsearch) (err error) {
	logrus.Info("Started reconciliation")

	// Ensure existence of services
	err = k8shandler.CreateOrUpdateServices(es)
	if err != nil {
		return fmt.Errorf("Failed to reconcile Services for Elasticsearch cluster: %v", err)
	}

	// Ensure existence of servicesaccount
	serviceAccountName, err := k8shandler.CreateOrUpdateServiceAccount(es)
	if err != nil {
		return fmt.Errorf("Failed to reconcile ServiceAccount for Elasticsearch cluster: %v", err)
	}

	// Ensure existence of services
	configMapName, err := k8shandler.CreateOrUpdateConfigMaps(es)
	if err != nil {
		return fmt.Errorf("Failed to reconcile ConfigMaps for Elasticsearch cluster: %v", err)
	}

	// TODO: Ensure existence of storage?

	// Ensure Elasticsearch cluster itself is up to spec
	err = k8shandler.CreateOrUpdateElasticsearchCluster(es, configMapName, serviceAccountName)
	if err != nil {
		return fmt.Errorf("Failed to reconcile Elasticsearch deployment spec: %v", err)
	}

	return nil

}
