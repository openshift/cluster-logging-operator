package runtime

import (
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
)

// NewNamespace returns a corev1.Namespace with name.
func NewNamespace(name string) *corev1.Namespace {
	ns := &corev1.Namespace{}
	Initialize(ns, "", name)
	return ns
}

// NewConfigMap returns a corev1.ConfigMap with namespace, name and data.
func NewConfigMap(namespace, name string, data map[string]string) *corev1.ConfigMap {
	if data == nil {
		data = map[string]string{}
	}
	cm := &corev1.ConfigMap{Data: data}
	Initialize(cm, namespace, name)
	return cm
}

// NewPod returns a corev1.Pod with namespace, name, containers.
func NewPod(namespace, name string, containers ...corev1.Container) *corev1.Pod {
	pod := &corev1.Pod{Spec: corev1.PodSpec{Containers: containers}}
	Initialize(pod, namespace, name)
	return pod
}

// NewService returns a corev1.Service with namespace and name.
func NewService(namespace, name string) *corev1.Service {
	svc := &corev1.Service{}
	Initialize(svc, namespace, name)
	return svc
}

// NewServiceAccount returns a corev1.ServiceAccount with namespace and name.
func NewServiceAccount(namespace, name string) *corev1.ServiceAccount {
	obj := &corev1.ServiceAccount{}
	Initialize(obj, namespace, name)
	return obj
}

// NewSecret returns a corev1.Secret with namespace and name.
func NewSecret(namespace, name string, data map[string][]byte) *corev1.Secret {
	if data == nil {
		data = map[string][]byte{}
	}
	s := &corev1.Secret{Data: data}
	Initialize(s, namespace, name)
	return s
}

// NewDaemonSet returns a daemon set.
func NewDaemonSet(namespace, name string) *appsv1.DaemonSet {
	ds := &appsv1.DaemonSet{}
	Initialize(ds, namespace, name)
	return ds
}

// NewDeployment returns a daemon set.
func NewDeployment(namespace, name string) *appsv1.Deployment {
	dep := &appsv1.Deployment{}
	Initialize(dep, namespace, name)
	return dep
}

//NewRole returns a role with namespace, names, rules
func NewRole(namespace, name string, rules ...rbacv1.PolicyRule) *rbacv1.Role {
	role := &rbacv1.Role{
		Rules: rules,
	}
	Initialize(role, namespace, name)
	return role
}

//NewClusterRole returns a role with namespace, names, rules
func NewClusterRole(name string, rules ...rbacv1.PolicyRule) *rbacv1.ClusterRole {
	role := &rbacv1.ClusterRole{
		Rules: rules,
	}
	Initialize(role, "", name)
	return role
}

//NewRoleBinding returns a role with namespace, names, rules
func NewRoleBinding(namespace, name string, roleRef rbacv1.RoleRef, subjects ...rbacv1.Subject) *rbacv1.RoleBinding {
	binding := &rbacv1.RoleBinding{
		RoleRef:  roleRef,
		Subjects: subjects,
	}
	Initialize(binding, namespace, name)
	return binding
}

//NewClusterRoleBinding returns a role with namespace, names, rules
func NewClusterRoleBinding(name string, roleRef rbacv1.RoleRef, subjects ...rbacv1.Subject) *rbacv1.ClusterRoleBinding {
	binding := &rbacv1.ClusterRoleBinding{
		RoleRef:  roleRef,
		Subjects: subjects,
	}
	Initialize(binding, "", name)
	return binding
}
