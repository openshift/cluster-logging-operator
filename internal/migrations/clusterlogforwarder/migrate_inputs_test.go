package clusterlogforwarder

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	logging "github.com/openshift/cluster-logging-operator/api/logging/v1"
	"github.com/openshift/cluster-logging-operator/internal/constants"
)

var _ = Describe("migrateInputs", func() {
	Context("for reserved input names", func() {

		It("should stub 'application' as an input when referenced", func() {
			spec := logging.ClusterLogForwarderSpec{
				Pipelines: []logging.PipelineSpec{
					{InputRefs: []string{logging.InputNameApplication}},
				},
			}
			extras := map[string]bool{}
			result, _, _ := MigrateInputs("", "", spec, nil, extras, "", "")
			Expect(result.Inputs).To(HaveLen(1))
			Expect(result.Inputs[0]).To(Equal(logging.InputSpec{Name: logging.InputNameApplication, Application: &logging.Application{}}))
			Expect(extras).To(Equal(map[string]bool{constants.MigrateInputApplication: true}))
		})
		It("should stub 'infrastructure' as an input when referenced", func() {
			spec := logging.ClusterLogForwarderSpec{
				Pipelines: []logging.PipelineSpec{
					{InputRefs: []string{logging.InputNameInfrastructure}},
				},
			}
			extras := map[string]bool{}
			result, _, _ := MigrateInputs("", "", spec, nil, extras, "", "")
			Expect(result.Inputs).To(HaveLen(1))
			Expect(result.Inputs[0]).To(Equal(logging.InputSpec{Name: logging.InputNameInfrastructure,
				Infrastructure: &logging.Infrastructure{
					Sources: logging.InfrastructureSources.List(),
				}}))
			Expect(extras).To(Equal(map[string]bool{constants.MigrateInputInfrastructure: true}))
		})
		It("should stub 'audit' as an input when referenced", func() {
			spec := logging.ClusterLogForwarderSpec{
				Pipelines: []logging.PipelineSpec{
					{InputRefs: []string{logging.InputNameAudit}},
				},
			}
			extras := map[string]bool{}
			result, _, _ := MigrateInputs("", "", spec, nil, extras, "", "")
			Expect(result.Inputs).To(HaveLen(1))
			Expect(result.Inputs[0]).To(Equal(logging.InputSpec{Name: logging.InputNameAudit,
				Audit: &logging.Audit{
					Sources: logging.AuditSources.List(),
				}}))
			Expect(extras).To(Equal(map[string]bool{constants.MigrateInputAudit: true}))
		})
	})

	Context("for input receiver types", func() {
		var (
			http = &logging.HTTPReceiver{
				Port:   8080,
				Format: logging.FormatKubeAPIAudit,
			}
			syslog = &logging.SyslogReceiver{
				Port: 8080,
			}
		)

		It("should set 'http' receiver type HTTPReceiver declared", func() {
			spec := logging.ClusterLogForwarderSpec{
				Inputs: []logging.InputSpec{
					{
						Name: "my-custom-input",
						Receiver: &logging.ReceiverSpec{
							ReceiverTypeSpec: &logging.ReceiverTypeSpec{
								HTTP: http,
							},
						},
					},
				},
			}
			extras := map[string]bool{}
			result, _, _ := EnsureInputsHasType("", "", spec, nil, extras, "", "")
			Expect(result.Inputs).To(HaveLen(1))
			Expect(result.Inputs[0]).To(Equal(logging.InputSpec{
				Name: "my-custom-input",
				Receiver: &logging.ReceiverSpec{
					Type: logging.ReceiverTypeHttp,
					ReceiverTypeSpec: &logging.ReceiverTypeSpec{
						HTTP: http,
					},
				},
			}))
		})

		It("should do nothing if receiver type declared as 'http'", func() {
			spec := logging.ClusterLogForwarderSpec{
				Inputs: []logging.InputSpec{
					{
						Name: "my-custom-input",
						Receiver: &logging.ReceiverSpec{
							Type: logging.ReceiverTypeHttp,
							ReceiverTypeSpec: &logging.ReceiverTypeSpec{
								HTTP: http,
							},
						},
					},
				},
			}
			extras := map[string]bool{}
			result, _, _ := MigrateInputs("", "", spec, nil, extras, "", "")
			Expect(result.Inputs).To(HaveLen(1))
			Expect(result.Inputs[0]).To(Equal(logging.InputSpec{
				Name: "my-custom-input",
				Receiver: &logging.ReceiverSpec{
					Type: logging.ReceiverTypeHttp,
					ReceiverTypeSpec: &logging.ReceiverTypeSpec{
						HTTP: http,
					},
				},
			}))
			Expect(extras).To(BeEmpty())
		})

		It("should set 'syslog' receiver type SyslogReceiver declared", func() {
			spec := logging.ClusterLogForwarderSpec{
				Inputs: []logging.InputSpec{
					{
						Name: "my-custom-input",
						Receiver: &logging.ReceiverSpec{
							ReceiverTypeSpec: &logging.ReceiverTypeSpec{
								Syslog: syslog,
							},
						},
					},
				},
			}
			extras := map[string]bool{}
			result, _, _ := EnsureInputsHasType("", "", spec, nil, extras, "", "")
			Expect(result.Inputs).To(HaveLen(1))
			Expect(result.Inputs[0]).To(Equal(logging.InputSpec{
				Name: "my-custom-input",
				Receiver: &logging.ReceiverSpec{
					Type: logging.ReceiverTypeSyslog,
					ReceiverTypeSpec: &logging.ReceiverTypeSpec{
						Syslog: syslog,
					},
				},
			}))
		})

		It("should do nothing if receiver type declared as 'syslog'", func() {
			spec := logging.ClusterLogForwarderSpec{
				Inputs: []logging.InputSpec{
					{
						Name: "my-custom-input",
						Receiver: &logging.ReceiverSpec{
							Type: logging.ReceiverTypeSyslog,
							ReceiverTypeSpec: &logging.ReceiverTypeSpec{
								Syslog: syslog,
							},
						},
					},
				},
			}
			extras := map[string]bool{}
			result, _, _ := MigrateInputs("", "", spec, nil, extras, "", "")
			Expect(result.Inputs).To(HaveLen(1))
			Expect(result.Inputs[0]).To(Equal(logging.InputSpec{
				Name: "my-custom-input",
				Receiver: &logging.ReceiverSpec{
					Type: logging.ReceiverTypeSyslog,
					ReceiverTypeSpec: &logging.ReceiverTypeSpec{
						Syslog: syslog,
					},
				},
			}))
			Expect(extras).To(BeEmpty())
		})

		It("should do nothing if receiver.type declared as syslog but receiverTypeSpec is nil", func() {
			spec := logging.ClusterLogForwarderSpec{
				Inputs: []logging.InputSpec{
					{
						Name: "my-custom-input",
						Receiver: &logging.ReceiverSpec{
							Type: logging.ReceiverTypeSyslog,
						},
					},
				},
			}
			extras := map[string]bool{}
			result, _, _ := MigrateInputs("", "", spec, nil, extras, "", "")
			Expect(result.Inputs).To(HaveLen(1))
			Expect(result.Inputs[0]).To(Equal(logging.InputSpec{
				Name: "my-custom-input",
				Receiver: &logging.ReceiverSpec{
					Type: logging.ReceiverTypeSyslog,
				},
			}))
			Expect(extras).To(BeEmpty())
		})

	})
})
