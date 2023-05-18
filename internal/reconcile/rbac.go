package reconcile

import (
	"context"
	"fmt"

	log "github.com/ViaQ/logerr/v2/log/static"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	"github.com/openshift/cluster-logging-operator/internal/runtime"
	"github.com/openshift/cluster-logging-operator/internal/utils"
	"github.com/openshift/cluster-logging-operator/internal/utils/comparators/clusterrole"
	"github.com/openshift/cluster-logging-operator/internal/utils/comparators/clusterrolebinding"
	"github.com/openshift/cluster-logging-operator/internal/utils/comparators/role"
	"github.com/openshift/cluster-logging-operator/internal/utils/comparators/rolebinding"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/retry"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func Role(er record.EventRecorder, k8Client client.Client, desired *rbacv1.Role) error {
	reason := constants.EventReasonGetObject

	retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		current := &rbacv1.Role{}
		key := client.ObjectKeyFromObject(desired)
		if err := k8Client.Get(context.TODO(), key, current); err != nil {
			if apierrors.IsNotFound(err) {
				reason = constants.EventReasonCreateObject
				return k8Client.Create(context.TODO(), desired)
			}
			return fmt.Errorf("failed to get %v role: %w", key, err)
		}
		if role.AreSame(current, desired) && utils.HasSameOwner(current.OwnerReferences, desired.OwnerReferences) {
			log.V(3).Info("Roles are the same skipping update")
			return nil
		}
		reason = constants.EventReasonUpdateObject
		current.Rules = desired.Rules
		current.OwnerReferences = desired.OwnerReferences

		return k8Client.Update(context.TODO(), current)
	})
	eventType := corev1.EventTypeNormal
	msg := fmt.Sprintf("%s Role %s/%s", eventType, desired.Namespace, desired.Name)
	if retryErr != nil {
		eventType = corev1.EventTypeWarning
		msg = fmt.Sprintf("Unable to %s: %v", msg, retryErr)
	}
	er.Event(desired, eventType, reason, msg)
	return retryErr
}

func RoleBinding(er record.EventRecorder, k8Client client.Client, desired *rbacv1.RoleBinding) error {
	reason := constants.EventReasonGetObject

	retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		current := &rbacv1.RoleBinding{}
		key := client.ObjectKeyFromObject(desired)
		if err := k8Client.Get(context.TODO(), key, current); err != nil {
			if apierrors.IsNotFound(err) {
				reason = constants.EventReasonCreateObject
				return k8Client.Create(context.TODO(), desired)
			}
			return fmt.Errorf("failed to get %v rolebinding: %w", key, err)
		}
		if rolebinding.AreSame(current, desired) && utils.HasSameOwner(current.OwnerReferences, desired.OwnerReferences) {
			log.V(3).Info("RoleBindings are the same skipping update")
			return nil
		}
		reason = constants.EventReasonUpdateObject

		current.RoleRef = desired.RoleRef
		current.Subjects = desired.Subjects
		current.OwnerReferences = desired.OwnerReferences

		return k8Client.Update(context.TODO(), current)
	})
	eventType := corev1.EventTypeNormal
	msg := fmt.Sprintf("%s RoleBinding %s/%s", eventType, desired.Namespace, desired.Name)
	if retryErr != nil {
		eventType = corev1.EventTypeWarning
		msg = fmt.Sprintf("Unable to %s: %v", msg, retryErr)
	}
	er.Event(desired, eventType, reason, msg)
	return retryErr
}

func ClusterRole(k8sClient client.Client, name string, generator func() *rbacv1.ClusterRole) error {
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

	return retryErr
}

func ClusterRoleBinding(k8sClient client.Client, name string, generator func() *rbacv1.ClusterRoleBinding) error {
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
	return retryErr
}

func DeleteClusterRole(k8sClient client.Client, name string) error {
	object := runtime.NewClusterRole(name)
	log.V(3).Info("Deleting", "object", object)
	return k8sClient.Delete(context.TODO(), object)
}

func DeleteClusterRoleBinding(k8sClient client.Client, name string) error {
	object := runtime.NewClusterRoleBinding(name, rbacv1.RoleRef{})
	log.V(3).Info("Deleting", "object", object)
	return k8sClient.Delete(context.TODO(), object)
}
