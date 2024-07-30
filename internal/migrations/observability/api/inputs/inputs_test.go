package inputs

import (
	. "github.com/onsi/ginkgo"

	. "github.com/onsi/gomega"
	logging "github.com/openshift/cluster-logging-operator/api/logging/v1"
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
)

var _ = Describe("#ConvertInputs", func() {
	It("should convert all logging.Inputs into observability.Inputs", func() {
		loggingCLFSpec := &logging.ClusterLogForwarderSpec{}

		loggingCLFSpec.Inputs = []logging.InputSpec{
			{
				Name: "my-apps",
				Application: &logging.Application{
					Includes: []logging.NamespaceContainerSpec{
						{
							Namespace: "foo",
							Container: "bar",
						},
					},
					ContainerLimit: &logging.LimitSpec{
						MaxRecordsPerSecond: 100,
					},
				},
			},
			{
				Name: "my-infra",
				Infrastructure: &logging.Infrastructure{
					Sources: []string{"foo"},
				},
			},
			{
				Name: "my-audit",
				Audit: &logging.Audit{
					Sources: []string{"bar"},
				},
			},
			{
				Name: "my-http-receiver",
				Receiver: &logging.ReceiverSpec{
					Type: logging.ReceiverTypeHttp,
					ReceiverTypeSpec: &logging.ReceiverTypeSpec{
						HTTP: &logging.HTTPReceiver{
							Port:   5000,
							Format: logging.FormatKubeAPIAudit,
						},
					},
				},
			},
			{
				Name: "my-syslog-receiver",
				Receiver: &logging.ReceiverSpec{
					Type: logging.ReceiverTypeSyslog,
					ReceiverTypeSpec: &logging.ReceiverTypeSpec{
						Syslog: &logging.SyslogReceiver{
							Port: 9999,
						},
					},
				},
			},
		}

		expObsInputs := []obs.InputSpec{
			{
				Name: "my-apps",
				Type: obs.InputTypeApplication,
				Application: &obs.Application{
					Includes: []obs.NamespaceContainerSpec{
						{
							Namespace: "foo",
							Container: "bar",
						},
					},
					Tuning: &obs.ContainerInputTuningSpec{
						RateLimitPerContainer: &obs.LimitSpec{
							MaxRecordsPerSecond: 100,
						},
					},
				},
			},
			{
				Name: "my-infra",
				Type: obs.InputTypeInfrastructure,
				Infrastructure: &obs.Infrastructure{
					Sources: []obs.InfrastructureSource{"foo"},
				},
			},
			{
				Name: "my-audit",
				Type: obs.InputTypeAudit,
				Audit: &obs.Audit{
					Sources: []obs.AuditSource{"bar"},
				},
			},
			{
				Name: "my-http-receiver",
				Type: obs.InputTypeReceiver,
				Receiver: &obs.ReceiverSpec{
					Type: logging.ReceiverTypeHttp,
					Port: 5000,
					HTTP: &obs.HTTPReceiver{
						Format: logging.FormatKubeAPIAudit,
					},
				},
			},
			{
				Name: "my-syslog-receiver",
				Type: obs.InputTypeReceiver,
				Receiver: &obs.ReceiverSpec{
					Type: logging.ReceiverTypeSyslog,
					Port: 9999,
				},
			},
		}

		Expect(ConvertInputs(loggingCLFSpec)).To(Equal(expObsInputs))
	})
})
