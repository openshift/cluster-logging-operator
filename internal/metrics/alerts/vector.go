package alerts

const VectorPrometheusAlerts = `
"groups":
- "name": "logging_collector.alerts"
  "rules":
  - "alert": CollectorNodeDown
    "annotations":
      "message": "Prometheus could not scrape {{ $labels.container }} for more than 10m."
      "summary": "Collector cannot be scraped"
    "expr": |
      up{job = "collector", container = "collector"} == 0 or absent(up{job="collector", container="collector"}) == 1
    "for": "10m"
    "labels":
      "service": "collector"
      "severity": "critical"
      namespace: "openshift-logging"
  - alert: CollectorHighErrorRate
    annotations:
      message: |-
        {{ $value }}% of records have resulted in an error by vector {{ $labels.instance }} component.
      summary: "Vector component errors are high"
    expr: |
      100 * (
          sum by(pod, instance)(rate(vector_component_errors_total[2m]))
        /
          sum by(pod, instance)(rate(vector_component_received_events_total[2m]))
        ) > 10
    for: 15m
    labels:
      severity: warning
      namespace: "openshift-logging"
  - alert: CollectorVeryHighErrorRate
    annotations:
      message: |-
        {{ $value }}% of records have resulted in an error by vector {{ $labels.instance }} component.
      summary: "Vector component errors are very high"
    expr: |
      100 * (
          sum by(pod, instance)(rate(vector_component_errors_total[2m]))
        /
          sum by(pod, instance)(rate(vector_component_received_events_total[2m]))
        ) > 25
    for: 15m
    labels:
      severity: critical
      namespace: "openshift-logging"
- "name": "logging_clusterlogging_telemetry.rules"
  "rules":
  - "expr": |
      sum by(cluster)(log_collected_bytes_total)
    "record": "cluster:log_collected_bytes_total:sum"
  - "expr": |
      sum by(cluster)(log_logged_bytes_total)
    "record": "cluster:log_logged_bytes_total:sum"
`
