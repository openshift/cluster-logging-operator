package observability_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	. "github.com/openshift/cluster-logging-operator/internal/api/observability"
)

var _ = Describe("#InputSources", func() {

	Context("for audit input type", func() {
		It("should return all audit sources when Sources is empty", func() {
			inputs := Inputs{
				{Name: "audit", Type: obs.InputTypeAudit, Audit: &obs.Audit{Sources: []obs.AuditSource{}}},
			}
			Expect(inputs.InputSources(obs.InputTypeAudit)).To(Equal(ReservedAuditSources.List()))
		})

		It("should return all audit sources when Sources is nil", func() {
			inputs := Inputs{
				{Name: "audit", Type: obs.InputTypeAudit, Audit: &obs.Audit{}},
			}
			Expect(inputs.InputSources(obs.InputTypeAudit)).To(Equal(ReservedAuditSources.List()))
		})

		It("should return all audit sources when Audit is nil", func() {
			inputs := Inputs{
				{Name: "audit", Type: obs.InputTypeAudit},
			}
			Expect(inputs.InputSources(obs.InputTypeAudit)).To(Equal(ReservedAuditSources.List()))
		})

		It("should return only specified sources when Sources is explicit", func() {
			inputs := Inputs{
				{Name: "audit", Type: obs.InputTypeAudit, Audit: &obs.Audit{Sources: []obs.AuditSource{obs.AuditSourceKube}}},
			}
			Expect(inputs.InputSources(obs.InputTypeAudit)).To(Equal([]string{obs.AuditSourceKube.String()}))
		})
	})

	Context("for infrastructure input type", func() {
		It("should return all infra sources when Sources is empty", func() {
			inputs := Inputs{
				{Name: "infra", Type: obs.InputTypeInfrastructure, Infrastructure: &obs.Infrastructure{Sources: []obs.InfrastructureSource{}}},
			}
			Expect(inputs.InputSources(obs.InputTypeInfrastructure)).To(Equal(ReservedInfrastructureSources.List()))
		})

		It("should return all infra sources when Infrastructure is nil", func() {
			inputs := Inputs{
				{Name: "infra", Type: obs.InputTypeInfrastructure},
			}
			Expect(inputs.InputSources(obs.InputTypeInfrastructure)).To(Equal(ReservedInfrastructureSources.List()))
		})

		It("should return only specified sources when Sources is explicit", func() {
			inputs := Inputs{
				{Name: "infra", Type: obs.InputTypeInfrastructure, Infrastructure: &obs.Infrastructure{Sources: []obs.InfrastructureSource{obs.InfrastructureSourceNode}}},
			}
			Expect(inputs.InputSources(obs.InputTypeInfrastructure)).To(Equal([]string{obs.InfrastructureSourceNode.String()}))
		})
	})
})
