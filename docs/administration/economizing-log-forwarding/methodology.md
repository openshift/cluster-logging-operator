# Economizing Log Forwarding — Methodology

This page documents the full test methodology behind the data in the
[Economizing Log Forwarding](economizing-log-forwarding.md) guide.

## Test Environment

### Cluster

- 6-node OpenShift cluster on GCP (dev cluster)
- Default platform workloads (no customer applications beyond synthetic load)

### Outputs Deployed

Three output types were deployed simultaneously to capture
serialization-specific differences:

| Output | Purpose |
|--------|---------|
| **HTTP receiver** | Lightweight pod accepting JSON logs to disk. Output-agnostic byte baseline. |
| **LokiStack** | 1x.small instance backed by in-cluster MinIO. Tests label-sensitive routing. |
| **Google Cloud Logging** | Cloud output serialization. Requires `.hostname` in all tiers. |

Per-output recommendations for other output types (Elasticsearch, Splunk,
Kafka, Syslog, S3, OTLP, Azure Monitor) are derived from their documented
field requirements and source code, not live measurement.

### Synthetic Workloads

Log traffic was generated using the
`quay.io/openshift-logging/cluster-logging-load-client:0.2` image from the
CLO functional benchmarker:

- 3+ namespaces with ~5 pods each
- ~100 log lines/second per pod
- Pods annotated with realistic metadata: Argo/Flux deployment labels,
  Prometheus scrape configs, and long annotation values simulating
  real-world bloat
- Pods labeled with 10–20 labels including dots/slashes (triggers ViaQ
  dedotting)

## Log Type Groups

| Group | log_type | log_source | Notes |
|-------|----------|------------|-------|
| Container | application + infrastructure | container | Same structure, only `log_type` differs |
| Journal | infrastructure | node | systemd journal logs |
| K8s/OCP API Audit | audit | kubeAPI + openshiftAPI | Same schema, grouped |
| Host/OVN Audit | audit | auditd + ovn | Simpler message-based formats |

## Measurement Approach

### Metrics Collected

| Metric | Source | Purpose |
|--------|--------|---------|
| `vector_component_sent_bytes_total` | Collector `/metrics` endpoint | Total bytes forwarded per output component |
| Raw JSON entries | HTTP receiver disk output | Per-entry size and per-field byte cost |
| CPU / Memory | `oc adm top pods` (sampled every 30s) | Collector resource overhead |

### Test Run Parameters

- **Warmup:** 2 minutes after CLF apply (data discarded)
- **Steady-state measurement:** 10 minutes
- **CPU/memory sampling interval:** 30 seconds
- **Entry sample size:** 100 entries per log group per tier (for per-entry
  size analysis); 500 entries per log group for baseline field cost analysis

### Per-Tier Procedure

Each tier was measured independently to avoid cross-contamination:

1. Apply the tier's CLF spec
2. Wait for the collector DaemonSet rollout to complete
3. Validate the CLF status condition (`observability.openshift.io/Valid`)
4. Scrape `vector_component_sent_bytes_total` (start snapshot)
5. Wait 2 minutes (warmup — discard)
6. Scrape `vector_component_sent_bytes_total` (measurement start)
7. Begin CPU/memory sampling every 30 seconds
8. Wait 10 minutes (steady-state collection)
9. Scrape `vector_component_sent_bytes_total` (measurement end)
10. Capture 100-entry JSON samples per log group from the HTTP receiver
11. Compute byte deltas: `end - start` per output component, excluding
    internal `prometheus_output`

### Calculations

- **Per-entry size:** Each sampled JSON entry is re-serialized to compact
  JSON (`json.dumps(e, separators=(",", ":"))`) and its byte length
  measured. The average across 100 entries is reported.
- **Per-field byte cost:** From 500-entry baseline samples, each field's
  value is serialized independently and its byte length measured. Averages
  are computed per field, along with presence percentage across entries.
- **Total bytes forwarded:** Sum of `vector_component_sent_bytes_total` deltas
  across all collector pods, excluding internal `prometheus_output`.
- **Savings percentage:** `(baseline - tier) / baseline × 100`
- **Bloat ratio:** `sum(entry_bytes) / sum(message_field_bytes)` across
  the tier's 100-entry sample.
- **Resource averages:** Mean of CPU (millicores) and memory (MiB) across
  all collector pods and all sample points in the 10-minute window.

## Scripts and Manifests

All test scripts, CLF manifests, and analysis tools used in this
investigation are available in the
[LOG-9673 branch](https://github.com/Clee2691/clo_investigations/tree/LOG-9673)
of the clo_investigations repository.

## Caveats

- **Single-cluster measurement.** Results reflect one cluster's workload
  mix. Clusters with heavier annotations, more audit traffic, or different
  log volume ratios will see different absolute numbers. The relative
  savings per log group should be broadly similar.
- **Synthetic workload.** The log stressor generates realistic metadata
  but uniform message content. Real workloads with longer or shorter
  messages will shift the bloat ratio.
- **Output-specific overhead.** Different outputs serialize entries
  differently. The HTTP receiver captures raw JSON; GCL and LokiStack add
  their own framing. Total savings vary by output.
