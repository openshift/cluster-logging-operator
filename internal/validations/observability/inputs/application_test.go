package inputs

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	. "github.com/openshift/cluster-logging-operator/test/matchers"
)

var _ = Describe("#ValidateApplication", func() {

	const (
		expConditionTypeRE = obs.ConditionValidInputPrefix + "-.*"
	)
	var (
		input obs.InputSpec
	)
	BeforeEach(func() {
		input = obs.InputSpec{
			Name:        "myapp",
			Type:        obs.InputTypeApplication,
			Application: &obs.Application{},
		}
	})
	It("should skip the validation when not an application type", func() {
		input.Type = obs.InputTypeInfrastructure
		conds := ValidateApplication(input)
		Expect(conds).To(BeEmpty())
	})
	It("should fail when an application type but has no application input", func() {
		input.Application = nil
		conds := ValidateApplication(input)
		Expect(conds).To(HaveCondition(expConditionTypeRE, false, obs.ReasonMissingSpec, "myapp has nil application spec"))
	})
	It("should pass for a valid application input", func() {
		conds := ValidateApplication(input)
		Expect(conds).To(HaveCondition(expConditionTypeRE, true, obs.ReasonValidationSuccess, `input.*is valid`))
	})

	Context("of includes and excludes", func() {
		It("should fail invalid exclude Namespaces", func() {
			input.Application.Excludes = []obs.NamespaceContainerSpec{
				{
					Namespace: "$my-namespace123_",
				},
				{
					Namespace: "bar",
				},
			}
			Expect(ValidateApplication(input)).To(HaveCondition(expConditionTypeRE, false, obs.ReasonValidationFailure, `globs.*match.*excludes.*namespace`))
		})
		It("should fail invalid container includes and excludes", func() {
			input.Application.Includes = []obs.NamespaceContainerSpec{
				{
					Container: "$my-namespace123_",
				},
			}
			input.Application.Excludes = []obs.NamespaceContainerSpec{
				{
					Container: "$my!123",
				},
			}
			Expect(ValidateApplication(input)).To(HaveCondition(expConditionTypeRE, false, obs.ReasonValidationFailure, `globs.*match.*excludes.*container.*includes.*container`))
		})
		It("should fail invalid container includes", func() {
			input.Application.Includes = []obs.NamespaceContainerSpec{
				{
					Container: "$my-namespace123_",
				},
				{
					Container: "bar",
				},
			}
			Expect(ValidateApplication(input)).To(HaveCondition(expConditionTypeRE, false, obs.ReasonValidationFailure, `globs.*match.*includes.*container`))
		})
		It("should fail invalid container excludes", func() {
			input.Application.Excludes = []obs.NamespaceContainerSpec{
				{
					Container: "$my-namespace123_",
				},
				{
					Container: "bar",
				},
			}
			Expect(ValidateApplication(input)).To(HaveCondition(expConditionTypeRE, false, obs.ReasonValidationFailure, `globs.*match.*excludes.*container`))
		})
		It("should pass when valid", func() {
			input.Application.Excludes = []obs.NamespaceContainerSpec{
				{
					Namespace: "my-namespace123",
					Container: "my-namespace123",
				},
				{
					Namespace: "my-namespace123",
					Container: "bar",
				},
				{
					Namespace: "my-namespace123",
					Container: "**one*with***stars*",
				},
			}
			input.Application.Includes = []obs.NamespaceContainerSpec{
				{
					Namespace: "my-namespace123",
					Container: "my-namespace123",
				},
				{
					Namespace: "bar",
					Container: "my-namespace123",
				},
				{
					Namespace: "**one*with***stars*",
					Container: "my-namespace123",
				},
			}
			Expect(ValidateApplication(input)).To(HaveCondition(expConditionTypeRE, true, obs.ReasonValidationSuccess, `input.*is valid`))
		})
	})
})
