package reconcile

import (
	"context"
	"fmt"

	log "github.com/ViaQ/logerr/v2/log/static"
	"github.com/openshift/cluster-logging-operator/internal/runtime"
	rbacv1 "k8s.io/api/rbac/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

func Role(k8Client client.Client, desired *rbacv1.Role) error {
	role := runtime.NewRole(desired.Namespace, desired.Name)
	op, err := controllerutil.CreateOrUpdate(context.TODO(), k8Client, role, func() error {
		// Update the role with our desired state
		role.Rules = desired.Rules
		role.OwnerReferences = desired.OwnerReferences
		return nil
	})

	if err == nil {
		log.V(3).Info(fmt.Sprintf("reconciled role - operation: %s", op))
	}
	return err
}

func RoleBinding(k8Client client.Client, desired *rbacv1.RoleBinding) error {
	existing := runtime.NewRoleBinding(desired.Namespace, desired.Name, rbacv1.RoleRef{})
	err := k8Client.Get(context.TODO(), client.ObjectKeyFromObject(existing), existing)
	if apierrors.IsNotFound(err) {
		log.V(3).Info("Creating roleBinding", "name", desired.Name, "namespace", desired.Namespace)
		return k8Client.Create(context.TODO(), desired)
	}
	if err != nil {
		return err
	}

	if existing.RoleRef != desired.RoleRef {
		log.V(3).Info("Deleting roleBinding due to roleRef change", "name", desired.Name, "namespace", desired.Namespace)
		if err := k8Client.Delete(context.TODO(), existing); err != nil {
			return err
		}
		log.V(3).Info("Recreating roleBinding", "name", desired.Name, "namespace", desired.Namespace)
		return k8Client.Create(context.TODO(), desired)
	}

	existing.Subjects = desired.Subjects
	existing.OwnerReferences = desired.OwnerReferences
	log.V(3).Info("Updating roleBinding", "name", desired.Name, "namespace", desired.Namespace)
	return k8Client.Update(context.TODO(), existing)
}

func ClusterRoleBinding(k8sClient client.Client, name string, generator func() *rbacv1.ClusterRoleBinding) error {
	desired := generator()
	existing := runtime.NewClusterRoleBinding(name, rbacv1.RoleRef{})
	err := k8sClient.Get(context.TODO(), client.ObjectKeyFromObject(existing), existing)
	if apierrors.IsNotFound(err) {
		log.V(3).Info("Creating clusterRoleBinding", "name", name)
		return k8sClient.Create(context.TODO(), desired)
	}
	if err != nil {
		return err
	}

	if existing.RoleRef != desired.RoleRef {
		log.V(3).Info("Deleting clusterRoleBinding due to roleRef change", "name", name)
		if err := k8sClient.Delete(context.TODO(), existing); err != nil {
			return err
		}
		log.V(3).Info("Recreating clusterRoleBinding", "name", name)
		return k8sClient.Create(context.TODO(), desired)
	}

	existing.Subjects = desired.Subjects
	log.V(3).Info("Updating clusterRoleBinding", "name", name)
	return k8sClient.Update(context.TODO(), existing)
}

func DeleteClusterRoleBinding(k8sClient client.Client, name string) error {
	object := runtime.NewClusterRoleBinding(name, rbacv1.RoleRef{})
	log.V(3).Info("Deleting ClusterRoleBinding", "name", name)
	err := k8sClient.Delete(context.TODO(), object)
	// Ignore NotFound errors - resource is already deleted
	if apierrors.IsNotFound(err) {
		return nil
	}
	return err
}

func DeleteRole(k8sClient client.Client, namespace, name string) error {
	object := runtime.NewRole(namespace, name)
	log.V(3).Info("Deleting Role", "namespace", namespace, "name", name)
	err := k8sClient.Delete(context.TODO(), object)
	// Ignore NotFound errors - resource is already deleted
	if apierrors.IsNotFound(err) {
		return nil
	}
	return err
}
