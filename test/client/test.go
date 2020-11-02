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

package client

import (
	"fmt"

	"github.com/onsi/ginkgo"
	"github.com/openshift/cluster-logging-operator/test"
	"github.com/openshift/cluster-logging-operator/test/runtime"
	corev1 "k8s.io/api/core/v1"
)

// Test wraps the singleton test client with setup/teardown and convenience methods
// for testing.
type Test struct {
	*Client
	NS *corev1.Namespace
}

// NewTest creates a new Test, which includes creating a new test namespace.
func NewTest() *Test {
	t := &Test{
		Client: Get(),
		NS:     runtime.NewUniqueNamespace(),
	}
	test.Must(t.Create(t.NS))
	fmt.Fprintf(ginkgo.GinkgoWriter, "test namespace: %v\n", t.NS.Name)
	return t
}

// Close removes the test namespace unless called from a failed test.
func (t *Test) Close() {
	if !ginkgo.CurrentGinkgoTestDescription().Failed {
		_ = t.Remove(t.NS)
	}
}
