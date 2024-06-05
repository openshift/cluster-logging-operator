package api

import (
	. "github.com/onsi/ginkgo"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	. "github.com/onsi/gomega"
	logging "github.com/openshift/cluster-logging-operator/api/logging/v1"
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
)

var _ = Describe("#ConvertInputs", func() {
	It("should map logging.NamespaceContainerSpec to observability.NamespaceContainerSpec", func() {
		loggingNSCont := []logging.NamespaceContainerSpec{
			{
				Namespace: "foo",
				Container: "bar",
			},
			{
				Namespace: "baz",
				Container: "foobar",
			},
		}

		expObsNSCont := []obs.NamespaceContainerSpec{
			{
				Namespace: "foo",
				Container: "bar",
			},
			{
				Namespace: "baz",
				Container: "foobar",
			},
		}

		Expect(mapNamespacedContainers(loggingNSCont)).To(Equal(expObsNSCont))
	})

	It("should map logging.Application to observability.Application", func() {
		loggingApp := logging.Application{
			Selector: &logging.LabelSelector{
				MatchLabels: map[string]string{},
				MatchExpressions: []metav1.LabelSelectorRequirement{
					{
						Key:      "foo",
						Operator: metav1.LabelSelectorOpIn,
						Values:   []string{"bar"},
					},
				},
			},
			Includes: []logging.NamespaceContainerSpec{
				{
					Namespace: "foo",
					Container: "bar",
				},
			},
			Excludes: []logging.NamespaceContainerSpec{
				{
					Namespace: "bz",
					Container: "fz",
				},
			},
			ContainerLimit: &logging.LimitSpec{
				MaxRecordsPerSecond: 1000,
			},
		}

		expObsApp := &obs.Application{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{},
				MatchExpressions: []metav1.LabelSelectorRequirement{
					{
						Key:      "foo",
						Operator: metav1.LabelSelectorOpIn,
						Values:   []string{"bar"},
					},
				},
			},
			Includes: []obs.NamespaceContainerSpec{
				{
					Namespace: "foo",
					Container: "bar",
				},
			},
			Excludes: []obs.NamespaceContainerSpec{
				{
					Namespace: "bz",
					Container: "fz",
				},
			},
			Tuning: &obs.ContainerInputTuningSpec{
				RateLimitPerContainer: &obs.LimitSpec{
					MaxRecordsPerSecond: 1000,
				},
			},
		}
		Expect(mapApplicationInput(&loggingApp)).To(Equal(expObsApp))
	})

	It("should map logging.Infrastructure to observability.Infrastructure", func() {
		loggingInfra := logging.Infrastructure{
			Sources: []string{"foo", "bar", "baz"},
		}

		expObsInfra := &obs.Infrastructure{
			Sources: []obs.InfrastructureSource{"foo", "bar", "baz"},
		}

		Expect(mapInfrastructureInput(&loggingInfra)).To(Equal(expObsInfra))
	})

	It("should map logging.Audit to observability.Audit", func() {
		loggingAudit := logging.Audit{
			Sources: []string{"foo", "bar", "baz"},
		}

		expObsAudit := &obs.Audit{
			Sources: []obs.AuditSource{"foo", "bar", "baz"},
		}

		Expect(mapAuditInput(&loggingAudit)).To(Equal(expObsAudit))
	})

	It("should map logging http receiver to observability http receiver", func() {
		loggingReceiverSpec := logging.ReceiverSpec{
			Type: logging.ReceiverTypeHttp,
			ReceiverTypeSpec: &logging.ReceiverTypeSpec{
				HTTP: &logging.HTTPReceiver{
					Port:   9000,
					Format: logging.FormatKubeAPIAudit,
				},
			},
		}
		expObsReceiverSpec := &obs.ReceiverSpec{
			Type: obs.ReceiverTypeHTTP,
			Port: 9000,
			HTTP: &obs.HTTPReceiver{
				Format: obs.HTTPReceiverFormatKubeAPIAudit,
			},
		}

		Expect(mapReceiverInput(&loggingReceiverSpec)).To(Equal(expObsReceiverSpec))
	})

	It("should map logging syslog receiver to observability syslog receiver", func() {
		loggingReceiverSpec := logging.ReceiverSpec{
			Type: logging.ReceiverTypeSyslog,
			ReceiverTypeSpec: &logging.ReceiverTypeSpec{
				Syslog: &logging.SyslogReceiver{
					Port: 9000,
				},
			},
		}
		expObsReceiverSpec := &obs.ReceiverSpec{
			Type: obs.ReceiverTypeSyslog,
			Port: 9000,
		}

		Expect(mapReceiverInput(&loggingReceiverSpec)).To(Equal(expObsReceiverSpec))
	})

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

		Expect(convertInputs(loggingCLFSpec)).To(Equal(expObsInputs))
	})
})
