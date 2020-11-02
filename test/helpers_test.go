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

package test_test

import (
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/openshift/cluster-logging-operator/test"
	"k8s.io/apimachinery/pkg/util/validation"
)

const (
	// Match unique name suffix (-HHMMSSxxxxxxxx).
	suffix    = "-[0-9]{6}[0-9a-f]{8}"
	suffixLen = 1 + 6 + 8
)

var _ = Describe("Helpers", func() {
	Describe("Unmarshal", func() {
		It("unmarshals YAML", func() {
			var m map[string]string
			MustUnmarshal("a: b\nx: \"y\"\n", &m)
			Expect(m).To(Equal(map[string]string{"a": "b", "x": "y"}))
		})

		It("unmarshals JSON", func() {
			var m map[string]string
			MustUnmarshal(`{"a":"b", "x":"y"}`, &m)
			Expect(m).To(Equal(map[string]string{"a": "b", "x": "y"}))
		})
	})

	Describe("UniqueName", func() {
		It("generates unique names", func() {
			names := map[string]bool{}
			for i := 0; i < 100; i++ {
				name := UniqueName("x")
				Expect(name).To(MatchRegexp("x" + suffix))
				Expect(names).NotTo(HaveKey(name), "not unique")
			}
		})

		It("cleans up an illegal name", func() {
			name := UniqueName("x--y!-@#z--")
			Expect(validation.IsDNS1123Label(name)).To(BeNil(), name)
			Expect(name).To(MatchRegexp("x-y-z" + suffix))
		})

		It("truncates a long prefix", func() {
			name := UniqueName(strings.Repeat("ghijklmnop", 100))
			Expect(validation.IsDNS1035Label(name)).To(BeNil())
			Expect(name).To(MatchRegexp(name[:validation.DNS1123LabelMaxLength-suffixLen] + suffix))
		})
	})

	Describe("CurrentUniqueName", func() {
		It("uses test name", func() {
			Expect(UniqueNameForTest()).To(MatchRegexp("uses-test-name" + suffix))
		})
	})
})
