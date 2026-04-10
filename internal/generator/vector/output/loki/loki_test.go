package loki

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
)

var _ = Describe("Loki generator helpers", func() {
	defer GinkgoRecover()

	DescribeTable("#formatLokiLabelValue should correctly format", func(label, exp string) {
		Expect(fmt.Sprintf("{{%s}}", exp)).To(Equal(formatLokiLabelValue(label)))
	},
		Entry(" for a complex label", "kubernetes.labels.app.kubernetes.io/name", `kubernetes.labels.\"app_kubernetes_io_name\"`),
		Entry(" for a label with only a slash", "kubernetes.labels.foo/bar", `kubernetes.labels.\"foo_bar\"`),
		Entry(" for a namespacelabel with only a slash", "kubernetes.namespace_labels.foo/bar", `kubernetes.namespace_labels.\"foo_bar\"`),
		Entry(" for a simple label", "kubernetes.labels.foo", `kubernetes.labels.\"foo\"`),
		Entry(" for a label without kubernetes.labels prefix", "kubernetes.host", `kubernetes.host`),
	)

	DescribeTable("#remapLabelKeys", func(labelKeys []string, expected []string) {
		lo := &obs.Loki{LabelKeys: labelKeys}
		Expect(remapLabelKeys(lo)).To(BeEquivalentTo(expected))
	},
		Entry("returns containerLabels when no custom keys",
			nil,
			containerLabels,
		),
		Entry("includes custom direct record field keys",
			[]string{"kubernetes.labels.app"},
			[]string{"kubernetes.container_name", "kubernetes.labels.app", "kubernetes.namespace_name", "kubernetes.pod_name"},
		),
		Entry("excludes keys with custom values (env vars, mapped fields)",
			[]string{"kubernetes.host", "k8s.node_name"},
			[]string{"kubernetes.container_name", "kubernetes.namespace_name", "kubernetes.pod_name"},
		),
		Entry("includes only direct record fields from mixed input",
			[]string{"kubernetes.labels.app", "kubernetes.host", "kubernetes.namespace_name"},
			[]string{"kubernetes.container_name", "kubernetes.labels.app", "kubernetes.namespace_name", "kubernetes.pod_name"},
		),
	)

	DescribeTable("#lokiLabels should correctly format labels", func(label, expKey, expValue string) {
		lo := &obs.Loki{
			LabelKeys: []string{label},
		}
		labels := lokiLabels(lo)
		Expect(labels).To(ContainElement(LokiLabel{
			Name:  expKey,
			Value: expValue,
		}))
	},
		Entry(" for kubernetes.host", "kubernetes.host", `kubernetes_host`, `${VECTOR_SELF_NODE_NAME}`),
		Entry(" for kubernetes.namespace_name", "kubernetes.namespace_name", "kubernetes_namespace_name", `{{kubernetes.namespace_name}}`),
		Entry(" for complex label", `kubernetes.labels.foo-bar-xyz.abc/bar`, "kubernetes_labels_foo_bar_xyz_abc_bar", `{{kubernetes.labels.\"foo-bar-xyz_abc_bar\"}}`),
	)

})
