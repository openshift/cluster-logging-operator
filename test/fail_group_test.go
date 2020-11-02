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
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/openshift/cluster-logging-operator/test"
)

var _ = Describe("FailGroup", func() {
	It("handles concurrent pass and fail", func() {
		g := test.FailGroup{}
		c := make(chan bool) // Blocking channel to ensure goroutines run concurrently
		fails := InterceptGomegaFailures(func() {
			g.Go(func() { <-c; Expect("bad").To(Equal("good")) })
			g.Go(func() { c <- true; Expect("hello").To(Equal("hello")) })
			g.Wait()
		})
		Expect(fails).To(ConsistOf("Expected\n    <string>: bad\nto equal\n    <string>: good"))
	})
})
