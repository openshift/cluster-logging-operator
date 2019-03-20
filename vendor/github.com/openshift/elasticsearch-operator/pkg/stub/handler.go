package stub

import (
	"context"
	"fmt"

	"github.com/openshift/elasticsearch-operator/pkg/apis/elasticsearch/v1alpha1"
	"github.com/openshift/elasticsearch-operator/pkg/k8shandler"
	"github.com/operator-framework/operator-sdk/pkg/sdk"

	"github.com/sirupsen/logrus"
)

func NewHandler() sdk.Handler {
	return &Handler{}
}

type Handler struct{}

func (h *Handler) Handle(ctx context.Context, event sdk.Event) error {

	switch o := event.Object.(type) {
	case *v1alpha1.Elasticsearch:
		if event.Deleted {
			Flush(o)
			return nil
		}

		return Reconcile(o)
	}
	return nil
}

func Flush(cluster *v1alpha1.Elasticsearch) {
	logrus.Infof("Flushing nodes for cluster %v in %v", cluster.Name, cluster.Namespace)
	k8shandler.FlushNodes(cluster.Name, cluster.Namespace)
}

// Reconcile reconciles the cluster's state to the spec specified
func Reconcile(cluster *v1alpha1.Elasticsearch) (err error) {

	if cluster.Spec.ManagementState == v1alpha1.ManagementStateUnmanaged {
		return nil
	}

	// Ensure existence of servicesaccount
	if err = k8shandler.CreateOrUpdateServiceAccount(cluster); err != nil {
		return fmt.Errorf("Failed to reconcile ServiceAccount for Elasticsearch cluster: %v", err)
	}

	// Ensure existence of clusterroles and clusterrolebindings
	if err := k8shandler.CreateOrUpdateRBAC(cluster); err != nil {
		return fmt.Errorf("Failed to reconcile Roles and RoleBindings for Elasticsearch cluster: %v", err)
	}

	// Ensure existence of config maps
	if err = k8shandler.CreateOrUpdateConfigMaps(cluster); err != nil {
		return fmt.Errorf("Failed to reconcile ConfigMaps for Elasticsearch cluster: %v", err)
	}

	if err = k8shandler.CreateOrUpdateServices(cluster); err != nil {
		return fmt.Errorf("Failed to reconcile Services for Elasticsearch cluster: %v", err)
	}

	// Ensure Elasticsearch cluster itself is up to spec
	//if err = k8shandler.CreateOrUpdateElasticsearchCluster(cluster, "elasticsearch", "elasticsearch"); err != nil {
	if err = k8shandler.CreateOrUpdateElasticsearchCluster(cluster); err != nil {
		return fmt.Errorf("Failed to reconcile Elasticsearch deployment spec: %v", err)
	}

	// Ensure existence of service monitors
	if err = k8shandler.CreateOrUpdateServiceMonitors(cluster); err != nil {
		return fmt.Errorf("Failed to reconcile Service Monitors for Elasticsearch cluster: %v", err)
	}

	// Ensure existence of prometheus rules
	if err = k8shandler.CreateOrUpdatePrometheusRules(cluster); err != nil {
		return fmt.Errorf("Failed to reconcile PrometheusRules for Elasticsearch cluster: %v", err)
	}

	return nil
}
