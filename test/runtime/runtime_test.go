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

package runtime_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	"github.com/openshift/cluster-logging-operator/test"
	. "github.com/openshift/cluster-logging-operator/test/matchers"
	"github.com/openshift/cluster-logging-operator/test/runtime"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("Object", func() {
	var (
		nsFoo = runtime.NewNamespace("foo")
		clf   = runtime.NewClusterLogForwarder()
	)

	DescribeTable("Decode",
		func(manifest string, o runtime.Object) {
			got := runtime.Decode(manifest)
			Expect(got).To(EqualDiff(o), "%#v", manifest)
		},
		Entry("YAML string ns", test.YAMLString(nsFoo), nsFoo),
		Entry("JSON string ns", test.JSONLine(nsFoo), nsFoo),
		Entry("YAML string clf", test.YAMLString(clf), clf),
	)

	It("panics on bad manifest string", func() {
		Expect(func() { _ = runtime.Decode("bad manifest") }).To(Panic())
	})

	DescribeTable("New",
		func(got, want runtime.Object) { Expect(got).To(EqualDiff(want)) },
		Entry("NewNamespace", runtime.NewNamespace("foo"), &corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{Name: "foo"},
			TypeMeta:   metav1.TypeMeta{Kind: "Namespace", APIVersion: "v1"},
		}),
		Entry("NewConfigMap", runtime.NewConfigMap("ns", "foo", nil), &corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{Name: "foo", Namespace: "ns"},
			TypeMeta:   metav1.TypeMeta{Kind: "ConfigMap", APIVersion: "v1"},
			Data:       map[string]string{},
		}),
	)
})
