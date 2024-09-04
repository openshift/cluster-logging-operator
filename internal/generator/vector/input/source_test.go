package input

import (
	"fmt"
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	"github.com/openshift/cluster-logging-operator/internal/factory"
	"github.com/openshift/cluster-logging-operator/internal/generator/framework"
	. "github.com/openshift/cluster-logging-operator/test/matchers"
)

var _ = Describe("inputs", func() {

	const (
		secretName = "instance-myreceiver"
	)

	var (
		secrets = map[string]*corev1.Secret{
			secretName: {
				Data: map[string][]byte{
					constants.ClientCertKey:      []byte("-- crt-- "),
					constants.ClientPrivateKey:   []byte("-- key-- "),
					constants.TrustedCABundleKey: []byte("-- ca-bundle -- "),
					constants.Passphrase:         []byte("foo"),
				},
			},
		}
	)

	DescribeTable("#NewSource", func(input obs.InputSpec, expFile string) {
		exp, err := tomlContent.ReadFile(expFile)
		if err != nil {
			Fail(fmt.Sprintf("Error reading the file %q with exp config: %v", expFile, err))
		}
		clf := obs.ClusterLogForwarder{
			ObjectMeta: metav1.ObjectMeta{
				Name:      constants.SingletonName,
				Namespace: constants.OpenshiftNS,
			},
		}
		conf, _ := NewSource(input, constants.OpenshiftNS, *factory.ResourceNames(clf), secrets, framework.NoOptions)
		Expect(string(exp)).To(EqualConfigFrom(conf))
	},
		Entry("with an application input should generate a container source", obs.InputSpec{
			Name: string(obs.InputTypeApplication),
			Type: obs.InputTypeApplication,
		},
			"application.toml",
		),
		Entry("with an application input should generate a container source", obs.InputSpec{
			Name:        string(obs.InputTypeApplication),
			Type:        obs.InputTypeApplication,
			Application: &obs.Application{},
		},
			"application.toml",
		),
		Entry("with a throttled application input should generate a container source with throttling", obs.InputSpec{
			Name: string(obs.InputTypeApplication),
			Type: obs.InputTypeApplication,
			Application: &obs.Application{
				Tuning: &obs.ContainerInputTuningSpec{
					RateLimitPerContainer: &obs.LimitSpec{
						MaxRecordsPerSecond: 1024,
					},
				},
			},
		},
			"application_with_throttle.toml",
		),
		Entry("with an application that specs including a container from all namespaces", obs.InputSpec{
			Name: "my-app",
			Type: obs.InputTypeApplication,
			Application: &obs.Application{
				Includes: []obs.NamespaceContainerSpec{
					{
						Container: "log-*",
					},
				},
			},
		},
			"application_includes_container.toml",
		),
		Entry("with an application that specs excluding a container from all namespaces", obs.InputSpec{
			Name: "my-app",
			Type: obs.InputTypeApplication,
			Application: &obs.Application{
				Excludes: []obs.NamespaceContainerSpec{
					{
						Container: "log-*",
					},
				},
			},
		},
			"application_excludes_container.toml",
		),
		Entry("with an application that specs specific namespaces and exclude namespaces", obs.InputSpec{
			Name: "my-app",
			Type: obs.InputTypeApplication,
			Application: &obs.Application{
				Excludes: []obs.NamespaceContainerSpec{
					{
						Namespace: "test-ns1",
						Container: "mesh*",
					},
					{
						Namespace: "test-ns2",
						Container: "mesh*",
					},
				},
				Includes: []obs.NamespaceContainerSpec{
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
		Entry("with an application that specs infra namespaces and exclude namespaces", obs.InputSpec{
			Name: "my-app",
			Type: obs.InputTypeApplication,
			Application: &obs.Application{
				Excludes: []obs.NamespaceContainerSpec{
					{
						Namespace: "test-ns1",
					},
				},
				Includes: []obs.NamespaceContainerSpec{
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
		Entry("with an application that collects infra namespaces and excludes a container", obs.InputSpec{
			Name: "my-app",
			Type: obs.InputTypeApplication,
			Application: &obs.Application{
				Excludes: []obs.NamespaceContainerSpec{
					{
						Namespace: "openshift-logging",
						Container: "mesh",
					},
				},
				Includes: []obs.NamespaceContainerSpec{
					{
						Namespace: "openshift-logging",
					},
				},
			},
		},
			"application_exclude_container_from_infra.toml",
		),
		Entry("with an application that specs infra namespaces and excludes infra namespaces", obs.InputSpec{
			Name: "my-app",
			Type: obs.InputTypeApplication,
			Application: &obs.Application{
				Excludes: []obs.NamespaceContainerSpec{
					{
						Namespace: "test-ns1",
					},
					{
						Namespace: "openshift-logging",
					},
				},
				Includes: []obs.NamespaceContainerSpec{
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
		Entry("with an application that specs specific infra namespace and excludes infra namespaces", obs.InputSpec{
			Name: "my-app",
			Type: obs.InputTypeApplication,
			Application: &obs.Application{
				Excludes: []obs.NamespaceContainerSpec{
					{
						Namespace: "test-ns1",
					},
					{
						Namespace: "openshift*",
					},
				},
				Includes: []obs.NamespaceContainerSpec{
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
		Entry("with an application that specs specific match labels", obs.InputSpec{
			Name: "my-app",
			Type: obs.InputTypeApplication,
			Application: &obs.Application{
				Selector: &metav1.LabelSelector{
					MatchLabels: map[string]string{
						"key1": "value1",
						"key2": "value2",
					},
				},
			},
		},
			"application_with_matchLabels.toml",
		),
		Entry("with an infrastructure input should generate a container and journal source", obs.InputSpec{
			Name: string(obs.InputTypeInfrastructure),
			Type: obs.InputTypeInfrastructure,
		},
			"infrastructure.toml",
		),
		Entry("with an infrastructure input should generate a container and journal source", obs.InputSpec{
			Name:           string(obs.InputTypeInfrastructure),
			Type:           obs.InputTypeInfrastructure,
			Infrastructure: &obs.Infrastructure{},
		},
			"infrastructure.toml",
		),
		Entry("with an infrastructure input for containers should generate only a container source", obs.InputSpec{
			Name: "myinfra",
			Type: obs.InputTypeInfrastructure,
			Infrastructure: &obs.Infrastructure{
				Sources: []obs.InfrastructureSource{obs.InfrastructureSourceContainer},
			},
		},
			"infrastructure_container.toml",
		),
		Entry("with an infrastructure input for node should generate only a journal source", obs.InputSpec{
			Name: "myinfra",
			Type: obs.InputTypeInfrastructure,
			Infrastructure: &obs.Infrastructure{
				Sources: []obs.InfrastructureSource{obs.InfrastructureSourceNode},
			},
		},
			"infrastructure_journal.toml",
		),
		Entry("with an audit input should generate file sources", obs.InputSpec{
			Name:  string(obs.InputTypeAudit),
			Type:  obs.InputTypeAudit,
			Audit: &obs.Audit{},
		},
			"audit.toml",
		),
		Entry("with an audit input should generate file sources", obs.InputSpec{
			Name: string(obs.InputTypeAudit),
			Type: obs.InputTypeAudit,
		},
			"audit.toml",
		),
		Entry("with an audit input for auditd logs should generate auditd file source", obs.InputSpec{
			Name: "myaudit",
			Type: obs.InputTypeAudit,
			Audit: &obs.Audit{
				Sources: []obs.AuditSource{obs.AuditSourceAuditd},
			},
		},
			"audit_host.toml",
		),
		Entry("with an audit input for kube logs should generate kube audit file source", obs.InputSpec{
			Name: "myaudit",
			Type: obs.InputTypeAudit,
			Audit: &obs.Audit{
				Sources: []obs.AuditSource{obs.AuditSourceKube},
			},
		},
			"audit_kube.toml",
		),
		Entry("with an audit input for openshift logs should generate openshift audit file source", obs.InputSpec{
			Name: "myaudit",
			Type: obs.InputTypeAudit,
			Audit: &obs.Audit{
				Sources: []obs.AuditSource{obs.AuditSourceOpenShift},
			},
		},
			"audit_openshift.toml",
		),
		Entry("with an audit input for OVN logs should generate OVN audit file source", obs.InputSpec{
			Name: "myaudit",
			Type: obs.InputTypeAudit,
			Audit: &obs.Audit{
				Sources: []obs.AuditSource{obs.AuditSourceOVN},
			},
		},
			"audit_ovn.toml",
		),
		Entry("with an http audit receiver input should generate an http receiver audit source", obs.InputSpec{
			Type: obs.InputTypeReceiver,
			Name: "myreceiver",
			Receiver: &obs.ReceiverSpec{
				Type: obs.ReceiverTypeHTTP,
				Port: 12345,
				HTTP: &obs.HTTPReceiver{
					Format: obs.HTTPReceiverFormatKubeAPIAudit,
				},
				TLS: &obs.InputTLSSpec{
					Certificate: &obs.ValueReference{
						Key:        constants.ClientCertKey,
						SecretName: secretName,
					},
					Key: &obs.SecretReference{
						Key:        constants.ClientPrivateKey,
						SecretName: secretName,
					},
				},
			},
		},
			"receiver_http_audit.toml",
		),
		Entry("with a syslog receiver input should generate VIAQ syslog receiver", obs.InputSpec{
			Type: obs.InputTypeReceiver,
			Name: "myreceiver",
			Receiver: &obs.ReceiverSpec{
				Type: obs.ReceiverTypeSyslog,
				Port: 12345,
				TLS: &obs.InputTLSSpec{
					CA: &obs.ValueReference{
						Key:        constants.TrustedCABundleKey,
						SecretName: secretName,
					},
					Certificate: &obs.ValueReference{
						Key:        constants.ClientCertKey,
						SecretName: secretName,
					},
					Key: &obs.SecretReference{
						Key:        constants.ClientPrivateKey,
						SecretName: secretName,
					},
					KeyPassphrase: &obs.SecretReference{
						Key:        constants.Passphrase,
						SecretName: secretName,
					},
				},
			},
		},
			"receiver_syslog.toml",
		),
		Entry("with a syslog receiver and tls from configmaps", obs.InputSpec{
			Type: obs.InputTypeReceiver,
			Name: "myreceiver",
			Receiver: &obs.ReceiverSpec{
				Type: obs.ReceiverTypeSyslog,
				Port: 12345,
				TLS: &obs.InputTLSSpec{
					CA: &obs.ValueReference{
						Key:           "ca.crt",
						ConfigMapName: secretName,
					},
					Certificate: &obs.ValueReference{
						Key:           "my.crt",
						ConfigMapName: secretName,
					},
					Key: &obs.SecretReference{
						Key:        constants.ClientPrivateKey,
						SecretName: secretName,
					},
				},
			},
		},
			"receiver_syslog_tls_from_configmap.toml",
		),
	)
})
