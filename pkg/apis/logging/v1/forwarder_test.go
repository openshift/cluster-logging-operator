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

package v1_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/openshift/cluster-logging-operator/pkg/apis/logging/v1"
)

var _ = Describe("ClusterLogForwarderSpec", func() {

	It("calculates routes", func() {
		spec := ClusterLogForwarderSpec{
			Pipelines: []PipelineSpec{
				{
					InputRefs:  []string{InputNameApplication},
					OutputRefs: []string{"X", "Y"},
				},
				{
					InputRefs:  []string{InputNameInfrastructure, InputNameAudit},
					OutputRefs: []string{"Y", "Z"},
				},
				{
					InputRefs:  []string{InputNameAudit},
					OutputRefs: []string{"X", "Z"},
				},
			},
		}
		routes := NewRoutes(spec.Pipelines)
		Expect(routes.ByInput).To(Equal(RouteMap{
			InputNameAudit:          {"X": {}, "Y": {}, "Z": {}},
			InputNameApplication:    {"X": {}, "Y": {}},
			InputNameInfrastructure: {"Y": {}, "Z": {}},
		}))
		Expect(routes.ByOutput).To(Equal(RouteMap{
			"X": {InputNameApplication: {}, InputNameAudit: {}},
			"Y": {InputNameApplication: {}, InputNameInfrastructure: {}, InputNameAudit: {}},
			"Z": {InputNameInfrastructure: {}, InputNameAudit: {}},
		}))
	})
})

var _ = Describe("inputs", func() {
	It("has built-in input types", func() {
		Expect(ReservedInputNames.List()).To(ConsistOf("infrastructure", "application", "audit"))
	})
})
