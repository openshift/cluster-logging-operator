package inputs

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	. "github.com/openshift/cluster-logging-operator/test/matchers"
)

var _ = Describe("#ValidateInfrastructure", func() {

	var (
		input obs.InputSpec
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
	It("should fail when an audit type but has no infrastructure input", func() {
		input.Audit = nil
		conds := ValidateAudit(input)
		Expect(conds).To(HaveCondition(obs.ValidationCondition, true, obs.ReasonMissingSpec, "myapp has nil audit spec"))
	})
	It("should pass for a valid audit input", func() {
		conds := ValidateAudit(input)
		Expect(conds).To(BeEmpty())
	})
	It("should fail when no sources are defined", func() {
		input.Audit.Sources = []obs.AuditSource{}
		Expect(ValidateAudit(input)).To(HaveCondition(obs.ValidationCondition, true, obs.ReasonMissingSources, "must define at least one valid source"))
	})
})
