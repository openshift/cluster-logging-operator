package api_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	vectorapi "github.com/openshift/cluster-logging-operator/internal/generator/vector/api"
	"sigs.k8s.io/yaml"
)

var _ = Describe("Api", func() {

	It("omits playground when false", func() {
		b, err := yaml.Marshal(vectorapi.Api{Enabled: true, Address: "127.0.0.1:8686"})
		Expect(err).NotTo(HaveOccurred())
		Expect(string(b)).NotTo(ContainSubstring("playground"))
	})
	It("emits playground when true", func() {
		b, err := yaml.Marshal(vectorapi.Api{Enabled: true, Playground: true})
		Expect(err).NotTo(HaveOccurred())
		Expect(string(b)).To(ContainSubstring("playground: true"))
	})

})
