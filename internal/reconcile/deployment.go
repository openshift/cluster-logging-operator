package reconcile

import (
	"context"
	"fmt"

	log "github.com/ViaQ/logerr/v2/log/static"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	util "github.com/openshift/cluster-logging-operator/internal/utils"
	"github.com/openshift/cluster-logging-operator/internal/utils/comparators/pod"
	apps "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/retry"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Deployment reconciles a Deployment to the desired spec returning an error
// if there is an issue creating or updating to the desired state
func Deployment(er record.EventRecorder, k8Client client.Client, desired *apps.Deployment) error {
	reason := constants.EventReasonGetObject
	updateReason := ""
	retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		current := &apps.Deployment{}
		key := client.ObjectKeyFromObject(desired)
		if err := k8Client.Get(context.TODO(), key, current); err != nil {
			if errors.IsNotFound(err) {
				reason = constants.EventReasonCreateObject
				return k8Client.Create(context.TODO(), desired)
			}
			return fmt.Errorf("failed to get %v Deployment: %w", key, err)
		}
		same := false

		if same, updateReason = pod.AreSame(&current.Spec.Template.Spec, &desired.Spec.Template.Spec, current.Name); same && util.HasSameOwner(current.OwnerReferences, desired.OwnerReferences) {
			log.V(3).Info("Deployments are the same skipping update", "deploymentName", current.Name)
			return nil
		}
		reason = constants.EventReasonUpdateObject
		current.Labels = desired.Labels
		current.Spec = desired.Spec
		current.OwnerReferences = desired.OwnerReferences
		return k8Client.Update(context.TODO(), current)
	})

	eventType := v1.EventTypeNormal
	msg := fmt.Sprintf("%s Deployment %s/%s", reason, desired.Namespace, desired.Name)
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
