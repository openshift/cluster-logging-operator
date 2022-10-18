package reconcile

import (
	"context"
	"fmt"
	log "github.com/ViaQ/logerr/v2/log/static"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	"github.com/openshift/cluster-logging-operator/internal/visualization/kibana"
	loggingv1 "github.com/openshift/elasticsearch-operator/apis/logging/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/retry"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func ReconcileKibana(er record.EventRecorder, k8Client client.Client, desired *loggingv1.Kibana) error {
	reason := constants.EventReasonGetObject
	updateReason := ""
	retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		current := &loggingv1.Kibana{}
		key := client.ObjectKeyFromObject(desired)
		if err := k8Client.Get(context.TODO(), key, current); err != nil {
			if errors.IsNotFound(err) {
				reason = constants.EventReasonCreateObject
				return k8Client.Create(context.TODO(), desired)
			}
			return fmt.Errorf("failed to get %v Kibana: %w", key, err)
		}
		same := false

		if same, updateReason = kibana.AreSame(*current, *desired); same {
			log.V(3).Info("Kibana are the same skipping update")
			return nil
		}
		reason = constants.EventReasonUpdateObject
		current.Spec = desired.Spec
		return k8Client.Update(context.TODO(), current)
	})

	eventType := v1.EventTypeNormal
	msg := fmt.Sprintf("%s Kibana %s/%s", reason, desired.Namespace, desired.Name)
	if updateReason != "" {
		msg = fmt.Sprintf("%s because of change in %s.", msg, updateReason)
	}
	if retryErr != nil {
		eventType = v1.EventTypeWarning
		msg = fmt.Sprintf("Unable to %s: %v", msg, retryErr)
	}
	er.Event(desired, eventType, reason, msg)
	return retryErr
}
