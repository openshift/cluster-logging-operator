package fluentd_test

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/openshift/cluster-logging-operator/test/runtime"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/openshift/cluster-logging-operator/test"
	"github.com/openshift/cluster-logging-operator/test/client"
	"github.com/openshift/cluster-logging-operator/test/helpers/certificate"
	"github.com/openshift/cluster-logging-operator/test/helpers/fluentd"
	. "github.com/openshift/cluster-logging-operator/test/matchers"
)

var _ = Describe("Receiver", func() {
	It("receives and saves data", func() {
		By("creating a receiver")
		t := client.NewTest()
		defer func() { t.Close() }()

		r := fluentd.NewReceiver(t.NS.Name, "receiver")
		r.AddSource(&fluentd.Source{Name: "foo", Type: "forward", Port: 24224})
		r.AddSource(&fluentd.Source{Name: "bar", Type: "http", Port: 24225})
		r.Sources["bar"].CA = certificate.NewCA(nil, "bar")
		r.Sources["bar"].Cert = certificate.NewCert(r.Sources["bar"].CA, r.Host())
		ExpectOK(r.Create(t.Client))

		Expect(r.Sources["foo"].HasOutput()).To(BeFalse())

		var g test.FailGroup
		defer g.Wait()
		msg := `{"hello":"world"}`

		By("sending to a forward source")
		g.Go(func() {
			s := r.Sources["foo"]
			cmd := runtime.Exec(r.Pod, "fluent-cat", "-p", strconv.Itoa(s.Port), "-h", s.Host(), "test.tag")
			cmd.Stdin = strings.NewReader(msg)
			out, err := cmd.CombinedOutput()
			ExpectOK(err, "%v\n%v", cmd.Args, string(out))
		})

		By("sending to a http+TLS source")
		g.Go(func() {
			s := r.Sources["bar"]
			url := fmt.Sprintf("https://%v:%v/test.tag", s.Host(), s.Port)
			cmd := runtime.Exec(r.Pod, "curl", "-kv", "--key", r.ConfigPath("bar-key.pem"), "--cert", r.ConfigPath("bar-cert.pem"), "--cacert", r.ConfigPath("bar-ca.pem"), "-d", "json="+msg, url)
			out, err := cmd.CombinedOutput()
			ExpectOK(err, "%v\n%v", cmd.Args, string(out))
		})

		By("checking for data")
		for _, s := range r.Sources {
			tr := s.TailReader()
			g.Go(func() {
				defer tr.Close()
				line, err := tr.ReadLine()
				ExpectOK(err)
				Expect(strings.TrimSpace(line)).To(Equal(msg))
				Expect(r.Sources["foo"].HasOutput()).To(BeTrue())
			})
		}
	})
})
