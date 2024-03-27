package input

import (
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	logging "github.com/openshift/cluster-logging-operator/api/logging/v1"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	"github.com/openshift/cluster-logging-operator/internal/factory"
	"github.com/openshift/cluster-logging-operator/internal/generator/framework"
	. "github.com/openshift/cluster-logging-operator/test/matchers"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("inputs", func() {
	DescribeTable("#NewSource", func(input logging.InputSpec, expFile string) {
		exp, err := tomlContent.ReadFile(expFile)
		if err != nil {
			Fail(fmt.Sprintf("Error reading the file %q with exp config: %v", expFile, err))
		}
		clf := logging.ClusterLogForwarder{
			ObjectMeta: v1.ObjectMeta{
				Name:      constants.SingletonName,
				Namespace: constants.OpenshiftNS,
			},
		}
		conf, _ := NewSource(input, constants.OpenshiftNS, factory.GenerateResourceNames(clf), framework.NoOptions)
		Expect(string(exp)).To(EqualConfigFrom(conf))
	},
		Entry("with an application input should generate a container source", logging.InputSpec{
			Name:        logging.InputNameApplication,
			Application: &logging.Application{},
		},
			"application.toml",
		),
		Entry("with a throttled application input should generate a VIAQ container with throttling", logging.InputSpec{
			Name: logging.InputNameApplication,
			Application: &logging.Application{
				ContainerLimit: &logging.LimitSpec{
					MaxRecordsPerSecond: 1024,
				},
			},
		},
			"application_with_throttle.toml",
		),
		Entry("[migrate deprecated] with an application that specs specific namespaces", logging.InputSpec{
			Name: "my-app",
			Application: &logging.Application{
				Namespaces: []string{"test-ns1", "test-ns2"},
			},
		},
			"application_with_includes.toml",
		),
		Entry("with an application that specs specific exclude namespaces and containers", logging.InputSpec{
			Name: "my-app",
			Application: &logging.Application{
				Excludes: []logging.NamespaceContainerSpec{
					{
						Namespace: "test-ns1",
						Container: "mesh*",
					},
					{
						Namespace: "test-ns2",
						Container: "mesh*",
					},
				},
			},
		},
			"application_with_excludes.toml",
		),
		Entry("with an application that specs including a container from all namespaces", logging.InputSpec{
			Name: "my-app",
			Application: &logging.Application{
				Includes: []logging.NamespaceContainerSpec{
					{
						Container: "log-*",
					},
				},
			},
		},
			"application_includes_container.toml",
		),
		Entry("with an application that specs excluding a container from all namespaces", logging.InputSpec{
			Name: "my-app",
			Application: &logging.Application{
				Excludes: []logging.NamespaceContainerSpec{
					{
						Container: "log-*",
					},
				},
			},
		},
			"application_excludes_container.toml",
		),
		Entry("with an application that specs specific namespaces and exclude namespaces", logging.InputSpec{
			Name: "my-app",
			Application: &logging.Application{
				Excludes: []logging.NamespaceContainerSpec{
					{
						Namespace: "test-ns1",
						Container: "mesh*",
					},
					{
						Namespace: "test-ns2",
						Container: "mesh*",
					},
				},
				Includes: []logging.NamespaceContainerSpec{
					{
						Namespace: "test-ns-foo",
					},
					{
						Namespace: "test-ns-bar",
					},
				},
			},
		},
			"application_with_includes_excludes.toml",
		),
		Entry("with an application that specs infra namespaces and exclude namespaces", logging.InputSpec{
			Name: "my-app",
			Application: &logging.Application{
				Excludes: []logging.NamespaceContainerSpec{
					{
						Namespace: "test-ns1",
					},
				},
				Includes: []logging.NamespaceContainerSpec{
					{
						Namespace: "test-ns-foo",
						Container: "mesh",
					},
					{
						Namespace: "openshift-logging",
						Container: "mesh",
					},
					{
						Namespace: "kube-apiserver",
						Container: "mesh",
					},
				},
			},
		},
			"application_with_infra_includes_excludes.toml",
		),
		Entry("with an application that collects infra namespaces and excludes a container", logging.InputSpec{
			Name: "my-app",
			Application: &logging.Application{
				Excludes: []logging.NamespaceContainerSpec{
					{
						Namespace: "openshift-logging",
						Container: "mesh",
					},
				},
				Includes: []logging.NamespaceContainerSpec{
					{
						Namespace: "openshift-logging",
					},
				},
			},
		},
			"application_exclude_container_from_infra.toml",
		),
		Entry("with an application that specs infra namespaces and excludes infra namespaces", logging.InputSpec{
			Name: "my-app",
			Application: &logging.Application{
				Excludes: []logging.NamespaceContainerSpec{
					{
						Namespace: "test-ns1",
					},
					{
						Namespace: "openshift-logging",
					},
				},
				Includes: []logging.NamespaceContainerSpec{
					{
						Namespace: "test-ns-foo",
					},
					{
						Namespace: "openshift*",
					},
					{
						Namespace: "kube-apiserver",
					},
				},
			},
		},
			"application_with_infra_includes_infra_excludes.toml",
		),
		Entry("with an application that specs specific infra namespace and excludes infra namespaces", logging.InputSpec{
			Name: "my-app",
			Application: &logging.Application{
				Excludes: []logging.NamespaceContainerSpec{
					{
						Namespace: "test-ns1",
					},
					{
						Namespace: "openshift*",
					},
				},
				Includes: []logging.NamespaceContainerSpec{
					{
						Namespace: "test-ns-foo",
					},
					{
						Namespace: "openshift-logging",
					},
					{
						Namespace: "kube-apiserver",
					},
				},
			},
		},
			"application_with_specific_infra_includes_infra_excludes.toml",
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
			"application_with_matchLabels.toml",
		),
		Entry("with an infrastructure input should generate a container and journal source", logging.InputSpec{
			Name:           logging.InputNameInfrastructure,
			Infrastructure: &logging.Infrastructure{},
		},
			"infrastructure.toml",
		),
		Entry("with an infrastructure input for containers should generate only a container source", logging.InputSpec{
			Name: "myinfra",
			Infrastructure: &logging.Infrastructure{
				Sources: []string{logging.InfrastructureSourceContainer},
			},
		},
			"infrastructure_container.toml",
		),
		Entry("with an infrastructure input for node should generate only a journal source", logging.InputSpec{
			Name: "myinfra",
			Infrastructure: &logging.Infrastructure{
				Sources: []string{logging.InfrastructureSourceNode},
			},
		},
			"infrastructure_journal.toml",
		),
		Entry("with an audit input should generate file sources", logging.InputSpec{
			Name:  logging.InputNameAudit,
			Audit: &logging.Audit{},
		},
			"audit.toml",
		),
		Entry("with an audit input for auditd logs should generate auditd file source", logging.InputSpec{
			Name: "myaudit",
			Audit: &logging.Audit{
				Sources: []string{logging.AuditSourceAuditd},
			},
		},
			"audit_host.toml",
		),
		Entry("with an audit input for kube logs should generate kube audit file source", logging.InputSpec{
			Name: "myaudit",
			Audit: &logging.Audit{
				Sources: []string{logging.AuditSourceKube},
			},
		},
			"audit_kube.toml",
		),
		Entry("with an audit input for openshift logs should generate openshift audit file source", logging.InputSpec{
			Name: "myaudit",
			Audit: &logging.Audit{
				Sources: []string{logging.AuditSourceOpenShift},
			},
		},
			"audit_openshift.toml",
		),
		Entry("with an audit input for OVN logs should generate OVN audit file source", logging.InputSpec{
			Name: "myaudit",
			Audit: &logging.Audit{
				Sources: []string{logging.AuditSourceOVN},
			},
		},
			"audit_ovn.toml",
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
			"receiver_http_audit.toml",
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
			"receiver_syslog.toml",
		),
	)
})
