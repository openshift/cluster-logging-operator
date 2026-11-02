package input_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	"github.com/openshift/cluster-logging-operator/internal/generator/impl/otelc/input/container"
	utilsyaml "github.com/openshift/cluster-logging-operator/internal/utils/yaml"
)

var _ = Describe("Input", func() {
	It("should create a file log receiver for an application input", func() {
		// Create an InputSpec for application logs
		spec := obs.InputSpec{
			Name: "my-application",
			Type: obs.InputTypeApplication,
			Application: &obs.Application{
				Includes: []obs.NamespaceContainerSpec{
					{Namespace: "my-namespace"},
					{Namespace: "another-namespace", Container: "web-server"},
				},
				Excludes: []obs.NamespaceContainerSpec{
					{Namespace: "test-*"},
				},
			},
		}

		// Generate the receivers
		receiver := container.New(spec)
		Expect(receiver).ToNot(BeNil())

		data, err := utilsyaml.Marshal(receiver)
		Expect(err).To(BeNil())

		yamlStr := string(data)
		Expect(yamlStr).To(ContainSubstring("/var/log/pods/my-namespace_*/*/*.log"))
		Expect(yamlStr).To(ContainSubstring("/var/log/pods/another-namespace_*/web-server/*.log"))
		Expect(yamlStr).To(ContainSubstring("/var/log/pods/test-*_*/*/*.log"))
		Expect(yamlStr).To(ContainSubstring("id: parser-crio"))
	})
})
