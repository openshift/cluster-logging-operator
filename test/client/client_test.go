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

package client_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/openshift/cluster-logging-operator/test/client"
	. "github.com/openshift/cluster-logging-operator/test/matchers"
	"github.com/openshift/cluster-logging-operator/test/runtime"
	corev1 "k8s.io/api/core/v1"
)

var _ = Describe("Client", func() {
	var (
		t    *Test
		data map[string]string
	)

	BeforeEach(func() {
		t = NewTest()
		data = map[string]string{"a": "b"}
	})

	AfterEach(func() { t.Close() })

	It("creates object with data and automatic labels", func() {
		cm := runtime.NewConfigMap(t.NS.Name, "foo", data)
		ExpectOK(t.Create(cm))
		cm2 := runtime.NewConfigMap(t.NS.Name, "foo", nil)
		ExpectOK(t.Get(cm2))
		Expect(cm2.Data).To(Equal(data))
		Expect(cm2.Labels).To(HaveKeyWithValue(LabelKey, LabelValue))
	})

	It("re-creates existing object", func() {
		cm := runtime.NewConfigMap(t.NS.Name, "foo", map[string]string{"a": "b"})
		ExpectOK(t.Create(cm))
		cm2 := runtime.NewConfigMap(t.NS.Name, "foo", map[string]string{"x": "y"})
		ExpectOK(t.Recreate(cm2))
		cm3 := runtime.NewConfigMap(t.NS.Name, "foo", nil)
		ExpectOK(t.Get(cm3))
		Expect(cm3.Data).To(Equal(map[string]string{"x": "y"}))
		ExpectOK(t.Recreate(t.NS))
	})

	It("creates non-existing object", func() {
		cm := runtime.NewConfigMap(t.NS.Name, "foo", data)
		ExpectOK(t.Recreate(cm))
		cm2 := runtime.NewConfigMap(t.NS.Name, "foo", nil)
		ExpectOK(t.Get(cm2))
		Expect(cm2.Data).To(Equal(data))
	})

	It("lists objects", func() {
		l := &corev1.NodeList{}
		ExpectOK(t.List(l))
		Expect(l.Items).NotTo(BeEmpty())
	})
})
