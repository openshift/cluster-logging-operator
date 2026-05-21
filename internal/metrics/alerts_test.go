package metrics

import (
	"bytes"
	"os"
	"path"
	"regexp"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	k8sYAML "k8s.io/apimachinery/pkg/util/yaml"
)

var _ = Describe("CollectorSourceDiscardedLogs alert", Ordered, func() {
	var discardAlert monitoringv1.Rule

	BeforeAll(func() {
		mdir, err := os.Getwd()
		Expect(err).NotTo(HaveOccurred())
		mdir = path.Dir(path.Dir(mdir))
		data, err := os.ReadFile(path.Join(mdir, "config", "prometheus", "collector_alerts.yaml"))
		Expect(err).NotTo(HaveOccurred())

		rule := &monitoringv1.PrometheusRule{}
		err = k8sYAML.NewYAMLOrJSONDecoder(bytes.NewReader(data), 1000).Decode(rule)
		Expect(err).NotTo(HaveOccurred())

		metricRegex := regexp.MustCompile(`(vector_\w+|logcollector_\w+)`)
		for _, group := range rule.Spec.Groups {
			for _, r := range group.Rules {
				if r.Alert == "" {
					continue
				}
				metrics := metricRegex.FindAllString(r.Expr.String(), -1)
				for _, metric := range metrics {
					Expect(collectorMinimalAllowlist.allowedMetrics).To(ContainElement(metric),
						"metric %q used in alert %q is not in the collector minimal allowlist", metric, r.Alert)
				}
				if r.Alert == "CollectorSourceDiscardedLogs" {
					discardAlert = r
				}
			}
		}
		Expect(discardAlert.Alert).NotTo(BeEmpty(), "CollectorSourceDiscardedLogs alert not found in collector_alerts.yaml")
	})

	It("should use discard and error metrics for source components", func() {
		expr := discardAlert.Expr.String()
		Expect(expr).To(ContainSubstring("vector_component_discarded_events_total"))
		Expect(expr).To(ContainSubstring("vector_component_errors_total"))
		Expect(expr).To(ContainSubstring(`component_kind="source"`))
		Expect(expr).To(ContainSubstring("reading_line_from_file"))
		Expect(expr).To(ContainSubstring("reading_line_from_kubernetes_log"))
	})

	It("should group by labels that identify the affected log stream", func() {
		expr := discardAlert.Expr.String()
		Expect(expr).To(ContainSubstring("namespace"))
		Expect(expr).To(ContainSubstring("app_kubernetes_io_instance"))
		Expect(expr).To(ContainSubstring("component_id"))
		Expect(expr).To(ContainSubstring("component_type"))
	})

	It("should have severity warning", func() {
		Expect(discardAlert.Labels["severity"]).To(Equal("warning"))
	})
})
