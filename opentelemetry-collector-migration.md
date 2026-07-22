# OpenTelemetry collector/operator migration

This document outlines the migration steps of CLO to the OpenTelemetry collector and operator.

The migration will be done in milestones:
1. Use OpenTelemetry collector in CLO instead of the current vector collector
2. Use OpenTelemetry collector CR in CLO instead of the current vector deployment.
3. Brainstorm the migration path of CLO to pure OpenTelemetry operator.

---

## Milestone 1: Replace Vector with OpenTelemetry Collector

**Goal:** CLO continues to manage the collector deployment and reconcile the `ClusterLogForwarder` CR, but generates an OpenTelemetry Collector YAML configuration instead of a Vector TOML configuration. The deployed pod runs the OTEL collector binary instead of Vector.

The `ClusterLogForwarder` CRD API remains unchanged.

### Architecture Change

```
Current:  ClusterLogForwarder CR → CLO → Vector TOML config → Vector DaemonSet
Target:   ClusterLogForwarder CR → CLO → OTEL Collector YAML config → OTEL Collector DaemonSet
```

### Feature Compatibility Analysis

#### 1. Inputs (Sources) → OTEL Receivers

| CLO Input | Vector Source | OTEL Receiver | Status | Notes |
|-----------|-------------|---------------|--------|-------|
| `application` (container logs) | `kubernetes_logs` | `filelog` + `k8sattributes` processor | ✅ Supported | `filelog` reads from `/var/log/pods/` or `/var/log/containers/`. `k8sattributes` enriches with pod name, namespace, labels, annotations, UID, node name. Glob include/exclude patterns for namespace filtering are supported. Partial log merge (Docker JSON) handled by `filelog`'s multiline config. |
| `infrastructure` (container logs) | `kubernetes_logs` (filtered) | `filelog` + `k8sattributes` processor | ✅ Supported | Same as application but filtered to `openshift-*`, `kube-*`, `default` namespaces via include glob patterns. |
| `infrastructure` (node/journal) | `journald` | `journalctl` receiver | ✅ Supported | OTEL contrib `journalctl` receiver reads systemd journal. Supports directory config, unit filtering. |
| `audit` (auditd, kube-api, openshift-api, ovn) | `file` (multiple paths) | `filelog` (multiple instances) | ✅ Supported | `filelog` receiver watches specific file paths (`/var/log/audit/audit.log`, `/var/log/kube-apiserver/audit.log`, etc.). Supports `start_at`, `max_log_size`, glob patterns. |
| `receiver` (HTTP) | `http_server` (JSON decoding) | `httpreceiver` (contrib) | ⚠️ Needs verification | CLO's HTTP receiver accepts arbitrary JSON (specifically `kubeAPIAudit` format). The OTEL `httpreceiver` (contrib) or a webhook receiver would work. Alternatively, could use `otlpreceiver` if clients send OTLP. |
| `receiver` (Syslog) | `syslog` (TCP) | `syslog` receiver | ✅ Supported | OTEL contrib `syslog` receiver supports TCP/UDP, RFC3164/5424, TLS. |

**Input-level tuning:**

| Feature | Vector | OTEL Equivalent | Status |
|---------|--------|-----------------|--------|
| `MaxRecordsPerSecond` (per-container throttle) | `throttle` transform with `key_field: {{ _internal.file }}` | Elastic's [`ratelimitprocessor`](https://pkg.go.dev/github.com/elastic/opentelemetry-collector-components/processor/ratelimitprocessor); [contrib #35204](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/35204) accepted | ⚠️ Partial — not in official contrib yet |
| `MaxMessageSize` | `max_line_bytes` / `max_read_bytes` | `filelog` receiver `max_log_size` | ✅ Supported |
| `IgnoreOlder` (audit files) | `ignore_older_secs` on file source | `filelog` receiver `start_at: end` + poll interval | ⚠️ Partial |

#### 2. Filters → OTEL Processors

| CLO Filter | Vector Transform | OTEL Processor | Status | Notes |
|------------|-----------------|----------------|--------|-------|
| `drop` | `remap` with VRL conditions + `abort` | `filter` processor with OTTL | ✅ Supported | OTTL supports regex matching (`IsMatch`), boolean AND/OR logic, field access. The `filter` processor can drop log records matching conditions. |
| `prune` (field removal) | `remap` with VRL `del()` / field iteration | `transform` processor with OTTL `delete_key` / `keep_keys` | ✅ Supported | OTTL `delete_key`, `delete_matching_keys`, `keep_matching_keys` functions handle both `in` and `notIn` modes. |
| `parse` (JSON parsing) | `remap` with `parse_json()` | `transform` processor with OTTL `ParseJSON` or `json_parser` operator in `filelog` | ✅ Supported | Can parse JSON message body into structured attributes. |
| `openshiftLabels` | `remap` setting `._internal.openshift.labels` | `transform` or `attributes` processor | ✅ Supported | Set resource/log attributes from static values. |
| `detectMultilineException` | `detect_exceptions` transform | No built-in equivalent | ❌ Gap | Vector's `detect_exceptions` transform detects and merges multi-line stack traces (Java, Python, Go, Ruby, etc.) after collection. OTEL's `filelog` receiver has `multiline` config at ingestion, but post-ingestion exception detection across grouped streams is not available as a processor. A custom processor or the `groupby` processor with heuristics would be needed. |
| `kubeAPIAudit` | Complex VRL with policy rules, wildcards, verb matching, field redaction | `transform` processor with OTTL | ⚠️ Complex | The Kube API audit filter implements K8s audit policy logic: rule evaluation with wildcards in users/groups/namespaces/resources, verb matching, `OmitStages`, `OmitResponseCodes`, severity level setting (`None`/`Metadata`/`Request`/`RequestResponse`), field redaction (`requestObject`/`responseObject`). OTTL can handle conditions and field deletion, but the wildcard matching and rule precedence logic is complex. May need a custom processor or very extensive OTTL statements. |

#### 3. Internal Normalization (VIAQ Data Model) → OTEL Processors

CLO applies extensive normalization to all logs before forwarding. This is currently implemented as VRL transforms.

| Feature | Vector VRL | OTEL Equivalent | Status | Notes |
|---------|-----------|-----------------|--------|-------|
| Envelope wrapping (`_internal`) | Wrap all fields in `._internal` | Not needed | ✅ N/A | OTEL's data model (resource attributes, log body, log attributes) provides natural separation. No envelope hack needed. |
| Log type classification | VRL namespace regex matching | `routing` processor or `transform` with OTTL | ✅ Supported | Route by namespace pattern to classify as `application` vs `infrastructure`. |
| Log source tagging | Set `.log_source` based on input | Resource attributes set by receiver pipeline | ✅ Supported | Each receiver pipeline naturally tags the source. |
| Cluster ID injection | `${OPENSHIFT_CLUSTER_ID}` env var | `resource` processor or `transform` | ✅ Supported | Add `openshift.cluster.uid` resource attribute from env var. |
| Hostname injection | `$VECTOR_SELF_NODE_NAME` env var | `resourcedetection` processor or `resource` processor | ✅ Supported | `resourcedetection` processor with `env` detector, or `k8snode` detector. |
| Sequence number | `to_unix_timestamp(now(), nanoseconds)` | Not directly equivalent | ⚠️ Minor | Could use observed timestamp or a custom processor. Sequence is used for ordering. |
| **Log level detection** (5-stage) | logfmt → klog → grok → pattern match → regex | `transform` processor with OTTL | ⚠️ Feasible | OTTL has [`ExtractGrokPatterns`](https://github.com/open-telemetry/opentelemetry-collector-contrib/blob/main/pkg/ottl/ottlfuncs/README.md) (uses Elastic Go-Grok library) which covers stage 3 (grok patterns for Logback, Log4j). Combined with OTTL regex (`IsMatch`) for stages 4-5, most detection is feasible. Stages 1-2 (logfmt, klog) have no dedicated OTTL parser but could be approximated with grok/regex patterns. Overall: achievable with OTTL, though the config will be verbose. |
| Dedot labels | Replace `.` and `/` with `_` in label keys | `transform` processor with OTTL `replace_all_patterns` | ✅ Supported | OTTL string functions can handle key normalization. |
| EventRouter log handling | Parse eventrouter JSON, extract event fields, fix timestamps | `transform` processor with OTTL | ⚠️ Complex | Detecting eventrouter pods, parsing nested JSON event structure, timestamp priority logic (lastTimestamp > firstTimestamp > eventTime > creationTimestamp). Doable with OTTL but complex. |
| Container I/O stream | Extract `stream` → `kubernetes.container_iostream` | `filelog` receiver attributes or `transform` | ✅ Supported | |

**Key insight:** Many of the VIAQ normalization transforms exist to reshape Vector's internal data model into a standard format. With the OTEL collector, logs are natively in the OTLP data model (resource attributes, scope, log record with body/attributes/severity/timestamp/trace context). This eliminates much of the envelope/reshape logic. However, for non-OTLP outputs (Elasticsearch, Splunk, Syslog, etc.), the OTEL exporters must produce output compatible with what Vector currently emits (the VIAQ format), or consumers must accept the OTLP-based format.

#### 4. Outputs (Sinks) → OTEL Exporters

| CLO Output | Vector Sink | OTEL Exporter | Status | Notes |
|------------|------------|---------------|--------|-------|
| `otlp` | `opentelemetry` (HTTP) | `otlp` / `otlphttp` exporter | ✅ Native | This is the OTEL collector's primary export path. Supports gRPC and HTTP. Compression, auth, TLS, batch, retry all built-in. Currently CLO only uses HTTP; OTEL supports both. |
| `elasticsearch` | `elasticsearch` | `elasticsearch` exporter (contrib) | ✅ Supported | Supports bulk API, auth (basic, bearer). Dynamic index routing is the default since [v0.122.0](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/exporter/elasticsearchexporter): set `elasticsearch.index` attribute on log records via `transform` processor to route per-record. CLO's index templates map directly. ES 6 support needs verification (may be dropped). |
| `kafka` | `kafka` | `kafka` exporter (contrib) | ✅ Supported | SASL auth (PLAIN, SCRAM-SHA-256/512), TLS, compression (gzip, snappy, lz4, zstd), topic routing, multiple brokers. |
| `loki` | `loki` | `loki` exporter (contrib) | ✅ Supported | Labels, tenant ID, auth (basic, bearer), compression, out-of-order handling. Label key mapping from OTLP attributes is native. |
| `lokiStack` | Multiple `loki` or `otlp` sinks (per tenant) | `loki` exporter with `routing` processor | ✅ Supported | Route by log type (application/infrastructure/audit) to tenant-specific endpoints. LokiStack OTLP data model (`Otel`) maps directly. |
| `splunk` | `splunk_hec_logs` | `splunk_hec` exporter (contrib) | ✅ Supported | HEC token, index, source, sourcetype, host. Supports event format. The complex VRL for source/sourcetype detection would move to `transform` processor + exporter config. |
| `http` (generic JSON/NDJSON) | `http` | No generic HTTP log exporter | ❌ Gap | Vector's HTTP sink sends arbitrary JSON/NDJSON to any HTTP endpoint with configurable method, headers, auth. The OTEL collector has `otlphttp` (OTLP format only) but no generic HTTP JSON exporter. A [JSON log exporter was proposed](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/10836) but not implemented. **Needs a custom exporter.** |
| `syslog` | `socket` (TCP/UDP/TLS) | `syslog` exporter (contrib) | ✅ Supported | RFC3164/5424, TCP/UDP/TLS. Facility, severity, appname, procid, msgid mapping. The complex VRL transforms for syslog field construction would move to `transform` processor config. |
| `cloudwatch` | `aws_cloudwatch_logs` | `awscloudwatchlogs` exporter (contrib) | ✅ Supported | Region, log group (templated), log stream, AWS auth (access key, IAM role, STS). Batch limits (1MB CloudWatch API cap). |
| `s3` | `aws_s3` | `awss3` exporter (contrib) | ✅ Supported | Region, bucket, key prefix, custom endpoint for S3-compatible stores, AWS auth, compression. |
| `googleCloudLogging` | `gcp_stackdriver_logs` | `googlecloud` exporter (contrib) | ✅ Supported | Project/folder/org/billing account ID, log ID, severity mapping, GCP credentials. |
| `azureMonitor` (DEPRECATED) | `azure_monitor_logs` | `azuremonitor` exporter (contrib) | ⚠️ Deprecated | Since this output is deprecated in CLO, lower priority. OTEL contrib has an Azure Monitor exporter but it may target Application Insights rather than Log Analytics. |
| `azureLogsIngestion` | `azure_logs_ingestion` (custom Vector sink) | No dedicated exporter | ❌ Gap | Azure Logs Ingestion API (Data Collection Rules) — no OTEL contrib exporter exists. [Issue #40478](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/40478) is open requesting DCR-based exporter. Needs a custom exporter. Workload identity and client secret auth are required. |

#### 5. Cross-Cutting Concerns

| Feature | Vector | OTEL Collector | Status | Notes |
|---------|--------|---------------|--------|-------|
| **TLS** (CA, cert, key, passphrase) | Per-source/sink TLS config | Per-receiver/exporter TLS config | ✅ Supported | OTEL supports `ca_file`, `cert_file`, `key_file`, min TLS version, cipher suites. Key passphrase support needs verification. |
| **OpenShift TLS Security Profiles** | Mapped to Vector TLS min version + ciphers | Map to OTEL TLS `min_version` + `cipher_suites` | ✅ Supported | Same mapping logic, different config format. |
| **Authentication (Bearer token)** | HTTP auth header or file-based token | `bearertokenauth` extension | ✅ Supported | OTEL's auth extensions support token from file (service account). |
| **Authentication (Basic auth)** | Username/password in sink config | `basicauth` extension | ✅ Supported | |
| **Authentication (AWS)** | AWS access key or IAM role (STS) | Built into AWS exporters + `sigv4auth` extension | ✅ Supported | AWS exporters natively support credentials/role chains. |
| **Delivery mode: AtLeastOnce** | Disk buffer (256MB), block when full, retry | `file_storage` extension + persistent sending queue | ✅ Supported | OTEL persistent queue uses `file_storage` extension for disk-backed queuing with retry. |
| **Delivery mode: AtMostOnce** | Memory buffer, drop newest when full | In-memory sending queue, `drop_on_queue_full: true` | ✅ Supported | |
| **Compression** | gzip, zstd, snappy, zlib, lz4 (varies by sink) | gzip, zstd, snappy, zlib (varies by exporter) | ✅ Mostly supported | lz4 (Kafka-only) may need verification. |
| **Batch config** (MaxWrite) | `batch.max_bytes` / `batch.max_events` | `batch` processor or exporter-level `sending_queue` | ✅ Supported | |
| **Retry config** (min/max duration) | `request.retry_initial_backoff_secs`, `request.retry_max_duration_sec` | Exporter `retry_on_failure` with `initial_interval`, `max_interval`, `max_elapsed_time` | ✅ Supported | |
| **Rate limiting (output-level)** | `throttle` transform before sink | No official contrib rate limiter; Elastic's [`ratelimitprocessor`](https://pkg.go.dev/github.com/elastic/opentelemetry-collector-components/processor/ratelimitprocessor) exists | ⚠️ Partial | Elastic's processor supports per-key overrides via client metadata matching. An official contrib rate limit processor is [accepted (#35204)](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/35204) but not yet implemented. Per-container-file keying (Vector's `key_field: {{ _internal.file }}`) would need per-attribute keying instead. |
| **Proxy support** | `proxy.http` / `proxy.https` on HTTP sinks | Exporter-level `proxy_url` or env vars (`HTTP_PROXY`) | ✅ Supported | |
| **Health checks / metrics** | Vector API + `internal_metrics` source | OTEL health check extension + `prometheus` exporter or `zpages` extension | ✅ Supported | |

#### 6. Dynamic Templating

CLO uses dynamic templates for output field values (index names, topic names, log group names, etc.). These templates reference log record fields with a `{{.path.to.field}}` syntax, resolved at config generation time to Vector's template syntax (`{{ field }}`).

| Template Usage | Vector | OTEL | Status |
|---------------|--------|------|--------|
| Elasticsearch index name | `{{ field }}` in index template | `elasticsearch.index` attribute on log records | ✅ Supported | Since [v0.122.0](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/exporter/elasticsearchexporter), dynamic routing via `elasticsearch.index` attribute is the default. Use `transform` processor to set the attribute per-record based on log fields. |
| Kafka topic name | `{{ field }}` in topic template | Kafka exporter `topic_from_attribute` | ✅ Supported | Kafka exporter can derive topic from log attributes. |
| CloudWatch log group / stream | `{{ field }}` templates | Exporter config | ⚠️ Needs verification | |
| S3 key prefix | `{{ field }}` template | Exporter config or `routing` | ⚠️ Needs verification | |
| Splunk index/source/sourcetype | `{{ field }}` templates | Exporter mapping config | ⚠️ Needs design | |
| Syslog facility/severity/appname | VRL field extraction + syslog codec | `transform` processor + exporter config | ✅ Supported | |

### Summary of Gaps

#### Critical Gaps (❌ — no built-in solution, needs custom component)

1. **Generic HTTP output** — No OTEL exporter sends arbitrary JSON/NDJSON to generic HTTP endpoints. [Proposed but not implemented](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/10836). Needs a custom exporter.

2. **Azure Logs Ingestion output** — No OTEL exporter for Azure Data Collection Rules (Logs Ingestion API). [Issue #40478](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/40478) is open. Needs a custom exporter.

3. **Multi-line exception detection** (post-ingestion) — No OTEL processor merges multi-line stack traces after collection the way Vector's `detect_exceptions` transform does (language-specific detection for Java, Python, Go, Ruby, etc.). The `filelog` receiver's `multiline` config works at file ingestion (before CRI log parsing), which could merge lines at the file level but requires CRI-aware regex patterns and doesn't provide the language-specific exception boundary detection. The `exceptions` connector extracts exceptions from trace spans, not from log messages.

#### Partial Support (⚠️ — exists outside official contrib, or needs significant OTTL)

4. **Per-key rate limiting (throttle)** — Elastic's [`ratelimitprocessor`](https://pkg.go.dev/github.com/elastic/opentelemetry-collector-components/processor/ratelimitprocessor) supports per-key overrides via client metadata matching. An official contrib rate limit processor is [accepted (#35204)](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/35204) but not yet implemented. Vector's per-container-file keying would need adaptation to per-attribute keying.

5. **Kube API audit policy filter** — The K8s audit policy logic (wildcard matching on users/groups/resources, verb filtering, level assignment, field redaction) is very complex. OTTL can express the conditions but the wildcard-to-regex conversion and rule precedence would produce very verbose config. A dedicated processor would be cleaner.

6. **Log level detection (5-stage)** — OTTL's [`ExtractGrokPatterns`](https://github.com/open-telemetry/opentelemetry-collector-contrib/blob/main/pkg/ottl/ottlfuncs/README.md) (Elastic Go-Grok library) covers grok-based detection (stage 3). OTTL regex (`IsMatch`) covers stages 4-5. Stages 1-2 (logfmt, klog) lack dedicated parsers but can be approximated with grok/regex. Overall feasible but verbose.

7. **Dynamic field templating for output destinations** — Varies per exporter. Elasticsearch: ✅ (dynamic index via `elasticsearch.index` attribute, default since v0.122.0). Kafka: ✅ (`topic_from_attribute`). CloudWatch, S3, Splunk: needs verification of per-record attribute-based routing support. May need `routing` processor fan-out for exporters that don't support per-record destination fields.

8. **VIAQ data model compatibility** — Non-OTLP outputs (Elasticsearch, Splunk, Syslog, HTTP) currently receive logs in the VIAQ data model. With OTEL collector, these exporters serialize from OTLP format. The output schema may differ, potentially breaking existing consumer configurations. This needs careful mapping and potentially a "VIAQ compatibility mode" transform.

#### Fully Supported (✅)

- Container log collection (`filelog` + `k8sattributes`)
- Journal/node log collection (`journalctl` receiver)
- Audit file collection (`filelog` receiver)
- Syslog receiver
- Drop filter (`filter` processor)
- Prune filter (`transform` processor)
- Parse filter (`transform` processor)
- OpenShift labels filter (`attributes` processor)
- OTLP output (native)
- Elasticsearch, Kafka, Loki, LokiStack, Splunk, Syslog, CloudWatch, S3, Google Cloud Logging outputs
- TLS configuration
- Authentication (bearer, basic, AWS, SASL)
- Delivery modes (persistent queue for AtLeastOnce)
- Compression, batch, retry configuration
- Log routing by type/source
- Kubernetes metadata enrichment

### Recommended Approach for Milestone 1

#### Phase 1: Core pipeline (OTLP-only output path)
Start with the simplest end-to-end path:
- `filelog` receiver for container + audit logs
- `journalctl` receiver for node logs
- `k8sattributes` processor for Kubernetes metadata
- `transform` processor for log classification (application/infrastructure/audit)
- `resource` processor for cluster ID, hostname
- `otlp`/`otlphttp` exporter

This validates the basic collection pipeline without needing complex transforms or exotic exporters.

#### Phase 2: Add supported outputs
Add exporters one by one, starting with the most commonly used:
- Elasticsearch, Kafka, Loki/LokiStack, CloudWatch, Splunk, Syslog, S3, Google Cloud Logging

#### Phase 3: Filters and normalization
Implement CLO filters as OTEL processors:
- Drop → `filter` processor
- Prune → `transform` processor
- Parse → `transform` processor
- OpenShift labels → `attributes` processor
- Log level detection → custom processor or OTTL regex approximation
- Kube API audit → custom processor or extensive OTTL

#### Phase 4: Address gaps
- Generic HTTP exporter (custom or community)
- Azure Logs Ingestion exporter (custom)
- Rate limiting processor (custom)
- Multi-line exception detection (custom processor or integration with filelog)
- VIAQ data model compatibility for non-OTLP outputs