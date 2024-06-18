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
			Type: obs.InputTypeAudit,
			Audit: &obs.Audit{
				Sources: []obs.AuditSource{obs.AuditSourceKube},
			},
		}
	})
	It("should skip the validation when not an audit type", func() {
		input.Type = obs.InputTypeApplication
		conds := ValidateAudit(input)
		Expect(conds).To(BeEmpty())
	})
	It("should fail when an audit type but has no audit input", func() {
		input.Audit = nil
		conds := ValidateAudit(input)
		Expect(conds).To(HaveCondition(expConditionTypeRE, false, obs.ReasonMissingSpec, "myapp has nil audit spec"))
	})
	It("should pass for a valid audit input", func() {
		conds := ValidateAudit(input)
		Expect(conds).To(HaveCondition(expConditionTypeRE, true, obs.ReasonValidationSuccess, `input.*is valid`))
	})
	It("should fail when no sources are defined", func() {
		input.Audit.Sources = []obs.AuditSource{}
		Expect(ValidateAudit(input)).To(HaveCondition(expConditionTypeRE, false, obs.ReasonValidationFailure, "must define at least one valid source"))
	})
})
