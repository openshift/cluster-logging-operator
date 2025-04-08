package outputs

import (
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	testruntime "github.com/openshift/cluster-logging-operator/test/runtime/observability"

	"net"

	openshiftv1 "github.com/openshift/api/config/v1"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	"github.com/openshift/cluster-logging-operator/internal/runtime"
	"github.com/openshift/cluster-logging-operator/test/framework/functional"
	"github.com/openshift/cluster-logging-operator/test/helpers/certificate"
	corev1 "k8s.io/api/core/v1"
)

var _ = Describe("[functional][outputs][tlssecurityprofile] Functional tests ", func() {
	var (
		framework *functional.CollectorFunctionalFramework
	)

	const (
		serverSecretName = "servercerts"
		clientSecretName = "clientcerts"
	)

	BeforeEach(func() {
		framework = functional.NewCollectorFunctionalFramework()
	})

	AfterEach(func() {
		framework.Cleanup()
	})
	DescribeTable("with an output TLS Security Policy should send message",
		func(profile openshiftv1.TLSProfileType, addDestinationContainer func(f *functional.CollectorFunctionalFramework, secret *corev1.Secret) runtime.PodBuilderVisitor) {
			ca := certificate.NewCA(nil, "Root CA") // Self-signed CA
			serverCert := certificate.NewCert(ca, "widgits.org", "localhost", net.IPv4(127, 0, 0, 1), net.IPv6loopback)
			serverSecret := runtime.NewSecret("", serverSecretName, map[string][]byte{
				constants.ClientPrivateKey: serverCert.PrivateKeyPEM(),
				constants.ClientCertKey:    serverCert.CertificatePEM(),
			})
			framework.AddSecret(serverSecret)
			framework.AddSecret(
				runtime.NewSecret("", clientSecretName, map[string][]byte{
					constants.TrustedCABundleKey: ca.CertificatePEM(),
				}))
			testruntime.NewClusterLogForwarderBuilder(framework.Forwarder).
				FromInput(obs.InputTypeApplication).
				ToHttpOutput(func(output *obs.OutputSpec) {
					output.HTTP.URL = strings.Replace(output.HTTP.URL, "http://", "https://", 1)
					output.TLS = &obs.OutputTLSSpec{
						TLSSpec: obs.TLSSpec{
							CA: &obs.ValueReference{
								Key:        constants.TrustedCABundleKey,
								SecretName: clientSecretName,
							},
						},
						TLSSecurityProfile: &openshiftv1.TLSSecurityProfile{
							Type: profile,
						},
					}
				})

			Expect(framework.DeployWithVisitor(addDestinationContainer(framework, serverSecret))).To(BeNil())

			Expect(framework.WritesApplicationLogs(10)).To(BeNil())

			Expect(framework.ReadRawApplicationLogsFrom(string(obs.OutputTypeHTTP))).ToNot(BeEmpty())

		},
		Entry("to an HTTP output with matching profiles", openshiftv1.TLSProfileIntermediateType, func(f *functional.CollectorFunctionalFramework, secret *corev1.Secret) runtime.PodBuilderVisitor {
			return func(b *runtime.PodBuilder) error {
				return f.AddVectorHttpOutputWithConfig(b, f.Forwarder.Spec.Outputs[0], openshiftv1.TLSProfileIntermediateType, secret)
			}
		}),
		Entry("to an HTTP output with different profiles", openshiftv1.TLSProfileIntermediateType, func(f *functional.CollectorFunctionalFramework, secret *corev1.Secret) runtime.PodBuilderVisitor {
			return func(b *runtime.PodBuilder) error {
				return f.AddVectorHttpOutputWithConfig(b, f.Forwarder.Spec.Outputs[0], openshiftv1.TLSProfileOldType, secret)
			}
		}),
	)

})
