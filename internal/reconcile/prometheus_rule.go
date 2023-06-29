package reconcile

import (
	"context"
	"fmt"
	"reflect"

	log "github.com/ViaQ/logerr/v2/log/static"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	"github.com/openshift/cluster-logging-operator/internal/utils"
	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/retry"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// PrometheusRule reconciles a PrometheusRule to the desired spec returning an error
// if there is an issue creating or updating to the desired state
func PrometheusRule(er record.EventRecorder, k8Client client.Client, desired *monitoringv1.PrometheusRule) error {
	reason := constants.EventReasonGetObject
	updateReason := "spec"
	retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		current := &monitoringv1.PrometheusRule{}
		key := client.ObjectKeyFromObject(desired)
		if err := k8Client.Get(context.TODO(), key, current); err != nil {
			if errors.IsNotFound(err) {
				reason = constants.EventReasonCreateObject
				return k8Client.Create(context.TODO(), desired)
			}
			return fmt.Errorf("failed to get %v Secret: %w", key, err)
		}
		if reflect.DeepEqual(current.Spec, desired.Spec) && utils.HasSameOwner(current.OwnerReferences, desired.OwnerReferences) {
			// identical; no need to update.
			log.V(3).Info("PrometheusRule are the same skipping update")
			return nil
		}
		current.Spec = desired.Spec
		current.OwnerReferences = desired.OwnerReferences
		reason = constants.EventReasonUpdateObject
		return k8Client.Update(context.TODO(), current)
	})

	eventType := corev1.EventTypeNormal
	msg := fmt.Sprintf("%s PrometheusRule %s/%s", reason, desired.Namespace, desired.Name)
	if updateReason != "" {
		msg = fmt.Sprintf("%s because of change in %s.", msg, updateReason)
	}
	if retryErr != nil {
		eventType = corev1.EventTypeWarning
		msg = fmt.Sprintf("Unable to %s: %v", msg, retryErr)
	}
	er.Event(desired, eventType, reason, msg)
	return retryErr
}
