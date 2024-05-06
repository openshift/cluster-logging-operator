package loki

import (
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
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

	DescribeTable("#lokiLabels should correctly format labels", func(label, expKey, expValue string) {
		lo := &obs.Loki{
			LabelKeys: []string{label},
		}
		labels := lokiLabels(lo)
		Expect(labels).To(ContainElement(Label{
			Name:  expKey,
			Value: expValue,
		}))
	},
		Entry(" for kubernetes.host", "kubernetes.host", `kubernetes_host`, `${VECTOR_SELF_NODE_NAME}`),
		Entry(" for kubernetes.namespace_name", "kubernetes.namespace_name", "kubernetes_namespace_name", `{{kubernetes.namespace_name}}`),
		Entry(" for complex label", `kubernetes.labels.foo-bar-xyz.abc/bar`, "kubernetes_labels_foo_bar_xyz_abc_bar", `{{kubernetes.labels.\"foo-bar-xyz_abc_bar\"}}`),
	)

})
