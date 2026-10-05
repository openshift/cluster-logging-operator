package filters

import (
	. "github.com/onsi/ginkgo/v2"
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

	Context("#validateOlderThan", func() {
		DescribeTable("accepts dates and RFC3339 timestamps with an explicit offset", func(value string) {
			Expect(validateOlderThan(value)).To(BeEmpty())
		},
			Entry("date", "2026-09-16"),
			Entry("leap day", "2024-02-29"),
			Entry("UTC timestamp", "2026-09-16T00:15:30Z"),
			Entry("negative offset", "2026-09-16T00:15:30-04:00"),
			Entry("positive offset", "2026-09-16T12:15:30+05:30"),
			Entry("maximum offset components", "2026-09-16T23:59:00+23:59"),
			Entry("fractional seconds", "2026-09-16T00:15:30.123456789Z"),
		)

		DescribeTable("rejects invalid dates and timestamp formats", func(value string) {
			Expect(validateOlderThan(value)).To(ContainSubstring("invalid olderThan"))
		},
			Entry("missing offset", "2026-09-16T00:15:30"),
			Entry("comma fractional separator", "2026-09-16T00:15:30,1Z"),
			Entry("offset hour 24", "2026-09-16T00:15:30+24:00"),
			Entry("negative offset hour 24", "2026-09-16T00:15:30-24:00"),
			Entry("offset minute 60", "2026-09-16T00:15:30+00:60"),
			Entry("negative offset minute 60", "2026-09-16T00:15:30-00:60"),
			Entry("wrong date separator", "2026/09/16"),
			Entry("impossible date", "2026-02-30"),
			Entry("impossible timestamp date", "2026-02-30T00:15:30Z"),
			Entry("single-digit month", "2026-9-16"),
			Entry("single-digit day", "2026-09-6"),
			Entry("single-digit hour", "2026-09-16T0:15:30Z"),
		)
	})

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
			Entry("should fail validation if notMatches contains an invalid regular expression",
				[]obs.DropTest{
					{
						DropConditions: []obs.DropCondition{
							{
								Field:      ".kubernetes.namespace_name",
								NotMatches: "[invalid",
							},
						},
					},
				},
				"[matches/notMatches must be a valid regular expression.]",
			),
			Entry("should fail validation if matches contains a single quote",
				[]obs.DropTest{
					{
						DropConditions: []obs.DropCondition{
							{
								Field:   ".kubernetes.namespace_name",
								Matches: "foo'bar",
							},
						},
					},
				},
				"[matches/notMatches must not contain single quotes, newlines, or carriage returns]",
			),
			Entry("should fail validation if notMatches contains a single quote",
				[]obs.DropTest{
					{
						DropConditions: []obs.DropCondition{
							{
								Field:      ".kubernetes.namespace_name",
								NotMatches: "x'''[sources.evil]",
							},
						},
					},
				},
				"[matches/notMatches must not contain single quotes, newlines, or carriage returns]",
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

		DescribeTable("valid olderThan drop condition", func(olderThan string) {
			spec := obs.FilterSpec{
				Name: myDrop,
				Type: obs.FilterTypeDrop,
				DropTestsSpec: []obs.DropTest{{
					DropConditions: []obs.DropCondition{{OlderThan: olderThan}},
				}},
			}
			Expect(ValidateFilter(spec)).To(MatchCondition(expConditionTypeRE, true, obs.ReasonValidationSuccess, `filter.*is valid`))
		},
			Entry("accepts a date-only value", "2026-09-16"),
			Entry("accepts an RFC3339 timestamp with an offset", "2026-09-16T00:15:30-04:00"),
		)

		DescribeTable("invalid drop condition structure", func(dropTests []obs.DropTest, errMsg string) {
			spec := obs.FilterSpec{
				Name:          myDrop,
				Type:          obs.FilterTypeDrop,
				DropTestsSpec: dropTests,
			}
			Expect(ValidateFilter(spec)).To(MatchCondition(expConditionTypeRE, false, obs.ReasonValidationFailure, errMsg))
		},
			Entry("rejects an empty test list", []obs.DropTest{}, "at least one"),
			Entry("rejects a test with no conditions", []obs.DropTest{{}}, "at least one condition"),
			Entry("rejects an empty condition for its invalid field path", []obs.DropTest{{DropConditions: []obs.DropCondition{{}}}}, "must start with a '.'"),
			Entry("rejects a matcher without a field", []obs.DropTest{{DropConditions: []obs.DropCondition{{Matches: "message"}}}}, "must start with a '.'"),
		)

		DescribeTable("invalid olderThan drop condition", func(olderThan, errMsg string) {
			spec := obs.FilterSpec{
				Name: myDrop,
				Type: obs.FilterTypeDrop,
				DropTestsSpec: []obs.DropTest{{
					DropConditions: []obs.DropCondition{{OlderThan: olderThan}},
				}},
			}
			Expect(ValidateFilter(spec)).To(MatchCondition(expConditionTypeRE, false, obs.ReasonValidationFailure, errMsg))
		},
			Entry("rejects a timestamp without an offset", "2026-09-16T00:15:30", "invalid olderThan"),
			Entry("rejects a comma fractional separator", "2026-09-16T00:15:30,1Z", "invalid olderThan"),
			Entry("rejects offset hour 24", "2026-09-16T00:15:30+24:00", "invalid olderThan"),
			Entry("rejects offset minute 60", "2026-09-16T00:15:30+00:60", "invalid olderThan"),
			Entry("rejects a date with the wrong separator", "2026/09/16", "invalid olderThan"),
			Entry("rejects an impossible date", "2026-02-30", "invalid olderThan"),
		)

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
