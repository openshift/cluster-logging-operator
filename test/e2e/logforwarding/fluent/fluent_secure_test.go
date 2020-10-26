package fluent_test

import (
	"fmt"

	. "github.com/onsi/ginkgo"
	"github.com/openshift/cluster-logging-operator/test"
	"github.com/openshift/cluster-logging-operator/test/helpers/certificate"
	. "github.com/openshift/cluster-logging-operator/test/matchers"
	"github.com/openshift/cluster-logging-operator/test/runtime"
)

const secureMessage = "My life is my top secret message."

var _ = Describe("[ClusterLogForwarder]", func() {
	var t *Test
	BeforeEach(func() { t = NewTest(secureMessage) })
	AfterEach(func() { t.Close() })

	XIt("connects to a secure destination", func() {
		const (
			tlsPort = 24225 + iota
		)
		ca := certificate.NewCA(nil) // Self-signed CA cert
		receiverCert := certificate.NewCert(ca, t.receiver.Host())
		s := t.receiver.AddSource("tcp", "forward", tlsPort)
		s.Cert, s.SharedKey = receiverCert, "top-secret"
		t.group.Go(func() { ExpectOK(t.receiver.Create(t.Client)) })

		clf := runtime.NewClusterLogForwarder()
		test.MustUnmarshal(
			fmt.Sprintf(`
outputs:
- name: tls
  type: fluentdForward
  URL:  tls://%v:%v
  secret: {name: mysecret}
pipelines:
- inputRefs: [application]
  outputRefs: [tls]
`, s.Host(), s.Port),
			&clf.Spec)

		clfCert := certificate.NewCert(ca, t.receiver.Host())
		secret := runtime.NewSecret(clf.Namespace, "mysecret", map[string][]byte{
			"tls.crt":       clfCert.CertificatePEM(),
			"tls.key":       clfCert.PrivateKeyPEM(),
			"ca-bundle.crt": ca.CertificatePEM(),
			"ca.key":        ca.PrivateKeyPEM(),
			"shared_key":    []byte("top-secret"),
		})
		t.group.Go(func() {
			ExpectOK(t.Recreate(secret)) // Create secret before CLF.
			ExpectOK(t.Recreate(clf))
		})
		r := t.Reader("tls")
		ExpectOK(r.ExpectLines(1, secureMessage, ""))
	})
})
