package clusterlogforwarder

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	v1 "github.com/openshift/cluster-logging-operator/api/logging/v1"
	. "github.com/openshift/cluster-logging-operator/test/matchers"
)

var _ = Describe("[internal][validations] ClusterLogForwarder: Filters", func() {
	const (
		myDrop  = "dropFilter"
		myPrune = "pruneFilter"
	)

	Context("#validateDropFilters", func() {
		DescribeTable("invalid fields and matches/notMatches", func(dropTests []v1.DropTest, errMsg string) {
			clf := &v1.ClusterLogForwarder{
				Spec: v1.ClusterLogForwarderSpec{
					Filters: []v1.FilterSpec{
						{
							Name: myDrop,
							Type: v1.FilterDrop,
							FilterTypeSpec: v1.FilterTypeSpec{
								DropTestsSpec: &dropTests,
							},
						},
					},
				},
			}
			err, status := ValidateFilters(*clf, nil, nil)
			Expect(err).ToNot(BeNil())
			dropStatus := myDrop + ": test[0]"
			Expect(status.Filters[dropStatus]).To(HaveCondition(v1.ConditionReady, false, v1.ReasonInvalid, errMsg))
		},
			Entry("should fail validation if field does not start with a `.`",
				[]v1.DropTest{
					{
						DropConditions: []v1.DropCondition{
							{
								Field:   "kubernetes.namespace_name",
								Matches: "fooName",
							},
						},
					},
					{
						DropConditions: []v1.DropCondition{
							{
								Field:   "log_type",
								Matches: "match",
							},
						},
					},
				},
				"[field must start with a '.']"),
			Entry("should fail validation if field is not a valid path expression",
				[]v1.DropTest{
					{
						DropConditions: []v1.DropCondition{
							{
								Field:   ".kubernetes.foo-bar/baz",
								Matches: "busybox",
							},
						},
					},
					{
						DropConditions: []v1.DropCondition{
							{
								Field:   ".kubernetes.foo-bar",
								Matches: "busybox",
							},
						},
					},
				},
				`[field must be a valid dot delimited path expression (.kubernetes.container_name or .kubernetes.\"test\-foo\")]`,
			),
			Entry("should fail validation if both matches and notMatches are spec'd for one condition",
				[]v1.DropTest{
					{
						DropConditions: []v1.DropCondition{
							{
								Field:      ".kubernetes.test",
								Matches:    "foobar",
								NotMatches: "baz",
							},
						},
					},
				},
				"[only one of matches or notMatches can be defined at once]",
			),
			Entry("should fail validation if any matches or notMatches contain invalid regular expressions",
				[]v1.DropTest{
					{
						DropConditions: []v1.DropCondition{
							{
								Field:   ".kubernetes.namespace_name",
								Matches: "[",
							},
							{
								Field:      ".level",
								NotMatches: "debug",
							},
						},
					},
				},
				"[matches/notMatches must be a valid regular expression.]",
			),
		)

		DescribeTable("valid drop filter spec", func(dropTests []v1.DropTest) {
			clf := &v1.ClusterLogForwarder{
				Spec: v1.ClusterLogForwarderSpec{
					Filters: []v1.FilterSpec{
						{
							Name: myDrop,
							Type: v1.FilterDrop,
							FilterTypeSpec: v1.FilterTypeSpec{
								DropTestsSpec: &dropTests,
							},
						},
					},
				},
			}
			Expect(ValidateFilters(*clf, nil, nil)).To(Succeed())
		},
			Entry("should pass validation if fields start with a '.'",
				[]v1.DropTest{
					{
						DropConditions: []v1.DropCondition{
							{
								Field:   `.log_type`,
								Matches: "test",
							},
							{
								Field:      `.kubernetes.namespace_name`,
								NotMatches: "fooNamespace",
							},
						},
					},
				},
			),
			Entry("should pass validation if fields are valid path expressions",
				[]v1.DropTest{
					{
						DropConditions: []v1.DropCondition{
							{
								Field:   `.kubernetes."foo-bar/baz"`,
								Matches: "busybox",
							},
							{
								Field:      `.kubernetes.namespace_name`,
								NotMatches: "fooNamespace",
							},
						},
					},
				},
			),
			Entry("should pass validation when fields are valid path expressions and matches/notMatches are valid regular expressions",
				[]v1.DropTest{
					{
						DropConditions: []v1.DropCondition{
							{
								Field:   ".kubernetes.namespace_name",
								Matches: "busybox",
							},
							{
								Field:      ".level",
								NotMatches: "d.+",
							},
						},
					},
					{
						DropConditions: []v1.DropCondition{
							{
								Field:   ".log_type",
								Matches: "application",
							},
						},
					},
				},
			),
		)

		It("should fail if no drop conditions spec'd", func() {
			clf := &v1.ClusterLogForwarder{
				Spec: v1.ClusterLogForwarderSpec{
					Filters: []v1.FilterSpec{
						{
							Name: myDrop,
							Type: v1.FilterDrop,
						},
					},
				},
			}
			err, status := ValidateFilters(*clf, nil, nil)
			Expect(err).ToNot(BeNil())
			Expect(status.Filters[myDrop]).To(HaveCondition(v1.ConditionReady, false, v1.ReasonInvalid, "drop filter must have at least one test spec'd"))
		})
	})

	Context("#validatePruneFilter", func() {
		var requiredFields = []string{".log_type", ".message"}
		DescribeTable("invalid field paths", func(pruneFilter v1.PruneFilterSpec, errMsg string) {
			clf := &v1.ClusterLogForwarder{
				Spec: v1.ClusterLogForwarderSpec{
					Filters: []v1.FilterSpec{
						{
							Name: myPrune,
							Type: v1.FilterPrune,
							FilterTypeSpec: v1.FilterTypeSpec{
								PruneFilterSpec: &pruneFilter,
							},
						},
					},
				},
			}
			err, status := ValidateFilters(*clf, nil, nil)
			Expect(err).ToNot(BeNil())
			Expect(status.Filters[myPrune]).To(HaveCondition(v1.ConditionReady, false, v1.ReasonInvalid, errMsg))
		},
			Entry("should fail validation if fields in `in` do not start with a '.'",
				v1.PruneFilterSpec{
					In: []string{"foo.bar", `-foo."bar-baz/other@"`},
				},
				"[field must start with a '.']",
			),
			Entry("should fail validation if fields in `Notin` do not start with a '.'",
				v1.PruneFilterSpec{
					NotIn: append(requiredFields, "foo.bar", `-foo."bar-baz/other@"`),
				},
				"[field must start with a '.']",
			),
			Entry("should fail validation if fields in `in` are not valid path expressions",
				v1.PruneFilterSpec{
					In: []string{".foo.bar-", `.foo.bar-baz/other@`, ".@timestamp"},
				},
				"[field must be a valid dot delimited path expression+]",
			),
			Entry("should fail validation if fields in `notIn` are not valid path expressions",
				v1.PruneFilterSpec{
					NotIn: append(requiredFields, ".foo.bar", `.foo.bar-baz/other@`, ".@timestamp"),
				},
				"[field must be a valid dot delimited path expression+]",
			),
		)

		DescribeTable("valid field paths", func(pruneFilter v1.PruneFilterSpec) {
			clf := &v1.ClusterLogForwarder{
				Spec: v1.ClusterLogForwarderSpec{
					Filters: []v1.FilterSpec{
						{
							Name: myPrune,
							Type: v1.FilterPrune,
							FilterTypeSpec: v1.FilterTypeSpec{
								PruneFilterSpec: &pruneFilter,
							},
						},
					},
				},
			}
			Expect(ValidateFilters(*clf, nil, nil)).To(Succeed())
		},
			Entry("should pass validation if fields in `in` start with a '.'",
				v1.PruneFilterSpec{
					In: []string{".foo"},
				},
			),
			Entry("should pass validation if fields in `notIn` start with a '.'",
				v1.PruneFilterSpec{
					NotIn: append(requiredFields, ".foo"),
				},
			),
			Entry("should pass validation if fields in `notIn` are valid path expressions",
				v1.PruneFilterSpec{
					NotIn: append(requiredFields, ".foo.bar", `.foo."bar-baz/test"`, `."@timestamp"`),
				},
			),
			Entry("should pass validation if fields in `in` are valid path expressions",
				v1.PruneFilterSpec{
					In: []string{".foo.bar", `.foo."bar-baz/test"`, `."@timestamp"`},
				},
			),
			Entry("should pass validation if fields in both `in` & `notIn` are valid path expressions",
				v1.PruneFilterSpec{
					NotIn: append(requiredFields, ".foo.bar", `.foo."bar-baz/test"`, `."@timestamp"`),
					In:    []string{".foo", `.foo."bar-baz/test"`, `."@timestamp"`, ".foo_bar.testing.valid"},
				},
			),
		)

		Context("required fields", func() {
			It("should pass validation if required fields are not in the `in` list", func() {
				clf := &v1.ClusterLogForwarder{
					Spec: v1.ClusterLogForwarderSpec{
						Filters: []v1.FilterSpec{
							{
								Name: myPrune,
								Type: v1.FilterPrune,
								FilterTypeSpec: v1.FilterTypeSpec{
									PruneFilterSpec: &v1.PruneFilterSpec{
										In: []string{".foo", ".bar", ".foo.bar.baz"},
									},
								},
							},
						},
					},
				}
				Expect(ValidateFilters(*clf, nil, nil)).To(Succeed())
			})

			It("should fail validation if required fields are in the `in` list", func() {
				clf := &v1.ClusterLogForwarder{
					Spec: v1.ClusterLogForwarderSpec{
						Filters: []v1.FilterSpec{
							{
								Name: myPrune,
								Type: v1.FilterPrune,
								FilterTypeSpec: v1.FilterTypeSpec{
									PruneFilterSpec: &v1.PruneFilterSpec{
										In: append(requiredFields, ".foo", ".bar", ".foo.bar.baz"),
									},
								},
							},
						},
					},
				}
				err, status := ValidateFilters(*clf, nil, nil)
				Expect(err).ToNot(BeNil())
				Expect(status.Filters[myPrune]).To(HaveCondition(v1.ConditionReady, false, v1.ReasonInvalid, ".+is/are required fields and must be removed.+"))
			})

			It("should fail validation if 1 required field is in the `in` list", func() {
				clf := &v1.ClusterLogForwarder{
					Spec: v1.ClusterLogForwarderSpec{
						Filters: []v1.FilterSpec{
							{
								Name: myPrune,
								Type: v1.FilterPrune,
								FilterTypeSpec: v1.FilterTypeSpec{
									PruneFilterSpec: &v1.PruneFilterSpec{
										In: []string{".foo", ".bar", ".foo.bar.baz", requiredFields[0]},
									},
								},
							},
						},
					},
				}
				err, status := ValidateFilters(*clf, nil, nil)
				Expect(err).ToNot(BeNil())
				Expect(status.Filters[myPrune]).To(HaveCondition(v1.ConditionReady, false, v1.ReasonInvalid, ".+is/are required fields and must be removed.+"))
			})

			It("should pass validation if required fields are in the `notIn` list", func() {
				clf := &v1.ClusterLogForwarder{
					Spec: v1.ClusterLogForwarderSpec{
						Filters: []v1.FilterSpec{
							{
								Name: myPrune,
								Type: v1.FilterPrune,
								FilterTypeSpec: v1.FilterTypeSpec{
									PruneFilterSpec: &v1.PruneFilterSpec{
										NotIn: append(requiredFields, ".foo", ".bar", ".foo.bar.baz"),
									},
								},
							},
						},
					},
				}
				Expect(ValidateFilters(*clf, nil, nil)).To(Succeed())
			})

			It("should fail validation if required fields are not in the `notIn` list", func() {
				clf := &v1.ClusterLogForwarder{
					Spec: v1.ClusterLogForwarderSpec{
						Filters: []v1.FilterSpec{
							{
								Name: myPrune,
								Type: v1.FilterPrune,
								FilterTypeSpec: v1.FilterTypeSpec{
									PruneFilterSpec: &v1.PruneFilterSpec{
										NotIn: []string{".foo", ".bar", ".foo.bar.baz"},
									},
								},
							},
						},
					},
				}
				err, status := ValidateFilters(*clf, nil, nil)
				Expect(err).ToNot(BeNil())
				Expect(status.Filters[myPrune]).To(HaveCondition(v1.ConditionReady, false, v1.ReasonInvalid, ".+is/are required fields and must be included.+"))
			})

			It("should fail validation if 1 required field is not in the `notIn` list", func() {
				clf := &v1.ClusterLogForwarder{
					Spec: v1.ClusterLogForwarderSpec{
						Filters: []v1.FilterSpec{
							{
								Name: myPrune,
								Type: v1.FilterPrune,
								FilterTypeSpec: v1.FilterTypeSpec{
									PruneFilterSpec: &v1.PruneFilterSpec{
										NotIn: []string{".foo", ".bar", ".foo.bar.baz", requiredFields[0]},
									},
								},
							},
						},
					},
				}
				err, status := ValidateFilters(*clf, nil, nil)
				Expect(err).ToNot(BeNil())
				Expect(status.Filters[myPrune]).To(HaveCondition(v1.ConditionReady, false, v1.ReasonInvalid, ".+is/are required fields and must be included.+"))
			})

			It("should fail validation if required fields are in the `notIn` list and in the `in` list", func() {
				clf := &v1.ClusterLogForwarder{
					Spec: v1.ClusterLogForwarderSpec{
						Filters: []v1.FilterSpec{
							{
								Name: myPrune,
								Type: v1.FilterPrune,
								FilterTypeSpec: v1.FilterTypeSpec{
									PruneFilterSpec: &v1.PruneFilterSpec{
										In:    append(requiredFields, ".foo", ".bar", ".foo.bar.baz"),
										NotIn: append(requiredFields, ".foo", ".bar", ".foo.bar.baz"),
									},
								},
							},
						},
					},
				}
				err, status := ValidateFilters(*clf, nil, nil)
				Expect(err).ToNot(BeNil())
				Expect(status.Filters[myPrune]).To(HaveCondition(v1.ConditionReady, false, v1.ReasonInvalid, ".+is/are required fields and must be removed.+"))
			})

			It("should fail validation if required fields are not in the `notIn` list and not in the `in` list", func() {
				clf := &v1.ClusterLogForwarder{
					Spec: v1.ClusterLogForwarderSpec{
						Filters: []v1.FilterSpec{
							{
								Name: myPrune,
								Type: v1.FilterPrune,
								FilterTypeSpec: v1.FilterTypeSpec{
									PruneFilterSpec: &v1.PruneFilterSpec{
										In:    []string{".foo", ".bar", ".foo.bar.baz"},
										NotIn: []string{".foo", ".bar", ".foo.bar.baz"},
									},
								},
							},
						},
					},
				}
				err, status := ValidateFilters(*clf, nil, nil)
				Expect(err).ToNot(BeNil())
				Expect(status.Filters[myPrune]).To(HaveCondition(v1.ConditionReady, false, v1.ReasonInvalid, ".+is/are required fields and must be included.+"))
			})

		})

		It("should fail validation if prune filter spec'd without pruneFilterSpec", func() {
			clf := &v1.ClusterLogForwarder{
				Spec: v1.ClusterLogForwarderSpec{
					Filters: []v1.FilterSpec{
						{
							Name: myPrune,
							Type: v1.FilterPrune,
						},
					},
				},
			}
			err, status := ValidateFilters(*clf, nil, nil)
			Expect(err).ToNot(BeNil())
			Expect(status.Filters[myPrune]).To(HaveCondition(v1.ConditionReady, false, v1.ReasonInvalid, "prune filter must have one or both of `in`, `notIn`"))
		})

	})
})
