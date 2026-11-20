package transport_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/api/types/transport"
	"sigs.k8s.io/yaml"
)

var _ = Describe("TlsEnabled", func() {

	It("omits enabled when false", func() {
		b, err := yaml.Marshal(transport.TlsEnabled{})
		Expect(err).NotTo(HaveOccurred())
		Expect(string(b)).NotTo(ContainSubstring("enabled"))
	})
	It("emits enabled when true", func() {
		b, err := yaml.Marshal(transport.TlsEnabled{Enabled: true})
		Expect(err).NotTo(HaveOccurred())
		Expect(string(b)).To(ContainSubstring("enabled: true"))
	})

})
