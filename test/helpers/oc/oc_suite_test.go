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

package oc_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

const podSpec = `
apiVersion: v1
kind: Pod
metadata:
  name: log-generator
  labels:
    component: test
  namespace: test-log-gen
spec:
  containers:
    - name: log-generator
      image: docker.io/library/busybox:1.31.1
      args: ["sh", "-c", "i=0; while true; do echo $i: Test message; i=$((i+1)) ; sleep 1; done"]
`

func TestOC(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Test OC Commands")
}
