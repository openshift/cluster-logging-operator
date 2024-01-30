package clusterlogforwarder

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	v1 "github.com/openshift/cluster-logging-operator/apis/logging/v1"
	. "github.com/openshift/cluster-logging-operator/test/matchers"
)

var _ = Describe("[internal][validations] ClusterLogForwarder: Filters", func() {
	const myDrop = "dropFilter"
	var clf = &v1.ClusterLogForwarder{
		Spec: v1.ClusterLogForwarderSpec{
			Filters: []v1.FilterSpec{
				{
					Name: myDrop,
					Type: v1.FilterDrop,
					FilterTypeSpec: v1.FilterTypeSpec{
						DropTestsSpec: &[]v1.DropTest{},
					},
				},
			},
		},
	}

	Context("#validateDropFilters", func() {
		It("should fail validation if field does not start with a `.`", func() {
			clf.Spec.Filters[0].DropTestsSpec = &[]v1.DropTest{
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
							Field:   ".log_type",
							Matches: "myType",
						},
					},
				},
			}
			err, status := ValidateFilters(*clf, nil, nil)
			Expect(err).ToNot(BeNil())
			Expect(status.Filters[myDrop]).To(HaveCondition(v1.ConditionReady, false, v1.ReasonInvalid, "[field must start with a '.']"))
		})

		It("should pass validation if fields start with a '.'", func() {
			clf.Spec.Filters[0].DropTestsSpec = &[]v1.DropTest{
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
			}
			Expect(ValidateFilters(*clf, nil, nil)).To(Succeed())
		})

		It("should fail validation if field is not a valid path expression", func() {
			clf.Spec.Filters[0].DropTestsSpec = &[]v1.DropTest{
				{
					DropConditions: []v1.DropCondition{
						{
							Field:   ".kubernetes.foo-bar/baz",
							Matches: "busybox",
						},
					},
				},
			}
			err, status := ValidateFilters(*clf, nil, nil)
			Expect(err).ToNot(BeNil())
			Expect(status.Filters[myDrop]).To(HaveCondition(v1.ConditionReady, false, v1.ReasonInvalid, `[field must be a valid dot delimited path expression (.kubernetes.container_name or .kubernetes.\"test\-foo\")]`))
		})

		It("should pass validation if fields are valid path expressions", func() {
			clf.Spec.Filters[0].DropTestsSpec = &[]v1.DropTest{
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
			}
			Expect(ValidateFilters(*clf, nil, nil)).To(Succeed())
		})

		It("should fail validation if both matches and notMatches are spec'd for one condition", func() {
			clf.Spec.Filters[0].DropTestsSpec = &[]v1.DropTest{
				{
					DropConditions: []v1.DropCondition{
						{
							Field:      ".kubernetes.test",
							Matches:    "foobar",
							NotMatches: "baz",
						},
					},
				},
			}
			err, status := ValidateFilters(*clf, nil, nil)
			Expect(err).ToNot(BeNil())
			Expect(status.Filters[myDrop]).To(HaveCondition(v1.ConditionReady, false, v1.ReasonInvalid, "[only one of matches or notMatches can be defined at once]"))
		})

		It("should fail validation if any matches or notMatches contain invalid regular expressions", func() {
			clf.Spec.Filters[0].DropTestsSpec = &[]v1.DropTest{
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
			}
			err, status := ValidateFilters(*clf, nil, nil)
			Expect(err).ToNot(BeNil())
			Expect(status.Filters[myDrop]).To(HaveCondition(v1.ConditionReady, false, v1.ReasonInvalid, "[matches/notMatches must be a valid regular expression.]"))
		})

		It("should pass validation when fields are valid path expressions and matches/notMatches are valid regular expressions", func() {
			clf.Spec.Filters[0].DropTestsSpec = &[]v1.DropTest{
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
			}
			Expect(ValidateFilters(*clf, nil, nil)).To(Succeed())
		})
	})
})
