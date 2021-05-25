package certificate_test

import (
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/openshift/cluster-logging-operator/test/helpers/certificate"
	. "github.com/openshift/cluster-logging-operator/test/matchers"
)

var _ = Describe("Certificate", func() {

	It("enables client-server TLS connection with local CA", func() {
		ca := certificate.NewCA(nil, "Root CA") // Self-signed CA
		cert := certificate.NewCert(ca, "Server", "localhost", net.IPv4(127, 0, 0, 1), net.IPv6loopback)
		server := cert.StartServer(http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) { fmt.Fprintln(w, "success!") }), nil)
		defer server.Close()

		client := ca.Client(nil)
		resp, err := client.Get(server.URL)
		ExpectOK(err)
		b, err := ioutil.ReadAll(resp.Body)
		ExpectOK(err)
		Expect(strings.TrimSpace(string(b))).To(Equal("success!"))

		_, err = http.Get(server.URL) // Should fail with plain HTTP
		Expect(err).To(HaveOccurred())
	})

	It("enables TLS mutual authentication", func() {
		ca := certificate.NewCA(nil, "Root CA") // Self-signed CA
		serverCert := certificate.NewCert(ca, "Server")
		server := serverCert.StartServer(http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) { fmt.Fprintln(w, "success!") }), ca)
		defer server.Close()

		clientCert := certificate.NewClient(ca, "Client")
		client := ca.Client(clientCert)
		resp, err := client.Get(server.URL)
		ExpectOK(err)
		b, err := ioutil.ReadAll(resp.Body)
		ExpectOK(err)
		Expect(strings.TrimSpace(string(b))).To(Equal("success!"))

		client = ca.Client(nil) // Should fail without mutual auth.
		_, err = client.Get(server.URL)
		Expect(err).To(HaveOccurred())

		_, err = http.Get(server.URL) // Should fail with plain HTTP
		Expect(err).To(HaveOccurred())

	})
})
