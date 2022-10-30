package reconcile

import (
	"context"
	"fmt"
	log "github.com/ViaQ/logerr/v2/log/static"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/retry"
	"reflect"
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
		if same, updateReason = areSame(current, desired); same {
			log.V(3).Info("ServiceAccount are the same skipping update")
			return nil
		}
		reason = constants.EventReasonUpdateObject
		current.Secrets = desired.Secrets
		current.ImagePullSecrets = desired.ImagePullSecrets
		current.AutomountServiceAccountToken = desired.AutomountServiceAccountToken
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

func areSame(current *v1.ServiceAccount, desired *v1.ServiceAccount) (bool, string) {
	if current.AutomountServiceAccountToken != desired.AutomountServiceAccountToken {
		return false, "automountServiceAccountToken"
	}
	if len(current.Secrets) != len(desired.Secrets) {
		return false, "len of secrets are not equal"
	}
	for i, secret := range current.Secrets {
		if reflect.DeepEqual(secret, desired.Secrets[0]) {
			return false, fmt.Sprintf("secret[%d]", i)
		}
	}
	if len(current.ImagePullSecrets) != len(desired.ImagePullSecrets) {
		return false, "len of imagePullSecrets are not equal"
	}
	for i, secret := range current.ImagePullSecrets {
		if reflect.DeepEqual(secret, desired.ImagePullSecrets[0]) {
			return false, fmt.Sprintf("imagePullSecrets[%d]", i)
		}
	}
	return true, ""
}
