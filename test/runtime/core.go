// Licensed to Red Hat, Inc under one or more contributor
// license agreements. See the NOTICE file distributed with
// this work for additional information regarding copyright
// ownership. Red Hat, Inc licenses this file to you under
// the Apache License, Version 2.0 (the "License"); you may
// not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

package runtime

import (
	"github.com/openshift/cluster-logging-operator/test"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

// NewNamespace returns a corev1.Namespace with name.
func NewNamespace(name string) *corev1.Namespace {
	ns := &corev1.Namespace{}
	Initialize(ns, "", name)
	return ns
}

// NewUniqueNamespace returns a corev1.Namespace with unique name.
func NewUniqueNamespace() *corev1.Namespace {
	return NewNamespace(test.UniqueNameForTest())
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
