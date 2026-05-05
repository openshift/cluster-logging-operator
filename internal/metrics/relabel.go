package metrics

import (
	"strings"

	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
)

type metricAllowlistConfig struct {
	allowedMetrics []string
}

type metricDropConfig struct {
	labelName      string
	labelValue     string
	excludeMetrics []string
}

var collectorMinimalAllowlist = &metricAllowlistConfig{
	allowedMetrics: []string{
		// Metrics used in alerts (collector_alerts.yaml)
		"logcollector_component_event_unmatched_count",
		"vector_http_client_errors_total",
		"vector_http_client_requests_sent_total",
		"vector_http_client_responses_total",
		"vector_buffer_byte_size",
		"vector_component_errors_total",
		"vector_component_received_events_total",

		// Metrics used in recording rules (collector_alerts.yaml, telemetry_rules.yaml)
		"vector_component_received_bytes_total",

		// Metrics used in dashboards (openshift-logging-dashboard.json)
		"vector_component_sent_bytes_total",
		"vector_component_received_event_bytes_total",
		"vector_open_files",
		"vector_component_discarded_events_total",

		// Additional buffer and event metrics
		"vector_buffer_discarded_events_total",
		"vector_buffer_events",
		"vector_buffer_sent_events_total",
		"vector_events_in_total",
	},
}

var collectorMinimalDropConfigs = []metricDropConfig{
	{
		labelName:  "component_kind",
		labelValue: "transform",
		excludeMetrics: []string{
			"vector_component_received_bytes_total",
			"vector_component_received_event_bytes_total",
			"vector_component_received_events_total",
			"vector_component_sent_bytes_total",
		},
	},
}

var collectorTelemetryAllowlist = &metricAllowlistConfig{
	allowedMetrics: []string{
		// Used in recording rule (telemetry_rules.yaml)
		"vector_component_received_bytes_total",
	},
}

var lfmeMinimalAllowlist = &metricAllowlistConfig{
	allowedMetrics: []string{
		// Used in recording rule (collector_alerts.yaml) and dashboard
		"log_logged_bytes_total",
	},
}

var lfmeTelemetryAllowlist = &metricAllowlistConfig{
	allowedMetrics: []string{
		// Used in recording rule (collector_alerts.yaml)
		"log_logged_bytes_total",
	},
}

var CollectorMinimalRelabelConfigs = buildRelabelConfigs(collectorMinimalAllowlist, collectorMinimalDropConfigs)
var LFMEMinimalRelabelConfigs = buildRelabelConfigs(lfmeMinimalAllowlist, nil)
var CollectorTelemetryRelabelConfigs = buildRelabelConfigs(collectorTelemetryAllowlist, nil)
var LFMETelemetryRelabelConfigs = buildRelabelConfigs(lfmeTelemetryAllowlist, nil)
var FullRelabelConfigs = buildRelabelConfigs(nil, nil)

func buildRelabelConfigs(allowlist *metricAllowlistConfig, dropConfigs []metricDropConfig) []*monitoringv1.RelabelConfig {
	configs := []*monitoringv1.RelabelConfig{
		{
			SourceLabels: []monitoringv1.LabelName{"__name__"},
			TargetLabel:  "__name__",
			Regex:        "(.*)-(.*)",
			Replacement:  "${1}_${2}",
		},
	}

	if allowlist != nil && len(allowlist.allowedMetrics) > 0 {
		configs = append(configs, &monitoringv1.RelabelConfig{
			Action:       "keep",
			SourceLabels: []monitoringv1.LabelName{"__name__"},
			Regex:        strings.Join(allowlist.allowedMetrics, "|"),
		})
	}

	for _, drop := range dropConfigs {
		configs = append(configs, &monitoringv1.RelabelConfig{
			Action:       "drop",
			SourceLabels: []monitoringv1.LabelName{monitoringv1.LabelName(drop.labelName), "__name__"},
			Regex:        drop.labelValue + ";(" + strings.Join(drop.excludeMetrics, "|") + ")",
		})
	}

	return configs
}
