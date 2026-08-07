package http

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	configv1 "github.com/openshift/api/config/v1"
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	"github.com/openshift/cluster-logging-operator/internal/runtime"
	obsruntime "github.com/openshift/cluster-logging-operator/internal/runtime/observability"
	framework "github.com/openshift/cluster-logging-operator/test/framework/e2e"
	corev1 "k8s.io/api/core/v1"
)

var _ = Describe("[ClusterLogForwarder] Forwards logs with TLS curves", func() {
	const (
		forwarderName = "my-forwarder"
		logGenName    = "log-generator"
	)

	var (
		e2e            = framework.NewE2ETestFramework()
		deployNS       string
		logGenNS       string
		serviceAccount *corev1.ServiceAccount
	)

	AfterEach(func() {
		e2e.Cleanup()
		e2e.WaitForCleanupCompletion(logGenNS, []string{"test"})
	})

	It("should send logs to vector HTTP receiver with custom TLS profile including curves (LOG-9077)", func() {
		deployNS = e2e.CreateTestNamespace()

		// Receiver accepts a broad set of ciphers and curves
		receiverProfile := configv1.TLSProfileSpec{
			Ciphers: []string{
				"TLS_AES_128_GCM_SHA256",
				"TLS_AES_256_GCM_SHA384",
				"TLS_CHACHA20_POLY1305_SHA256",
				"ECDHE-ECDSA-AES128-GCM-SHA256",
				"ECDHE-RSA-AES128-GCM-SHA256",
				"ECDHE-ECDSA-AES256-GCM-SHA384",
				"ECDHE-RSA-AES256-GCM-SHA384",
				"ECDHE-ECDSA-CHACHA20-POLY1305",
				"ECDHE-RSA-CHACHA20-POLY1305",
			},
			MinTLSVersion: configv1.VersionTLS12,
			Groups: []configv1.TLSGroup{
				configv1.TLSGroupX25519,
				configv1.TLSGroupSecP256r1,
				configv1.TLSGroupSecP384r1,
			},
		}

		// CLF output uses a subset of ciphers and curves to prove the
		// custom profile is applied rather than inheriting defaults
		outputProfile := configv1.TLSProfileSpec{
			Ciphers: []string{
				"TLS_AES_128_GCM_SHA256",
				"TLS_AES_256_GCM_SHA384",
				"ECDHE-ECDSA-AES128-GCM-SHA256",
				"ECDHE-RSA-AES128-GCM-SHA256",
			},
			MinTLSVersion: configv1.VersionTLS12,
			Groups: []configv1.TLSGroup{
				configv1.TLSGroupX25519,
				configv1.TLSGroupSecP256r1,
			},
		}

		receiver, err := e2e.DeployHttpReceiverWithTLS(deployNS, receiverProfile)
		Expect(err).To(BeNil(), "failed to deploy HTTP receiver with TLS")

		serviceAccount, err = e2e.BuildAuthorizationFor(deployNS, forwarderName).
			AllowClusterRole(framework.ClusterRoleCollectApplicationLogs).
			Create()
		Expect(err).To(BeNil())

		forwarder := obsruntime.NewClusterLogForwarder(deployNS, forwarderName, runtime.Initialize, func(clf *obs.ClusterLogForwarder) {
			clf.Spec.ServiceAccount.Name = serviceAccount.Name
			clf.Spec.Outputs = []obs.OutputSpec{
				{
					Name: "http-tls-curves",
					Type: obs.OutputTypeHTTP,
					HTTP: &obs.HTTP{
						URLSpec: obs.URLSpec{
							URL: receiver.ClusterLocalEndpoint(),
						},
						Method: "POST",
					},
					TLS: &obs.OutputTLSSpec{
						TLSSpec: obs.TLSSpec{
							CA: &obs.ValueReference{
								Key:        constants.TrustedCABundleKey,
								SecretName: framework.HttpReceiverTLSSecretName,
							},
						},
						TLSSecurityProfile: &configv1.TLSSecurityProfile{
							Type: configv1.TLSProfileCustomType,
							Custom: &configv1.CustomTLSProfile{
								TLSProfileSpec: outputProfile,
							},
						},
					},
				},
			}
			clf.Spec.Pipelines = []obs.PipelineSpec{
				{
					Name:       "app-logs",
					OutputRefs: []string{"http-tls-curves"},
					InputRefs:  []string{string(obs.InputTypeApplication)},
				},
			}
		})

		logGenNS = e2e.CreateTestNamespaceWithPrefix("clo-test-loader")
		if err = e2e.DeployLogGeneratorWithNamespaceName(logGenNS, logGenName, framework.NewDefaultLogGeneratorOptions()); err != nil {
			Fail(fmt.Sprintf("unable to deploy log generator %v.", err))
		}

		if err := e2e.CreateObservabilityClusterLogForwarder(forwarder); err != nil {
			Fail(fmt.Sprintf("Unable to create an instance of logforwarder: %v", err))
		}
		if err := e2e.WaitForDaemonSet(forwarder.Namespace, forwarder.Name); err != nil {
			Fail(err.Error())
		}

		hasLogs, err := receiver.HasApplicationLogs(framework.DefaultWaitForLogsTimeout)
		Expect(err).To(BeNil())
		Expect(hasLogs).To(BeTrue(), "expected to collect application logs via HTTP with TLS curves")
	})
})
