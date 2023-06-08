package elasticsearch

import (
	"context"
	"fmt"
	log "github.com/ViaQ/logerr/v2/log/static"
	logging "github.com/openshift/cluster-logging-operator/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/util/retry"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func Reconcile(k8sClient client.Client, logStore *logging.LogStoreSpec, namespace string, ownerRef v1.OwnerReference, fetchClusterLogging func() (*logging.ClusterLogging, error)) error {
	if err := ReconcileCustomResource(k8sClient, logStore, namespace, ownerRef); err != nil {
		return err
	}

	if err := UpdateStatus(k8sClient, namespace, fetchClusterLogging); err != nil {
		log.Error(err, "Failed to update Cluster Logging Elasticsearch status")
	}

	return nil
}

func ReconcileCustomResource(k8sClient client.Client, logStore *logging.LogStoreSpec, namespace string, ownerRef v1.OwnerReference) (err error) {

	retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		desired := NewElasticsearchCR(logStore, namespace, constants.ElasticsearchName, nil, ownerRef)
		current := NewEmptyElasticsearchCR(namespace, constants.ElasticsearchName)
		key := client.ObjectKeyFromObject(current)
		if err = k8sClient.Get(context.TODO(), key, current); err != nil {
			if apierrors.IsNotFound(err) {
				return k8sClient.Create(context.TODO(), desired)
			}
			return fmt.Errorf("Failed to get Elasticsearch CR: %v", err)
		}
		desired = NewElasticsearchCR(logStore, namespace, constants.ElasticsearchName, current, ownerRef)
		if current, different := IsElasticsearchCRDifferent(current, desired); different {
			return k8sClient.Update(context.TODO(), current)
		}
		log.V(3).Info("Elasticsearch CR is the same as spec. skipping update", "name", current.Name)
		return nil

	})

	return retryErr
}
