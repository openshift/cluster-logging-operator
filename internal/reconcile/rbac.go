package reconcile

import (
	"context"
	"fmt"

	log "github.com/ViaQ/logerr/v2/log/static"
	"github.com/openshift/cluster-logging-operator/internal/runtime"
	rbacv1 "k8s.io/api/rbac/v1"
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
	roleBinding := runtime.NewRoleBinding(desired.Namespace, desired.Name, desired.RoleRef)
	op, err := controllerutil.CreateOrUpdate(context.TODO(), k8Client, roleBinding, func() error {
		// Update the rolebinding with our desired state
		roleBinding.RoleRef = desired.RoleRef
		roleBinding.Subjects = desired.Subjects
		roleBinding.OwnerReferences = desired.OwnerReferences
		return nil
	})

	if err == nil {
		log.V(3).Info(fmt.Sprintf("reconciled roleBinding - operation: %s", op))
	}
	return err
}

func ClusterRoleBinding(k8sClient client.Client, name string, generator func() *rbacv1.ClusterRoleBinding) error {
	wantRoleBinding := generator()
	crb := runtime.NewClusterRoleBinding(name, rbacv1.RoleRef{})
	op, err := controllerutil.CreateOrUpdate(context.TODO(), k8sClient, crb, func() error {
		// Update the clusterrolebinding with our desired state
		crb.RoleRef = wantRoleBinding.RoleRef
		crb.Subjects = wantRoleBinding.Subjects
		return nil
	})

	if err == nil {
		log.V(3).Info(fmt.Sprintf("reconciled clusterRoleBinding - operation: %s", op))
	}
	return err
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
