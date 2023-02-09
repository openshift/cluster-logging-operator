package fluent_test

import (
	"fmt"
	"strings"

	"github.com/openshift/cluster-logging-operator/internal/runtime"
	"github.com/openshift/cluster-logging-operator/test/framework/functional"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	loggingv1 "github.com/openshift/cluster-logging-operator/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/test"
	"github.com/openshift/cluster-logging-operator/test/client"
	framework "github.com/openshift/cluster-logging-operator/test/framework/e2e"
	"github.com/openshift/cluster-logging-operator/test/helpers/certificate"
	"github.com/openshift/cluster-logging-operator/test/helpers/fluentd"
	. "github.com/openshift/cluster-logging-operator/test/matchers"
	corev1 "k8s.io/api/core/v1"
)

const secureMessage = "My life is my top secret message."

var _ = Describe("[ClusterLogForwarder]", func() {
	const basePort = 24230
	var (
		c                                 *client.Test
		f                                 *Fixture
		privateCA, serverCert, clientCert *certificate.CertKey
		sharedKey                         string
		portOffset                        int
		e2e                               *framework.E2ETestFramework
	)

	BeforeEach(func() {
		e2e = framework.NewE2ETestFramework()
		c = client.NewTest()
		f = NewFixture(c.NS.Name, secureMessage)

		// Receiver acts as TLS server.
		privateCA = certificate.NewCA(nil, "Root CA")
		serverCert = certificate.NewCert(privateCA, "Server", f.Receiver.Host()) // Receiver is server.
		clientCert = certificate.NewCert(privateCA, "Client")
		sharedKey = "top-secret"
		sources := []*fluentd.Source{
			{Name: "no-auth", Type: "forward", Cert: serverCert},
			{Name: "server-auth", Type: "forward", Cert: serverCert},
			{Name: "server-auth-shared", Type: "forward", Cert: serverCert, SharedKey: sharedKey},
			{Name: "mutual-auth", Type: "forward", Cert: serverCert, CA: privateCA},
			{Name: "mutual-auth-shared", Type: "forward", Cert: serverCert, CA: privateCA, SharedKey: sharedKey},
		}
		// The Receiver Sources act as TLS servers.
		for _, s := range sources {
			s.Port = basePort + portOffset
			f.Receiver.AddSource(s)
			portOffset++
		}
		clf := f.ClusterLogForwarder
		secrets := []*corev1.Secret{
			runtime.NewSecret(clf.Namespace, "no-auth", map[string][]byte{}),
			runtime.NewSecret(clf.Namespace, "server-auth", map[string][]byte{
				"ca-bundle.crt": privateCA.CertificatePEM(),
				"ca.key":        privateCA.PrivateKeyPEM(),
			}),
			runtime.NewSecret(clf.Namespace, "server-auth-shared", map[string][]byte{
				"ca-bundle.crt": privateCA.CertificatePEM(),
				"ca.key":        privateCA.PrivateKeyPEM(),
				"shared_key":    []byte(sharedKey),
			}),
			runtime.NewSecret(clf.Namespace, "mutual-auth", map[string][]byte{
				"ca-bundle.crt": privateCA.CertificatePEM(),
				"ca.key":        privateCA.PrivateKeyPEM(),
				"tls.crt":       clientCert.CertificatePEM(),
				"tls.key":       clientCert.PrivateKeyPEM(),
			}),
			runtime.NewSecret(clf.Namespace, "mutual-auth-shared", map[string][]byte{
				"ca-bundle.crt": privateCA.CertificatePEM(),
				"ca.key":        privateCA.PrivateKeyPEM(),
				"tls.crt":       clientCert.CertificatePEM(),
				"tls.key":       clientCert.PrivateKeyPEM(),
				"shared_key":    []byte(sharedKey),
			}),
		}
		g := test.FailGroup{}
		for _, secret := range secrets {
			secret := secret // Don't bind to range variable.
			g.Go(func() { ExpectOK(c.Recreate(secret)) })
		}
		g.Wait() // Create secrets before creating CLF to pass validation first time
	})

	AfterEach(func() {
		c.Close()
		e2e.Cleanup()
	})

	It("connects to secure destinations", func() {
		// Set up the CLF, one output per receiver source.
		clf := f.ClusterLogForwarder
		clf.Spec.Pipelines = []loggingv1.PipelineSpec{{Name: "functional_fluent_pipeline_0_", InputRefs: []string{loggingv1.InputNameApplication}}}
		for _, s := range f.Receiver.Sources {
			secret := &loggingv1.OutputSecretSpec{Name: s.Name}
			var tls *loggingv1.OutputTLSSpec
			if s.Name == "no-auth" {
				secret = nil
				tls = &loggingv1.OutputTLSSpec{
					InsecureSkipVerify: true,
				}
			}
			clf.Spec.Outputs = append(clf.Spec.Outputs, loggingv1.OutputSpec{
				Name:   s.Name,
				Type:   "fluentdForward",
				URL:    fmt.Sprintf("tls://%v:%v", s.Host(), s.Port),
				TLS:    tls,
				Secret: secret,
			})
			clf.Spec.Pipelines[0].OutputRefs = append(clf.Spec.Pipelines[0].OutputRefs, s.Name)
		}
		f.Create(c.Client)
		// Verify log lines at readers.
		g := test.FailGroup{}
		for _, s := range f.Receiver.Sources {
			r := s.TailReader()
			g.Go(func() {
				for i := 0; i < 10; {
					l, err := r.ReadLine()
					ExpectOK(err)
					Expect(l).To(ContainSubstring(`"log_type":"app`)) // Only app logs
					if strings.Contains(l, secureMessage) {
						i++ // Count our own app messages, ignore others.
					}
				}
			})
		}
		g.Wait()
	})

	It("fails to send without permission", func() {
		clf := f.ClusterLogForwarder
		// Secure URL without secret is invalid - need CAr to authenticate server.
		s := f.Receiver.Sources["server-auth"]
		builder := functional.NewClusterLogForwarderBuilder(clf)
		builder.FromInput(loggingv1.InputNameApplication).
			ToOutputWithVisitor(func(spec *loggingv1.OutputSpec) {
				spec.Name = s.Name
				spec.Type = loggingv1.OutputTypeFluentdForward
				spec.URL = fmt.Sprintf("tls://%v:%v", s.Host(), s.Port)
			}, s.Name)

		// shared-key server but no shared-key on Output.
		s = f.Receiver.Sources["server-auth-shared"]
		builder.FromInput(loggingv1.InputNameApplication).
			ToOutputWithVisitor(func(spec *loggingv1.OutputSpec) {
				spec.Name = s.Name
				spec.Type = loggingv1.OutputTypeFluentdForward
				spec.URL = fmt.Sprintf("tls://%v:%v", s.Host(), s.Port)
				spec.Secret = &loggingv1.OutputSecretSpec{Name: "server-auth"}
			}, s.Name)

		// Mutual-auth server but no client certificate.
		s = f.Receiver.Sources["mutual-auth"]
		builder.FromInput(loggingv1.InputNameApplication).
			ToOutputWithVisitor(func(spec *loggingv1.OutputSpec) {
				spec.Name = s.Name
				spec.Type = loggingv1.OutputTypeFluentdForward
				spec.URL = fmt.Sprintf("tls://%v:%v", s.Host(), s.Port)
				spec.Secret = &loggingv1.OutputSecretSpec{Name: "server-auth"}
			}, s.Name)

		f.Create(c.Client)
		for _, s := range f.Receiver.Sources {
			Expect(s.HasOutput()).To(BeFalse(), s.Name)
		}
	})

	It("works when multiple outputs use same Secret", func() {
		clf := f.ClusterLogForwarder
		clf.Spec.Pipelines = []loggingv1.PipelineSpec{{Name: "functional_fluent_pipeline_1_", InputRefs: []string{loggingv1.InputNameApplication}}}
		s := f.Receiver.Sources["server-auth"]
		for i := 0; i < 2; i++ {
			name := fmt.Sprintf("%v%v", s.Name, i)
			clf.Spec.Outputs = append(clf.Spec.Outputs, loggingv1.OutputSpec{
				Name:   name,
				Type:   "fluentdForward",
				URL:    fmt.Sprintf("tls://%v:%v", s.Host(), s.Port),
				Secret: &loggingv1.OutputSecretSpec{Name: s.Name},
			})
			clf.Spec.Pipelines[0].OutputRefs = append(clf.Spec.Pipelines[0].OutputRefs, name)
		}
		f.Create(c.Client)
		By("verify log lines received")
		r := s.TailReader()
		for i := 0; i < 10; {
			l, err := r.ReadLine()
			ExpectOK(err)
			Expect(l).To(ContainSubstring(`"log_type":"app`)) // Only app logs
			if strings.Contains(l, secureMessage) {
				i++ // Count our own app messages, ignore others.
			}
		}
	})
})
