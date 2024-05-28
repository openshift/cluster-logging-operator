package runtime

import (
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// NewNamespace returns a corev1.Namespace with name.
func NewNamespace(name string) *corev1.Namespace {
	ns := &corev1.Namespace{}
	Initialize(ns, "", name)
	return ns
}

// NewConfigMap returns a corev1.ConfigMap with namespace, name and data.
func NewConfigMap(namespace, name string, data map[string]string, visitors ...func(o runtime.Object)) *corev1.ConfigMap {
	if data == nil {
		data = map[string]string{}
	}
	cm := &corev1.ConfigMap{Data: data}
	Initialize(cm, namespace, name, visitors...)
	return cm
}

// NewPod returns a corev1.Pod with namespace, name, containers.
func NewPod(namespace, name string, containers ...corev1.Container) *corev1.Pod {
	pod := &corev1.Pod{Spec: corev1.PodSpec{Containers: containers}}
	Initialize(pod, namespace, name)
	return pod
}

// NewContainer stubs an instance of a corev1.Container
func NewContainer(name string, image string, pullPolicy corev1.PullPolicy, resources *corev1.ResourceRequirements) *corev1.Container {

	container := &corev1.Container{
		Name:            name,
		Image:           image,
		ImagePullPolicy: pullPolicy,
	}
	if resources != nil {
		container.Resources = *resources
	}
	return container
}

// NewService returns a corev1.Service with namespace and name.
func NewService(namespace, name string, visitors ...func(o runtime.Object)) *corev1.Service {
	svc := &corev1.Service{}
	Initialize(svc, namespace, name, visitors...)
	return svc
}

// NewServiceAccount returns a corev1.ServiceAccount with namespace and name.
func NewServiceAccount(namespace, name string) *corev1.ServiceAccount {
	obj := &corev1.ServiceAccount{}
	Initialize(obj, namespace, name)
	return obj
}

// NewSecret returns a corev1.Secret with namespace and name.
func NewSecret(namespace, name string, data map[string][]byte, visitors ...func(o runtime.Object)) *corev1.Secret {
	if data == nil {
		data = map[string][]byte{}
	}
	s := &corev1.Secret{Data: data}
	Initialize(s, namespace, name, visitors...)
	return s
}

// NewDaemonSet returns a daemon set.
func NewDaemonSet(namespace, name string, visitors ...func(o runtime.Object)) *appsv1.DaemonSet {
	ds := &appsv1.DaemonSet{}
	Initialize(ds, namespace, name, visitors...)
	return ds
}

// NewDeployment returns a deployment
func NewDeployment(namespace, name string, visitors ...func(o runtime.Object)) *appsv1.Deployment {
	dpl := &appsv1.Deployment{}
	Initialize(dpl, namespace, name, visitors...)
	return dpl
}
