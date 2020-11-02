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

package indexmanagement

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	logging "github.com/openshift/cluster-logging-operator/pkg/apis/logging/v1"
	esapi "github.com/openshift/elasticsearch-operator/pkg/apis/logging/v1"
)

var _ = Describe("Indexmanagement", func() {

	var retentionPolicy *logging.RetentionPoliciesSpec
	var defaultPolicy *logging.RetentionPoliciesSpec

	BeforeEach(func() {
		retentionPolicy = &logging.RetentionPoliciesSpec{
			App: &logging.RetentionPolicySpec{
				MaxAge: esapi.TimeUnit("1h"),
			},
			Infra: &logging.RetentionPolicySpec{
				MaxAge: esapi.TimeUnit("2h"),
			},
			Audit: &logging.RetentionPolicySpec{
				MaxAge: esapi.TimeUnit("3h"),
			},
		}
	})

	Describe("IndexManagement Policy creation failure", func() {
		Context("when retention policy is not defined", func() {
			BeforeEach(func() {
				retentionPolicy = nil
				defaultPolicy = &logging.RetentionPoliciesSpec{
					App: &logging.RetentionPolicySpec{
						MaxAge: esapi.TimeUnit("7d"),
					},
					Infra: &logging.RetentionPolicySpec{
						MaxAge: esapi.TimeUnit("7d"),
					},
					Audit: &logging.RetentionPolicySpec{
						MaxAge: esapi.TimeUnit("7d"),
					},
				}
			})
			It("should generate default index management", func() {
				spec := NewSpec(retentionPolicy)
				Expect(len(spec.Policies)).To(Equal(3))
				Expect(len(spec.Mappings)).To(Equal(3))
				Expect(spec.Policies[0].Phases.Delete.MinAge).To(Equal(defaultPolicy.App.MaxAge))
				Expect(spec.Policies[1].Phases.Delete.MinAge).To(Equal(defaultPolicy.Infra.MaxAge))
				Expect(spec.Policies[2].Phases.Delete.MinAge).To(Equal(defaultPolicy.Audit.MaxAge))
			})
		})
		Context("retention policy App log source has low maxAge", func() {
			BeforeEach(func() {
				retentionPolicy.App.MaxAge = "10s"
			})
			It("should not generate index management", func() {
				spec := NewSpec(retentionPolicy)
				Expect(spec).To(BeNil())
			})
		})
		Context("retention policy Infra log source has low maxAge", func() {
			BeforeEach(func() {
				retentionPolicy.Infra.MaxAge = "10s"
			})
			It("should not generate index management", func() {
				spec := NewSpec(retentionPolicy)
				Expect(spec).To(BeNil())
			})
		})
		Context("retention policy Audit log source has low maxAge", func() {
			BeforeEach(func() {
				retentionPolicy.Audit.MaxAge = "10s"
			})
			It("should not generate index management", func() {
				spec := NewSpec(retentionPolicy)
				Expect(spec).To(BeNil())
			})
		})

	})
	Describe("IndexManagement Policy creation success", func() {
		Context("Policy and Mapping generated", func() {
			It("For All log source types", func() {
				spec := NewSpec(retentionPolicy)
				Expect(len(spec.Policies)).To(Equal(3))
				Expect(len(spec.Mappings)).To(Equal(3))
			})
		})
		Context("Hot Phase durations in created spec ", func() {
			It("Must conform to the regex", func() {
				spec := NewSpec(retentionPolicy)
				Expect(agePattern.Match([]byte(spec.Policies[0].Phases.Hot.Actions.Rollover.MaxAge))).To(Equal(true))
				Expect(agePattern.Match([]byte(spec.Policies[1].Phases.Hot.Actions.Rollover.MaxAge))).To(Equal(true))
				Expect(agePattern.Match([]byte(spec.Policies[2].Phases.Hot.Actions.Rollover.MaxAge))).To(Equal(true))
			})
		})
		Context("Delete Phase durations in created spec ", func() {
			It("Must conform to the regex", func() {
				spec := NewSpec(retentionPolicy)
				Expect(agePattern.Match([]byte(spec.Policies[0].Phases.Delete.MinAge))).To(Equal(true))
				Expect(agePattern.Match([]byte(spec.Policies[1].Phases.Delete.MinAge))).To(Equal(true))
				Expect(agePattern.Match([]byte(spec.Policies[2].Phases.Delete.MinAge))).To(Equal(true))
			})
		})
		Context("Delete Phase durations in created spec", func() {
			It("Must be same as set in retention policy", func() {
				spec := NewSpec(retentionPolicy)
				Expect(spec.Policies[0].Phases.Delete.MinAge).To(Equal(retentionPolicy.App.MaxAge))
				Expect(spec.Policies[1].Phases.Delete.MinAge).To(Equal(retentionPolicy.Infra.MaxAge))
				Expect(spec.Policies[2].Phases.Delete.MinAge).To(Equal(retentionPolicy.Audit.MaxAge))
			})
		})
		Context("Spec Mappings", func() {
			It("Policy-ref should be same as Policy Name", func() {
				spec := NewSpec(retentionPolicy)
				Expect(spec.Mappings[0].PolicyRef).To(Equal(spec.Policies[0].Name))
				Expect(spec.Mappings[1].PolicyRef).To(Equal(spec.Policies[1].Name))
				Expect(spec.Mappings[2].PolicyRef).To(Equal(spec.Policies[2].Name))
			})
		})
	})
	Describe("Index Management Policy Partial creation", func() {
		Context("Retention policy is defined only for App Log Source", func() {
			BeforeEach(func() {
				retentionPolicy.Infra = nil
				retentionPolicy.Audit = nil
			})
			It("should generate index management for App log source and default the others", func() {
				spec := NewSpec(retentionPolicy)
				Expect(len(spec.Policies)).To(Equal(3))
				Expect(spec.Policies[0].Name).To(Equal(PolicyNameApp))

				Expect(spec.Policies[1].Phases.Delete.MinAge).To(Equal(defaultPolicy.Infra.MaxAge))
				Expect(spec.Policies[2].Phases.Delete.MinAge).To(Equal(defaultPolicy.Audit.MaxAge))

				Expect(len(spec.Mappings)).To(Equal(3))
				Expect(spec.Mappings[0].PolicyRef).To(Equal(PolicyNameApp))
				Expect(spec.Mappings[1].PolicyRef).To(Equal(PolicyNameInfra))
				Expect(spec.Mappings[2].PolicyRef).To(Equal(PolicyNameAudit))
			})
		})
	})
	Describe("TimeUnit tests", func() {
		var (
			time int
			unit byte
			err  error
		)
		Context("converting to lower units", func() {
			It("year to days", func() {
				time, unit, err = convertToLowerUnits(1, 'y')
				Expect(time).To(Equal(365), "time is incorrect")
				Expect(unit).To(Equal(byte('d')), "unit is incorrect")
				Expect(err).To(BeNil(), "error must be nil")
			})
			It("month to days", func() {
				time, unit, err = convertToLowerUnits(1, 'M')
				Expect(time).To(Equal(30), "time is incorrect")
				Expect(unit).To(Equal(byte('d')), "unit is incorrect")
				Expect(err).To(BeNil(), "error must be nil")
			})
			It("week to days", func() {
				time, unit, err = convertToLowerUnits(1, 'w')
				Expect(time).To(Equal(7), "time is incorrect")
				Expect(unit).To(Equal(byte('d')), "unit is incorrect")
				Expect(err).To(BeNil(), "error must be nil")
			})
			It("day to hours", func() {
				time, unit, err = convertToLowerUnits(1, 'd')
				Expect(time).To(Equal(24), "time is incorrect")
				Expect(unit).To(Equal(byte('h')), "unit is incorrect")
				Expect(err).To(BeNil(), "error must be nil")
			})
			It("hour to minutes", func() {
				time, unit, err = convertToLowerUnits(1, 'h')
				Expect(time).To(Equal(60), "time is incorrect")
				Expect(unit).To(Equal(byte('m')), "unit is incorrect")
				Expect(err).To(BeNil(), "error must be nil")
			})
			It("minutes to seconds", func() {
				time, unit, err = convertToLowerUnits(1, 'm')
				Expect(time).To(Equal(60), "time is incorrect")
				Expect(unit).To(Equal(byte('s')), "unit is incorrect")
				Expect(err).To(BeNil(), "error must be nil")
			})
			It("days to seconds", func() {
				time, unit, err = convertToLowerUnits(5, 'd')
				time, unit, err = convertToLowerUnits(time, unit)
				time, unit, err = convertToLowerUnits(time, unit)
				Expect(time).To(Equal(5*24*60*60), "time is incorrect")
				Expect(unit).To(Equal(byte('s')), "unit is incorrect")
				Expect(err).To(BeNil(), "error must be nil")
			})
		})
	})
})
