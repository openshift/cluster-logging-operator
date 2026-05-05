package metrics

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
)

var _ = Describe("buildRelabelConfigs", func() {
	It("should return only the rename rule when no allowlist or drops are provided", func() {
		configs := buildRelabelConfigs(nil, nil)
		Expect(configs).To(HaveLen(1))
		Expect(configs[0].TargetLabel).To(Equal("__name__"))
		Expect(configs[0].Regex).To(Equal("(.*)-(.*)"))
		Expect(configs[0].Replacement).To(Equal("${1}_${2}"))
	})

	It("should return rename + keep when an allowlist is provided", func() {
		allowlist := &metricAllowlistConfig{
			allowedMetrics: []string{"metric_a", "metric_b"},
		}
		configs := buildRelabelConfigs(allowlist, nil)
		Expect(configs).To(HaveLen(2))

		Expect(configs[0].Regex).To(Equal("(.*)-(.*)"))

		Expect(string(configs[1].Action)).To(Equal("keep"))
		Expect(configs[1].SourceLabels).To(Equal([]monitoringv1.LabelName{"__name__"}))
		Expect(configs[1].Regex).To(Equal("metric_a|metric_b"))
	})

	It("should return rename + keep + drop when allowlist and drops are provided", func() {
		allowlist := &metricAllowlistConfig{
			allowedMetrics: []string{"metric_a", "metric_b", "metric_c"},
		}
		drops := []metricDropConfig{
			{
				labelName:      "component_kind",
				labelValue:     "transform",
				excludeMetrics: []string{"metric_a", "metric_b"},
			},
		}
		configs := buildRelabelConfigs(allowlist, drops)
		Expect(configs).To(HaveLen(3))

		Expect(string(configs[1].Action)).To(Equal("keep"))
		Expect(configs[1].Regex).To(Equal("metric_a|metric_b|metric_c"))

		Expect(string(configs[2].Action)).To(Equal("drop"))
		Expect(configs[2].SourceLabels).To(Equal([]monitoringv1.LabelName{"component_kind", "__name__"}))
		Expect(configs[2].Regex).To(Equal("transform;(metric_a|metric_b)"))
	})

	It("should build valid CollectorMinimalRelabelConfigs", func() {
		Expect(CollectorMinimalRelabelConfigs).To(HaveLen(3))
		Expect(string(CollectorMinimalRelabelConfigs[1].Action)).To(Equal("keep"))
		Expect(string(CollectorMinimalRelabelConfigs[2].Action)).To(Equal("drop"))

		keepRegex := CollectorMinimalRelabelConfigs[1].Regex
		for _, m := range collectorMinimalAllowlist.allowedMetrics {
			Expect(keepRegex).To(ContainSubstring(m), "missing metric in keep regex: %s", m)
		}
	})

	It("should build valid LFMEMinimalRelabelConfigs", func() {
		Expect(LFMEMinimalRelabelConfigs).To(HaveLen(2))
		Expect(string(LFMEMinimalRelabelConfigs[1].Action)).To(Equal("keep"))
		Expect(LFMEMinimalRelabelConfigs[1].Regex).To(Equal("log_logged_bytes_total"))
	})

	It("should build valid CollectorTelemetryRelabelConfigs", func() {
		Expect(CollectorTelemetryRelabelConfigs).To(HaveLen(2))
		Expect(string(CollectorTelemetryRelabelConfigs[1].Action)).To(Equal("keep"))
		Expect(CollectorTelemetryRelabelConfigs[1].Regex).To(Equal("vector_component_received_bytes_total"))
	})

	It("should build valid LFMETelemetryRelabelConfigs", func() {
		Expect(LFMETelemetryRelabelConfigs).To(HaveLen(2))
		Expect(string(LFMETelemetryRelabelConfigs[1].Action)).To(Equal("keep"))
		Expect(LFMETelemetryRelabelConfigs[1].Regex).To(Equal("log_logged_bytes_total"))
	})

	It("should build FullRelabelConfigs with only the rename rule", func() {
		Expect(FullRelabelConfigs).To(HaveLen(1))
		Expect(FullRelabelConfigs[0].Regex).To(Equal("(.*)-(.*)"))
	})

	It("should return only the rename rule when allowlist has empty metrics", func() {
		allowlist := &metricAllowlistConfig{
			allowedMetrics: []string{},
		}
		configs := buildRelabelConfigs(allowlist, nil)
		Expect(configs).To(HaveLen(1))
		Expect(configs[0].TargetLabel).To(Equal("__name__"))
	})

	It("should return rename + drop when only drop configs are provided", func() {
		drops := []metricDropConfig{
			{
				labelName:      "component_kind",
				labelValue:     "transform",
				excludeMetrics: []string{"metric_a"},
			},
		}
		configs := buildRelabelConfigs(nil, drops)
		Expect(configs).To(HaveLen(2))

		Expect(configs[0].Regex).To(Equal("(.*)-(.*)"))
		Expect(string(configs[1].Action)).To(Equal("drop"))
		Expect(configs[1].SourceLabels).To(Equal([]monitoringv1.LabelName{"component_kind", "__name__"}))
		Expect(configs[1].Regex).To(Equal("transform;(metric_a)"))
	})
})
