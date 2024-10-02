package filters

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	. "github.com/openshift/cluster-logging-operator/test/matchers"
)

var _ = Describe("[internal][validations][observability][filters]", func() {
	const (
		myDrop             = "dropFilter"
		myPrune            = "pruneFilter"
		expConditionTypeRE = obs.ConditionTypeValidFilterPrefix + "-.*"
	)

	Context("#validateDropFilters", func() {
		DescribeTable("invalid fields and matches/notMatches", func(dropTests []obs.DropTest, errMsg string) {
			spec := obs.FilterSpec{
				Name:          myDrop,
				Type:          obs.FilterTypeDrop,
				DropTestsSpec: dropTests,
			}
			Expect(ValidateFilter(spec)).To(MatchCondition(expConditionTypeRE, false, obs.ReasonValidationFailure, errMsg))
		},
			Entry("should fail validation if field does not start with a `.`",
				[]obs.DropTest{
					{
						DropConditions: []obs.DropCondition{
							{
								Field:   "kubernetes.namespace_name",
								Matches: "fooName",
							},
						},
					},
					{
						DropConditions: []obs.DropCondition{
							{
								Field:   "log_type",
								Matches: "match",
							},
						},
					},
				},
				"[field must start with a '.']"),
			Entry("should fail validation if field is not a valid path expression",
				[]obs.DropTest{
					{
						DropConditions: []obs.DropCondition{
							{
								Field:   ".kubernetes.foo-bar/baz",
								Matches: "busybox",
							},
						},
					},
					{
						DropConditions: []obs.DropCondition{
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
				[]obs.DropTest{
					{
						DropConditions: []obs.DropCondition{
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
				[]obs.DropTest{
					{
						DropConditions: []obs.DropCondition{
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

		DescribeTable("valid drop filter spec", func(dropTests []obs.DropTest) {
			spec := obs.FilterSpec{
				Name:          myDrop,
				Type:          obs.FilterTypeDrop,
				DropTestsSpec: dropTests,
			}
			Expect(ValidateFilter(spec)).To(MatchCondition(expConditionTypeRE, true, obs.ReasonValidationSuccess, `filter.*is valid`))
		},
			Entry("should pass validation if fields start with a '.'",
				[]obs.DropTest{
					{
						DropConditions: []obs.DropCondition{
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
				[]obs.DropTest{
					{
						DropConditions: []obs.DropCondition{
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
				[]obs.DropTest{
					{
						DropConditions: []obs.DropCondition{
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
						DropConditions: []obs.DropCondition{
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
			spec := obs.FilterSpec{
				Name: myDrop,
				Type: obs.FilterTypeDrop,
			}
			Expect(ValidateFilter(spec)).To(MatchCondition(expConditionTypeRE, false, obs.ReasonValidationFailure, "at least one"))
		})
	})

	Context("#validatePruneFilter", func() {
		var requiredFields = []obs.FieldPath{".log_type", ".message", ".log_source"}
		DescribeTable("invalid field paths", func(pruneFilter obs.PruneFilterSpec, errMsg string) {
			spec := obs.FilterSpec{
				Name:            myPrune,
				Type:            obs.FilterTypePrune,
				PruneFilterSpec: &pruneFilter,
			}
			Expect(ValidateFilter(spec)).To(MatchCondition(expConditionTypeRE, false, obs.ReasonValidationFailure, errMsg))
		},
			Entry("should fail validation if fields in `in` do not start with a '.'",
				obs.PruneFilterSpec{
					In: []obs.FieldPath{"foo.bar", `-foo."bar-baz/other@"`},
				},
				"[field must start with a '.']",
			),
			Entry("should fail validation if fields in `Notin` do not start with a '.'",
				obs.PruneFilterSpec{
					NotIn: append(requiredFields, "foo.bar", `-foo."bar-baz/other@"`),
				},
				"[field must start with a '.']",
			),
			Entry("should fail validation if fields in `in` are not valid path expressions",
				obs.PruneFilterSpec{
					In: []obs.FieldPath{".foo.bar-", `.foo.bar-baz/other@`, ".@timestamp"},
				},
				"[field must be a valid dot delimited path expression+]",
			),
			Entry("should fail validation if fields in `notIn` are not valid path expressions",
				obs.PruneFilterSpec{
					NotIn: append(requiredFields, ".foo.bar", `.foo.bar-baz/other@`, ".@timestamp"),
				},
				"[field must be a valid dot delimited path expression+]",
			),
		)

		DescribeTable("valid field paths", func(pruneFilter obs.PruneFilterSpec) {
			spec := obs.FilterSpec{
				Name:            myPrune,
				Type:            obs.FilterTypePrune,
				PruneFilterSpec: &pruneFilter,
			}
			Expect(ValidateFilter(spec)).To(MatchCondition(expConditionTypeRE, true, obs.ReasonValidationSuccess, `filter.*is valid`))
		},
			Entry("should pass validation if fields in `in` start with a '.'",
				obs.PruneFilterSpec{
					In: []obs.FieldPath{".foo"},
				},
			),
			Entry("should pass validation if fields in `notIn` start with a '.'",
				obs.PruneFilterSpec{
					NotIn: append(requiredFields, ".foo"),
				},
			),
			Entry("should pass validation if fields in `notIn` are valid path expressions",
				obs.PruneFilterSpec{
					NotIn: append(requiredFields, ".foo.bar", `.foo."bar-baz/test"`, `."@timestamp"`),
				},
			),
			Entry("should pass validation if fields in `in` are valid path expressions",
				obs.PruneFilterSpec{
					In: []obs.FieldPath{".foo.bar", `.foo."bar-baz/test"`, `."@timestamp"`},
				},
			),
			Entry("should pass validation if fields in both `in` & `notIn` are valid path expressions",
				obs.PruneFilterSpec{
					NotIn: append(requiredFields, ".foo.bar", `.foo."bar-baz/test"`, `."@timestamp"`),
					In:    []obs.FieldPath{".foo", `.foo."bar-baz/test"`, `."@timestamp"`, ".foo_bar.testing.valid"},
				},
			),
		)

		Context("required fields", func() {
			It("should pass validation if required fields are not in the `in` list", func() {
				spec := obs.FilterSpec{
					Name: myPrune,
					Type: obs.FilterTypePrune,
					PruneFilterSpec: &obs.PruneFilterSpec{
						In: []obs.FieldPath{".foo", ".bar", ".foo.bar.baz"},
					},
				}
				Expect(ValidateFilter(spec)).To(MatchCondition(expConditionTypeRE, true, obs.ReasonValidationSuccess, `filter.*is valid`))
			})

			It("should fail validation if required fields are in the `in` list", func() {
				spec := obs.FilterSpec{
					Name: myPrune,
					Type: obs.FilterTypePrune,
					PruneFilterSpec: &obs.PruneFilterSpec{
						In: append(requiredFields, ".foo", ".bar", ".foo.bar.baz"),
					},
				}
				Expect(ValidateFilter(spec)).To(MatchCondition(expConditionTypeRE, false, obs.ReasonValidationFailure, ".+is/are required fields and must be removed.+"))
			})

			It("should fail validation if 1 required field is in the `in` list", func() {
				spec := obs.FilterSpec{
					Name: myPrune,
					Type: obs.FilterTypePrune,
					PruneFilterSpec: &obs.PruneFilterSpec{
						In: []obs.FieldPath{".foo", ".bar", ".foo.bar.baz", requiredFields[0]},
					},
				}
				Expect(ValidateFilter(spec)).To(MatchCondition(expConditionTypeRE, false, obs.ReasonValidationFailure, ".+is/are required fields and must be removed.+"))
			})

			It("should pass validation if required fields are in the `notIn` list", func() {
				spec := obs.FilterSpec{
					Name: myPrune,
					Type: obs.FilterTypePrune,
					PruneFilterSpec: &obs.PruneFilterSpec{
						NotIn: append(requiredFields, ".foo", ".bar", ".foo.bar.baz"),
					},
				}
				Expect(ValidateFilter(spec)).To(MatchCondition(expConditionTypeRE, true, obs.ReasonValidationSuccess, `filter.*is valid`))
			})

			It("should fail validation if required fields are not in the `notIn` list", func() {
				spec := obs.FilterSpec{
					Name: myPrune,
					Type: obs.FilterTypePrune,
					PruneFilterSpec: &obs.PruneFilterSpec{
						NotIn: []obs.FieldPath{".foo", ".bar", ".foo.bar.baz"},
					},
				}
				Expect(ValidateFilter(spec)).To(MatchCondition(expConditionTypeRE, false, obs.ReasonValidationFailure, ".+is/are required fields and must be included.+"))
			})

			It("should fail validation if 1 required field is not in the `notIn` list", func() {
				spec := obs.FilterSpec{
					Name: myPrune,
					Type: obs.FilterTypePrune,
					PruneFilterSpec: &obs.PruneFilterSpec{
						NotIn: []obs.FieldPath{".foo", ".bar", ".foo.bar.baz", requiredFields[0]},
					},
				}
				Expect(ValidateFilter(spec)).To(MatchCondition(expConditionTypeRE, false, obs.ReasonValidationFailure, ".+is/are required fields and must be included.+"))
			})

			It("should fail validation if required fields are in the `notIn` list and in the `in` list", func() {
				spec := obs.FilterSpec{
					Name: myPrune,
					Type: obs.FilterTypePrune,
					PruneFilterSpec: &obs.PruneFilterSpec{
						In:    append(requiredFields, ".foo", ".bar", ".foo.bar.baz"),
						NotIn: append(requiredFields, ".foo", ".bar", ".foo.bar.baz"),
					},
				}
				Expect(ValidateFilter(spec)).To(MatchCondition(expConditionTypeRE, false, obs.ReasonValidationFailure, ".+is/are required fields and must be removed.+"))
			})

			It("should fail validation if required fields are not in the `notIn` list and not in the `in` list", func() {
				spec := obs.FilterSpec{
					Name: myPrune,
					Type: obs.FilterTypePrune,
					PruneFilterSpec: &obs.PruneFilterSpec{
						In:    []obs.FieldPath{".foo", ".bar", ".foo.bar.baz"},
						NotIn: []obs.FieldPath{".foo", ".bar", ".foo.bar.baz"},
					},
				}
				Expect(ValidateFilter(spec)).To(MatchCondition(expConditionTypeRE, false, obs.ReasonValidationFailure, ".+is/are required fields and must be included.+"))
			})

		})

		It("should fail validation if prune filter spec'd without pruneFilterSpec", func() {
			spec := obs.FilterSpec{
				Name: myPrune,
				Type: obs.FilterTypePrune,
			}
			Expect(ValidateFilter(spec)).To(MatchCondition(expConditionTypeRE, false, obs.ReasonValidationFailure, "prune filter must have one or both of `in`, `notIn`"))
		})

	})
})
