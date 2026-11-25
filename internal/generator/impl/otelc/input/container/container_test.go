package container_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/openshift/cluster-logging-operator/internal/generator/impl/otelc/api/receivers/operators"
	"github.com/openshift/cluster-logging-operator/internal/generator/impl/otelc/input/container"
	utilsyaml "github.com/openshift/cluster-logging-operator/internal/utils/yaml"
)

var _ = Describe("Container Input Operators", func() {
	Context("NewOperators", func() {
		It("should create a complete operator pipeline", func() {
			ops := container.NewOperators(
				"cluster-123",
				"",
				"container",
				"application")

			// Verify we have all expected operators
			Expect(ops).To(HaveLen(15))

			// Verify operator IDs in order
			expectedIDs := []string{
				"parser-crio",
				"extract_metadata_from_filepath",
				"move_log_to_body",
				"set_iostream",
				"move_namespace_to_resource",
				"move_pod_name_to_resource",
				"move_pod_uid_to_resource",
				"move_container_name_to_resource",
				"remove_restartcount",
				"remove_logtag",
				"remove_time",
				"add_node_name",
				"add_cluster_uid",
				"add_log_source",
				"add_log_type",
			}

			for i, expectedID := range expectedIDs {
				Expect(ops[i].ID).To(Equal(expectedID), "Operator %d should have ID %s", i, expectedID)
			}
		})

		It("should marshal to YAML correctly", func() {
			ops := container.NewOperators("cluster-123",
				"",
				"container",
				"application")

			data, err := utilsyaml.Marshal(ops)
			Expect(err).To(BeNil())

			yamlStr := string(data)

			// Verify key components are present
			Expect(yamlStr).To(ContainSubstring("type: regex_parser"))
			Expect(yamlStr).To(ContainSubstring("id: parser-crio"))
			Expect(yamlStr).To(ContainSubstring("type: move"))
			Expect(yamlStr).To(ContainSubstring("type: remove"))
			Expect(yamlStr).To(ContainSubstring("k8s.namespace.name"))
			Expect(yamlStr).To(ContainSubstring("k8s.pod.name"))
			Expect(yamlStr).To(ContainSubstring("k8s.container.name"))
			Expect(yamlStr).To(ContainSubstring("log.iostream"))
		})

		It("should create proper CRI-O regex pattern", func() {
			ops := container.NewOperators("cluster-123",
				"",
				"container",
				"application")

			// First operator should be the CRI-O parser
			Expect(ops[0].Type).To(Equal(operators.OperatorTypeRegexParser))
			Expect(ops[0].ID).To(Equal("parser-crio"))

			config := ops[0].Config
			Expect(config).To(HaveKey("regex"))
			regex := config["regex"].(string)

			// Verify regex captures time, stream, logtag, and log
			Expect(regex).To(ContainSubstring("(?P<time>"))
			Expect(regex).To(ContainSubstring("(?P<stream>stdout|stderr)"))
			Expect(regex).To(ContainSubstring("(?P<logtag>"))
			Expect(regex).To(ContainSubstring("(?P<log>"))
		})
	})

})
