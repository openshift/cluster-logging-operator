package k8shandler

import (
	"context"
	"fmt"

	log "github.com/ViaQ/logerr/v2/log/static"
	logging "github.com/openshift/cluster-logging-operator/api/logging/v1"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	"github.com/openshift/cluster-logging-operator/internal/k8s/loader"
	"github.com/openshift/cluster-logging-operator/internal/utils"
	"github.com/openshift/cluster-logging-operator/internal/visualization/console"
	"github.com/openshift/cluster-logging-operator/internal/visualization/kibana"
	es "github.com/openshift/elasticsearch-operator/apis/logging/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	utilerrors "k8s.io/apimachinery/pkg/util/errors"
	"k8s.io/client-go/util/retry"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// CreateOrUpdateVisualization reconciles visualization (kibana or console log view) component for cluster logging
func (clusterRequest *ClusterLoggingRequest) CreateOrUpdateVisualization() error {
	if !clusterRequest.isManaged() {
		return nil
	}
	var errs []error
	spec := clusterRequest.Cluster.Spec

	if spec.Visualization != nil && spec.Visualization.Type == logging.VisualizationTypeKibana {
		errs = append(errs, clusterRequest.createOrUpdateKibana())
	}

	errs = append(errs, console.ReconcilePlugin(clusterRequest.Client, clusterRequest.Cluster, clusterRequest.Cluster, clusterRequest.ClusterVersion))
	return utilerrors.NewAggregate(errs)
}

func (clusterRequest *ClusterLoggingRequest) createOrUpdateKibana() (err error) {
	if clusterRequest.Cluster.Spec.Visualization == nil || clusterRequest.Cluster.Spec.Visualization.Type == "" {
		return clusterRequest.removeKibana()
	}

	cluster := clusterRequest.Cluster
	visSpec := cluster.Spec.Visualization
	cr := kibana.New(cluster.Namespace, constants.KibanaName, visSpec, cluster.Spec.LogStore, utils.AsOwner(cluster))
	if err = kibana.Reconcile(clusterRequest.EventRecorder, clusterRequest.Client, cr); err != nil {
		return
	}

	if err = clusterRequest.UpdateKibanaStatus(); err != nil {
		return
	}

	return nil
}

func (clusterRequest *ClusterLoggingRequest) removeKibana() (err error) {
	cluster := clusterRequest.Cluster
	cr := kibana.New(cluster.Namespace, constants.KibanaName, &logging.VisualizationSpec{}, cluster.Spec.LogStore, utils.AsOwner(cluster))

	err = clusterRequest.Client.Delete(context.TODO(), cr)
	if err != nil && !errors.IsNotFound(err) && !meta.IsNoMatchError(err) {
		return fmt.Errorf("unable to delete kibana cr. E: %s", err.Error())
	}
	return nil
}

func (clusterRequest *ClusterLoggingRequest) UpdateKibanaStatus() (err error) {
	kibanaStatus, err := clusterRequest.getKibanaStatus()
	if err != nil {
		log.Error(err, "Failed to get Kibana status for", "clusterName", clusterRequest.Cluster.Name)
		return
	}

	retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		instance, err, _ := loader.FetchClusterLogging(clusterRequest.Client, clusterRequest.Cluster.Namespace, clusterRequest.Cluster.Name, true)
		if err != nil {
			return err
		}

		if !kibana.CompareStatus(kibanaStatus, instance.Status.Visualization.KibanaStatus) {
			instance.Status.Visualization.KibanaStatus = kibanaStatus
			return clusterRequest.UpdateStatus(&instance)
		}
		return nil
	})
	if retryErr != nil {
		return fmt.Errorf("Failed to update Cluster Logging Kibana status: %v", retryErr)
	}
	return nil
}

func (clusterRequest *ClusterLoggingRequest) getKibanaCR() (*es.Kibana, error) {
	kb := &es.Kibana{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Kibana",
			APIVersion: es.GroupVersion.String(),
		},
	}

	err := clusterRequest.Client.Get(context.TODO(),
		client.ObjectKey{
			Namespace: clusterRequest.Cluster.Namespace,
			Name:      constants.KibanaName,
		}, kb)
	if err != nil {
		return nil, err
	}
	return kb, nil
}
