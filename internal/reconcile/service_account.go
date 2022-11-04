package reconcile

import (
	"context"
	"fmt"
	log "github.com/ViaQ/logerr/v2/log/static"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	"github.com/openshift/cluster-logging-operator/internal/utils"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/retry"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func ServiceAccount(er record.EventRecorder, k8Client client.Client, desired *v1.ServiceAccount) (*v1.ServiceAccount, error) {
	reason := constants.EventReasonGetObject
	updateReason := ""
	current := &v1.ServiceAccount{}
	retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		key := client.ObjectKeyFromObject(desired)
		if err := k8Client.Get(context.TODO(), key, current); err != nil {
			if errors.IsNotFound(err) {
				reason = constants.EventReasonCreateObject
				current = desired
				return k8Client.Create(context.TODO(), desired)
			}
			return fmt.Errorf("failed to get %v ServiceAccount: %w", key, err)
		}

		same := false
		if same = utils.AreMapsSame(current.Annotations, desired.Annotations); same {
			log.V(3).Info("ServiceAccount are the same skipping update")
			return nil
		}
		reason = constants.EventReasonUpdateObject
		if current.Annotations == nil {
			current.Annotations = map[string]string{}
		}
		if desired.Annotations != nil {
			for key, value := range desired.Annotations {
				current.Annotations[key] = value
			}
		}
		return k8Client.Update(context.TODO(), current)
	})

	eventType := v1.EventTypeNormal
	msg := fmt.Sprintf("%s ServiceAccount %s/%s", reason, desired.Namespace, desired.Name)
	if updateReason != "" {
		msg = fmt.Sprintf("%s because of change in %s.", msg, updateReason)
	}
	if retryErr != nil {
		eventType = v1.EventTypeWarning
		msg = fmt.Sprintf("Unable to %s: %v", msg, retryErr)
	}
	er.Event(desired, eventType, reason, msg)
	return current, retryErr
}
