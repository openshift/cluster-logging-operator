package reconcile

import (
	"context"
	"fmt"

	log "github.com/ViaQ/logerr/v2/log/static"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	util "github.com/openshift/cluster-logging-operator/internal/utils"
	"github.com/openshift/cluster-logging-operator/internal/utils/comparators/servicemonitor"
	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/retry"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// ServiceMonitor reconciles a ServiceMonitor to the desired spec returning an error
// if there is an issue creating or updating to the desired state
func ServiceMonitor(er record.EventRecorder, k8Client client.Client, desired *monitoringv1.ServiceMonitor) error {
	reason := constants.EventReasonGetObject
	retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		current := &monitoringv1.ServiceMonitor{}
		key := client.ObjectKeyFromObject(desired)
		if err := k8Client.Get(context.TODO(), key, current); err != nil {
			if errors.IsNotFound(err) {
				reason = constants.EventReasonCreateObject
				return k8Client.Create(context.TODO(), desired)
			}
			return fmt.Errorf("failed to get %v servicemonitor: %w", key, err)
		}
		if servicemonitor.AreSame(current, desired) && util.HasSameOwner(current.OwnerReferences, desired.OwnerReferences) {
			log.V(3).Info("ServiceMonitor are the same skipping update")
			return nil
		}
		reason = constants.EventReasonUpdateObject
		current.Labels = desired.Labels
		current.Spec = desired.Spec
		current.Annotations = desired.Annotations
		current.OwnerReferences = desired.OwnerReferences

		return k8Client.Update(context.TODO(), current)
	})
	eventType := v1.EventTypeNormal
	msg := fmt.Sprintf("%s ServiceMonitor %s/%s", eventType, desired.Namespace, desired.Name)
	if retryErr != nil {
		eventType = v1.EventTypeWarning
		msg = fmt.Sprintf("Unable to %s: %v", msg, retryErr)
	}
	er.Event(desired, eventType, reason, msg)
	return retryErr
}
