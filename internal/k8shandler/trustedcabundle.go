package k8shandler

import (
	"context"
	"fmt"
	"k8s.io/client-go/util/retry"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/openshift/cluster-logging-operator/internal/constants"
	"github.com/openshift/cluster-logging-operator/internal/utils"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
)

//createOrGetTrustedCABundleConfigMap creates or returns an existing Trusted CA Bundle ConfigMap.
//By setting label "config.openshift.io/inject-trusted-cabundle: true", the cert is automatically filled/updated.
func (clusterRequest *ClusterLoggingRequest) createOrGetTrustedCABundleConfigMap(name string) (*corev1.ConfigMap, error) {
	desired := NewConfigMap(
		name,
		clusterRequest.Cluster.Namespace,
		map[string]string{
			constants.TrustedCABundleKey: "",
		},
	)
	desired.ObjectMeta.Labels = make(map[string]string)
	desired.ObjectMeta.Labels[constants.InjectTrustedCABundleLabel] = "true"

	utils.AddOwnerRefToObject(desired, utils.AsOwner(clusterRequest.Cluster))

	reason := constants.EventReasonGetObject
	var current *corev1.ConfigMap
	retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		current = &corev1.ConfigMap{}
		key := client.ObjectKeyFromObject(desired)
		if err := clusterRequest.Client.Get(context.TODO(), key, current); err != nil {
			if errors.IsNotFound(err) {
				reason = constants.EventReasonCreateObject
				return clusterRequest.Client.Create(context.TODO(), desired)
			}
			return fmt.Errorf("failed to get %v ConfigMap: %w", key, err)
		}
		if val := current.ObjectMeta.Labels[constants.InjectTrustedCABundleLabel]; val == "true" {
			return nil
		}
		if current.ObjectMeta.Labels == nil {
			current.ObjectMeta.Labels = map[string]string{}
		}
		current.ObjectMeta.Labels[constants.InjectTrustedCABundleLabel] = "true"
		reason = constants.EventReasonUpdateObject
		return clusterRequest.Client.Update(context.TODO(), current)
	})
	eventType := corev1.EventTypeNormal
	msg := fmt.Sprintf("%s ConfigMap %s/%s", reason, desired.Namespace, desired.Name)
	if retryErr != nil {
		eventType = corev1.EventTypeWarning
		msg = fmt.Sprintf("Unable to %s: %v", msg, retryErr)
	}
	clusterRequest.EventRecorder.Event(desired, eventType, reason, msg)
	return current, retryErr
}

func calcTrustedCAHashValue(configMap *corev1.ConfigMap) (string, error) {
	hashValue := ""
	var err error

	if configMap == nil {
		return hashValue, nil
	}
	caBundle, ok := configMap.Data[constants.TrustedCABundleKey]
	if ok && caBundle != "" {
		hashValue, err = utils.CalculateMD5Hash(caBundle)
		if err != nil {
			return "", err
		}
	}

	if !ok {
		return "", fmt.Errorf("Expected key %v does not exist in %v", constants.TrustedCABundleKey, configMap.Name)
	}

	return hashValue, nil
}
