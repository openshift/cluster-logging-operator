package reconcile

import (
	"context"
	"fmt"
	log "github.com/ViaQ/logerr/v2/log/static"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	"github.com/openshift/cluster-logging-operator/internal/utils/comparators/clusterrole"
	"github.com/openshift/cluster-logging-operator/internal/utils/comparators/clusterrolebinding"
	v1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/retry"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func ClusterRole(er record.EventRecorder, k8sClient client.Client, name string, generator func() *rbacv1.ClusterRole) error {
	reason := constants.EventReasonGetObject
	updateReason := ""
	retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		current := &rbacv1.ClusterRole{}
		key := client.ObjectKey{Name: name}
		if err := k8sClient.Get(context.TODO(), key, current); err != nil {
			if !apierrors.IsNotFound(err) {
				return fmt.Errorf("failed to get ClusterRole: %w", err)
			}

			current = generator()
			if err := k8sClient.Create(context.TODO(), current); err != nil {
				return fmt.Errorf("failed to create ClusterRole: %w", err)
			}

			return nil
		}

		wantRole := generator()
		if clusterrole.AreSame(current, wantRole) {
			log.V(9).Info("ClusterRole matches.")
			return nil
		}

		current.Rules = wantRole.Rules

		return k8sClient.Update(context.TODO(), current)
	})

	eventType := v1.EventTypeNormal
	msg := fmt.Sprintf("%s ClusterRole %s", reason, generator().Name)
	if updateReason != "" {
		msg = fmt.Sprintf("%s because of change in %s.", msg, updateReason)
	}
	if retryErr != nil {
		eventType = v1.EventTypeWarning
		msg = fmt.Sprintf("Unable to %s: %v", msg, retryErr)
	}
	er.Event(generator(), eventType, reason, msg)
	return retryErr
}

func ClusterRoleBinding(er record.EventRecorder, k8sClient client.Client, name string, generator func() *rbacv1.ClusterRoleBinding) error {
	reason := constants.EventReasonGetObject
	updateReason := ""
	retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		current := &rbacv1.ClusterRoleBinding{}
		key := client.ObjectKey{Name: name}
		if err := k8sClient.Get(context.TODO(), key, current); err != nil {
			if !apierrors.IsNotFound(err) {
				return fmt.Errorf("failed to get ClusterRoleBinding: %w", err)
			}

			current = generator()
			if err := k8sClient.Create(context.TODO(), current); err != nil {
				return fmt.Errorf("failed to create ClusterRoleBinding: %w", err)
			}

			return nil
		}

		wantRoleBinding := generator()
		if clusterrolebinding.AreSame(current, wantRoleBinding) {
			log.V(9).Info("ClusterRoleBinding matches.")
			return nil
		}

		current.RoleRef = wantRoleBinding.RoleRef
		current.Subjects = wantRoleBinding.Subjects

		return k8sClient.Update(context.TODO(), current)
	})
	eventType := v1.EventTypeNormal
	msg := fmt.Sprintf("%s ClusterRoleBinding %s", reason, generator().Name)
	if updateReason != "" {
		msg = fmt.Sprintf("%s because of change in %s.", msg, updateReason)
	}
	if retryErr != nil {
		eventType = v1.EventTypeWarning
		msg = fmt.Sprintf("Unable to %s: %v", msg, retryErr)
	}
	er.Event(generator(), eventType, reason, msg)
	return retryErr
}
