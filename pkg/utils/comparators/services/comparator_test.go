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

package services_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	. "github.com/openshift/cluster-logging-operator/pkg/utils/comparators/services"
)

var _ = Describe("services#AreSame", func() {

	var (
		current, desired *v1.Service
	)

	BeforeEach(func() {
		current = &v1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Labels: map[string]string{},
			},
			Spec: v1.ServiceSpec{
				Selector: map[string]string{},
				Ports:    []v1.ServicePort{},
			},
		}
		desired = current.DeepCopy()

	})

	Context("when evaluating labels", func() {
		It("should recognize they are the same", func() {
			Expect(AreSame(current, desired)).To(BeTrue())
		})

		It("should recognize they are different", func() {
			desired.Labels["foo"] = "bar"
			Expect(AreSame(current, desired)).To(BeFalse())
		})
	})
	Context("when evaluating selectors", func() {
		It("should recognize they are the same", func() {
			Expect(AreSame(current, desired)).To(BeTrue())
		})

		It("should recognize they are different", func() {
			desired.Spec.Selector["foo"] = "bar"
			Expect(AreSame(current, desired)).To(BeFalse())
		})
	})
	Context("when evaluating ServicePorts", func() {
		It("should recognize they are the same", func() {
			Expect(AreSame(current, desired)).To(BeTrue())
		})

		It("should recognize they are different lengths", func() {
			desired.Spec.Ports = append(desired.Spec.Ports, v1.ServicePort{})
			Expect(AreSame(current, desired)).To(BeFalse())
		})

		It("should recognize they are different content", func() {
			current.Spec.Ports = append(desired.Spec.Ports, v1.ServicePort{Name: "bar", Port: 1051})
			desired.Spec.Ports = append(desired.Spec.Ports, v1.ServicePort{Name: "bar", Port: 1050})
			Expect(AreSame(current, desired)).To(BeFalse())
		})
	})
})
