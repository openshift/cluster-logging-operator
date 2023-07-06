package loki

import (
	"fmt"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
)

var _ = Describe("Loki generator helpers", func() {

	DescribeTable("#formatLokiLabelValue should correctly format", func(label, exp string) {
		Expect(fmt.Sprintf("{{%s}}", exp)).To(Equal(formatLokiLabelValue(label)))
	},
		Entry(" for a complex label", "kubernetes.labels.app.kubernetes.io/name", "kubernetes.labels.app_kubernetes_io_name"),
		Entry(" for a label with only a slash", "kubernetes.labels.foo/bar", "kubernetes.labels.foo_bar"),
		Entry(" for a namespacelabel with only a slash", "kubernetes.namespace_labels.foo/bar", "kubernetes.namespace_labels.foo_bar"),
		Entry(" for a simple label", "kubernetes.labels.foo", "kubernetes.labels.foo"),
		Entry(" for a label without kubernetes.labels prefix", "kubernetes.host", "kubernetes.host"),
	)

})
