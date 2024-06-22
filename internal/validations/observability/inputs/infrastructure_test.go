package inputs

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	. "github.com/openshift/cluster-logging-operator/test/matchers"
)

var _ = Describe("#ValidateInfrastructure", func() {

	var (
		input              obs.InputSpec
		expConditionTypeRE = obs.ConditionTypeValidInputPrefix + "-.*"
	)
	BeforeEach(func() {
		input = obs.InputSpec{
			Name: "myapp",
			Type: obs.InputTypeInfrastructure,
			Infrastructure: &obs.Infrastructure{
				Sources: []obs.InfrastructureSource{obs.InfrastructureSourceContainer},
			},
		}
	})
	It("should skip the validation when not an infrastructure type", func() {
		input.Type = obs.InputTypeApplication
		conds := ValidateInfrastructure(input)
		Expect(conds).To(BeEmpty())
	})
	It("should fail when an infrastructure type but has no infrastructure input", func() {
		input.Infrastructure = nil
		conds := ValidateInfrastructure(input)
		Expect(conds).To(HaveCondition(expConditionTypeRE, false, obs.ReasonMissingSpec, "myapp has nil infrastructure spec"))
	})
	It("should pass for a valid infrastructure input", func() {
		conds := ValidateInfrastructure(input)
		Expect(conds).To(HaveCondition(expConditionTypeRE, true, obs.ReasonValidationSuccess, `input.*is valid`))

	})
	It("should fail when no sources are defined", func() {
		input.Infrastructure.Sources = []obs.InfrastructureSource{}
		Expect(ValidateInfrastructure(input)).To(HaveCondition(expConditionTypeRE, false, obs.ReasonValidationFailure, "must define at least one valid source"))
	})
})
