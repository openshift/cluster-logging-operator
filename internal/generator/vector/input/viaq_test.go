package input

import (
	"fmt"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	logging "github.com/openshift/cluster-logging-operator/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	"github.com/openshift/cluster-logging-operator/internal/generator/framework"
	. "github.com/openshift/cluster-logging-operator/test/matchers"
)

var _ = Describe("inputs", func() {
	DescribeTable("#NewViaQ", func(input logging.InputSpec, expFile string) {
		exp, err := tomlContent.ReadFile(expFile)
		if err != nil {
			Fail(fmt.Sprintf("Error reading the file %q with exp config: %v", expFile, err))
		}
		conf, _ := NewViaQ(input, constants.OpenshiftNS, framework.NoOptions)
		Expect(string(exp)).To(EqualConfigFrom(conf))
	},
		Entry("with an application input should generate a VIAQ container source", logging.InputSpec{
			Name:        logging.InputNameApplication,
			Application: &logging.Application{},
		},
			"viaq_application.toml",
		),
		Entry("with a throttled application input should generate a VIAQ container with throttling", logging.InputSpec{
			Name: logging.InputNameApplication,
			Application: &logging.Application{
				ContainerLimit: &logging.LimitSpec{
					MaxRecordsPerSecond: 1024,
				},
			},
		},
			"viaq_application_with_throttle.toml",
		),
		Entry("with an application that specs specific namespaces", logging.InputSpec{
			Name: "my-app",
			Application: &logging.Application{
				Namespaces: []string{"test-ns1", "test-ns2"},
				Containers: &logging.InclusionSpec{
					Include: []string{"mesh*"},
				},
			},
		},
			"viaq_application_with_includes.toml",
		),
		Entry("with an application that specs specific exclude namespaces", logging.InputSpec{
			Name: "my-app",
			Application: &logging.Application{
				ExcludeNamespaces: []string{"test-ns1", "test-ns2"},
				Containers: &logging.InclusionSpec{
					Exclude: []string{"mesh*"},
				},
			},
		},
			"viaq_application_with_excludes.toml",
		),
		Entry("with an application that specs specific match labels", logging.InputSpec{
			Name: "my-app",
			Application: &logging.Application{
				Selector: &logging.LabelSelector{
					MatchLabels: map[string]string{
						"key1": "value1",
						"key2": "value2",
					},
				},
			},
		},
			"viaq_application_with_matchLabels.toml",
		),
		Entry("with an infrastructure input should generate a VIAQ container and journal source", logging.InputSpec{
			Name:           logging.InputNameInfrastructure,
			Infrastructure: &logging.Infrastructure{},
		},
			"viaq_infrastructure.toml",
		),
		Entry("with an infrastructure input for containers should generate only a VIAQ container source", logging.InputSpec{
			Name: "myinfra",
			Infrastructure: &logging.Infrastructure{
				Sources: []string{logging.InfrastructureSourceContainer},
			},
		},
			"viaq_infrastructure_container.toml",
		),
		Entry("with an infrastructure input for node should generate only a VIAQ journal source", logging.InputSpec{
			Name: "myinfra",
			Infrastructure: &logging.Infrastructure{
				Sources: []string{logging.InfrastructureSourceNode},
			},
		},
			"viaq_infrastructure_journal.toml",
		),
		Entry("with an audit input should generate VIAQ file sources", logging.InputSpec{
			Name:  logging.InputNameAudit,
			Audit: &logging.Audit{},
		},
			"viaq_audit.toml",
		),
		Entry("with an audit input for auditd logs should generate VIAQ auditd file source", logging.InputSpec{
			Name: "myaudit",
			Audit: &logging.Audit{
				Sources: []string{logging.AuditSourceAuditd},
			},
		},
			"viaq_audit_host.toml",
		),
		Entry("with an audit input for kube logs should generate VIAQ kube audit file source", logging.InputSpec{
			Name: "myaudit",
			Audit: &logging.Audit{
				Sources: []string{logging.AuditSourceKube},
			},
		},
			"viaq_audit_kube.toml",
		),
		Entry("with an audit input for openshift logs should generate VIAQ openshift audit file source", logging.InputSpec{
			Name: "myaudit",
			Audit: &logging.Audit{
				Sources: []string{logging.AuditSourceOpenShift},
			},
		},
			"viaq_audit_openshift.toml",
		),
		Entry("with an audit input for OVN logs should generate VIAQ OVN audit file source", logging.InputSpec{
			Name: "myaudit",
			Audit: &logging.Audit{
				Sources: []string{logging.AuditSourceOVN},
			},
		},
			"viaq_audit_ovn.toml",
		),
		Entry("with an http audit receiver input should generate VIAQ http receiver audit source", logging.InputSpec{
			Name: "myreceiver",
			Receiver: &logging.ReceiverSpec{
				Type: logging.ReceiverTypeHttp,
				ReceiverTypeSpec: &logging.ReceiverTypeSpec{
					HTTP: &logging.HTTPReceiver{
						Port:   12345,
						Format: logging.FormatKubeAPIAudit,
					},
				},
			},
		},
			"viaq_receiver_http_audit.toml",
		),
		Entry("with a syslog receiver input should generate VIAQ syslog receiver", logging.InputSpec{
			Name: "myreceiver",
			Receiver: &logging.ReceiverSpec{
				Type: logging.ReceiverTypeSyslog,
				ReceiverTypeSpec: &logging.ReceiverTypeSpec{
					Syslog: &logging.SyslogReceiver{
						Port: 12345,
					},
				},
			},
		},
			"viaq_receiver_syslog.toml",
		),
	)
})
