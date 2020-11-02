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
	"fmt"
	"os/exec"
	"strings"

	loggingv1 "github.com/openshift/cluster-logging-operator/pkg/apis/logging/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
)

// Object is an alias for this central type.
type Object = runtime.Object

// Codecs is a codec factory for the default scheme, including core and our custom types.
var Codecs = serializer.NewCodecFactory(scheme.Scheme)

func init() {
	must(loggingv1.AddToScheme(scheme.Scheme)) // Add our types
}

// Decode JSON or YAML resource manifest to a new typed struct.
func Decode(manifest string) runtime.Object {
	d := Codecs.UniversalDeserializer()
	o, _, err := d.Decode([]byte(manifest), nil, nil)
	must(err)
	return o
}

// Meta interface to get/set object metadata.
func Meta(o runtime.Object) metav1.Object {
	m, err := meta.Accessor(o)
	must(err)
	return m
}

// NamespacedName returns the namespaced name of an object.
func NamespacedName(o runtime.Object) types.NamespacedName {
	nn, err := client.ObjectKeyFromObject(o)
	must(err)
	return nn
}

// ID returns a human-readable identifier for the object.
func ID(o runtime.Object) string {
	m := Meta(o)
	return fmt.Sprintf("[%v/%v, Namespace=%v]", strings.ToLower(GroupVersionKind(o).Kind), m.GetName(), m.GetNamespace())
}

// GroupVersionKind deduces the Kind from the Go type.
func GroupVersionKind(o runtime.Object) schema.GroupVersionKind {
	gvk, err := apiutil.GVKForObject(o, scheme.Scheme)
	must(err)
	return gvk
}

// Labels returns the labels map for object, guaratneed to be non-nil.
func Labels(o runtime.Object) map[string]string {
	m := Meta(o)
	l := m.GetLabels()
	if l == nil {
		l = map[string]string{}
		m.SetLabels(l)
	}
	return l
}

// Initialize sets name, namespace and type metadata deduced from Go type.
func Initialize(o runtime.Object, namespace, name string) {
	m := Meta(o)
	m.SetNamespace(namespace)
	m.SetName(name)
	o.GetObjectKind().SetGroupVersionKind(GroupVersionKind(o))
}

// Exec returns an `oc exec` Cmd to run cmd on o.
func Exec(o runtime.Object, cmd string, args ...string) *exec.Cmd {
	m := Meta(o)
	ocCmd := append([]string{
		"exec",
		"-i",
		"-n", m.GetNamespace(),
		GroupVersionKind(o).Kind + "/" + m.GetName(),
		"--",
		cmd,
	}, args...)

	return exec.Command("oc", ocCmd...)
}

// ServiceDomainName returns "name.namespace.svc".
func ServiceDomainName(o runtime.Object) string {
	m := Meta(o)
	return fmt.Sprintf("%s.%s.svc", m.GetName(), m.GetNamespace())
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}
