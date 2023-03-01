//go:build vector
// +build vector

package outputs

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	"strings"

	openshiftv1 "github.com/openshift/api/config/v1"
	logging "github.com/openshift/cluster-logging-operator/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	"github.com/openshift/cluster-logging-operator/internal/runtime"
	"github.com/openshift/cluster-logging-operator/test/framework/functional"
	testfw "github.com/openshift/cluster-logging-operator/test/functional"
	"github.com/openshift/cluster-logging-operator/test/helpers/certificate"
	corev1 "k8s.io/api/core/v1"
	"net"
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
		framework = functional.NewCollectorFunctionalFrameworkUsingCollector(testfw.LogCollectionType)
	})

	AfterEach(func() {
		framework.Cleanup()
	})
	DescribeTable("with an output TLS Security Policy should send message",
		func(outputType string, profile openshiftv1.TLSProfileType, addDestinationContainer func(f *functional.CollectorFunctionalFramework, secret *corev1.Secret) runtime.PodBuilderVisitor) {
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
			functional.NewClusterLogForwarderBuilder(framework.Forwarder).
				FromInput(logging.InputNameApplication).
				ToOutputWithVisitor(func(spec *logging.OutputSpec) {
					spec.URL = strings.Replace(spec.URL, "http://", "https://", 1)
					spec.Secret = &logging.OutputSecretSpec{
						Name: clientSecretName,
					}
					spec.TLS = &logging.OutputTLSSpec{
						TLSSecurityProfile: &openshiftv1.TLSSecurityProfile{
							Type: profile,
						},
					}
				}, outputType)

			Expect(framework.DeployWithVisitor(addDestinationContainer(framework, serverSecret))).To(BeNil())

			Expect(framework.WritesApplicationLogs(10)).To(BeNil())

			Expect(framework.ReadRawApplicationLogsFrom(outputType)).ToNot(BeEmpty())

		},
		Entry("to an HTTP output with matching profiles", logging.OutputTypeHttp, openshiftv1.TLSProfileIntermediateType, func(f *functional.CollectorFunctionalFramework, secret *corev1.Secret) runtime.PodBuilderVisitor {
			return func(b *runtime.PodBuilder) error {
				return f.AddVectorHttpOutputWithConfig(b, f.Forwarder.Spec.Outputs[0], openshiftv1.TLSProfileIntermediateType, secret)
			}
		}),
		Entry("to an HTTP output with different profiles", logging.OutputTypeHttp, openshiftv1.TLSProfileIntermediateType, func(f *functional.CollectorFunctionalFramework, secret *corev1.Secret) runtime.PodBuilderVisitor {
			return func(b *runtime.PodBuilder) error {
				return f.AddVectorHttpOutputWithConfig(b, f.Forwarder.Spec.Outputs[0], openshiftv1.TLSProfileOldType, secret)
			}
		}),
	)

})
