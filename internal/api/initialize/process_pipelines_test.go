package initialize

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
)

var _ = Describe("[internal][api][initialize]", func() {
	Context("getInputTypeFromName", func() {
		var (
			forwarder = obs.ClusterLogForwarder{
				Spec: obs.ClusterLogForwarderSpec{
					Inputs: []obs.InputSpec{{}},
				},
			}
		)
		Context("reserved named inputs", func() {
			forwarder.Spec.Inputs = []obs.InputSpec{
				{
					Name: string(obs.InputTypeApplication),
				},
				{
					Name: string(obs.InputTypeInfrastructure),
				},
				{
					Name: string(obs.InputTypeAudit),
				},
			}
			It("should return type application when name is application", func() {
				inputType := getInputTypeFromName(forwarder.Spec, "application")
				Expect(inputType).To(Equal(string(obs.InputTypeApplication)))
			})
			It("should return type infrastructure when name is infrastructure", func() {
				inputType := getInputTypeFromName(forwarder.Spec, "infrastructure")
				Expect(inputType).To(Equal(string(obs.InputTypeInfrastructure)))
			})
			It("should return type audit when name is audit", func() {
				inputType := getInputTypeFromName(forwarder.Spec, "audit")
				Expect(inputType).To(Equal(string(obs.InputTypeAudit)))
			})
		})

		Context("named inputs", func() {
			forwarder.Spec.Inputs = []obs.InputSpec{
				{
					Name: "my-app",
					Type: obs.InputTypeApplication,
					Application: &obs.Application{
						Includes: []obs.NamespaceContainerSpec{
							{
								Container: "foo",
							},
						},
					},
				},
				{
					Name: "my-infra",
					Type: obs.InputTypeInfrastructure,
					Infrastructure: &obs.Infrastructure{
						Sources: []obs.InfrastructureSource{obs.InfrastructureSourceContainer},
					},
				},
				{
					Name: "my-audit",
					Type: obs.InputTypeAudit,
					Audit: &obs.Audit{
						Sources: []obs.AuditSource{obs.AuditSourceAuditd},
					},
				},
				{
					Name: "my-syslog-receiver",
					Type: obs.InputTypeReceiver,
					Receiver: &obs.ReceiverSpec{
						Type: obs.ReceiverTypeSyslog,
						Port: 10514,
					},
				},
				{
					Name: "my-http-receiver",
					Type: obs.InputTypeReceiver,
					Receiver: &obs.ReceiverSpec{
						Type: obs.ReceiverTypeHTTP,
						Port: 8080,
						HTTP: &obs.HTTPReceiver{
							Format: obs.HTTPReceiverFormatKubeAPIAudit,
						},
					},
				},
			}
			It("should return type application when type is application", func() {
				inputType := getInputTypeFromName(forwarder.Spec, "my-app")
				Expect(inputType).To(Equal(string(obs.InputTypeApplication)))
			})
			It("should return type infrastructure when type is infrastructure", func() {
				inputType := getInputTypeFromName(forwarder.Spec, "my-infra")
				Expect(inputType).To(Equal(string(obs.InputTypeInfrastructure)))
			})
			It("should return type infrastructure when type is syslog receiver", func() {
				inputType := getInputTypeFromName(forwarder.Spec, "my-syslog-receiver")
				Expect(inputType).To(Equal(string(obs.InputTypeInfrastructure)))
			})
			It("should return type audit when type is audit", func() {
				inputType := getInputTypeFromName(forwarder.Spec, "my-audit")
				Expect(inputType).To(Equal(string(obs.InputTypeAudit)))
			})
			It("should return type audit when type is http receiver with kubeAPIAudit format", func() {
				inputType := getInputTypeFromName(forwarder.Spec, "my-http-receiver")
				Expect(inputType).To(Equal(string(obs.InputTypeAudit)))
			})
			// TODO: update method to return type, and handle this scenario in validation
			It("should return empty when input name not found?", func() {
				inputType := getInputTypeFromName(forwarder.Spec, "notfound")
				Expect(inputType).To(BeEmpty())
			})
		})
	})
})
