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
        {{ $value }}% of records have resulted in an error by collector {{ $labels.instance }} component.
      summary: "Collector component errors are high"
    expr: |
         sum by(pod, instance)(rate(vector_component_errors_total[2m])) > 0.05
    for: 5m
    labels:
      severity: warning
      namespace: "openshift-logging"
  - alert: CollectorSyncFailed
    annotations:
      message: |-
        Collector {{ $labels.instance }} log syncing failed.
      summary: "Collector component errors"
    expr: |
         (sum by (pod, instance, component_name) (rate(vector_events_out_total{component_kind="sink"}[2m]))) == 0
    for: 5m
    labels:
      severity: critical
      namespace: "openshift-logging"
  - alert: CollectorVeryHighErrorRate
    annotations:
      message: |-
        {{ $value }}% of records have resulted in an error by collector {{ $labels.instance }} component.
      summary: "Collector component errors are very high"
    expr: |
      sum by(pod, instance)(rate(vector_component_errors_total[2m])) > 0.2
    for: 5m
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
