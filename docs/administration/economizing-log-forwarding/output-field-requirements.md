# Output Field Requirements for Log Pruning

This document lists the log entry fields that each output type depends on. When
using `prune` filters to reduce log entry size, ensure you preserve the fields
your output requires.

Fields that the ClusterLogForwarder **always** requires regardless of output
(pruning these is blocked by validation):

- `.log_type`
- `.log_source`
- `.message`

## LokiStack

**Required fields:**

| Field | Reason |
|-------|--------|
| `.log_type` | Used in `lokistack.go` for tenant routing (application/infrastructure/audit) |
| `.kubernetes.namespace_name` | Used in `loki.go` as default Loki stream label (`lokiLabelKubernetesNamespaceName`) |
| `.kubernetes.pod_name` | Used in `loki.go` as default Loki stream label (`lokiLabelKubernetesPodName`) |
| `.kubernetes.container_name` | Used in `loki.go` as default Loki stream label (`lokiLabelKubernetesContainerName`) for container logs |
| `.hostname` | Used in `loki.go` as Loki stream label (`lokiLabelKubernetesHost`), mapped to `${VECTOR_SELF_NODE_NAME}` |

**Safe to prune:** All other fields. Loki indexes by labels only; pruned fields
reduce storage cost without affecting queryability by the label dimensions.

**Gotchas:**
- If `labelKeys` is customized in the LokiStack spec, those custom label fields become required
- OTel-flavored labels (e.g., `k8s.namespace_name`, `openshift.log_type`) are automatically added when corresponding ViAQ labels are present
- The `RemapLabels` transform in `loki.go` creates empty string values for missing container labels to prevent stream label gaps

## Google Cloud Logging

**Required fields:**

| Field | Reason |
|-------|--------|
| `.hostname` | Used in `gcl.go` as the `node_name` in the `k8s_node` resource map |
| `.log_type` | Used in `gcl.go` `NormalizeSeverity` VRL to set audit log level to "INFO" |
| `.level` | Used in `gcl.go` as `SeverityKey` for GCL log severity |

**Safe to prune:** Most other fields. GCL uses the `logId` template (user-configured) and severity/hostname; other fields are passed through in the JSON encoding.

**Gotchas:**
- The `logId` template may reference additional fields depending on your configuration (e.g., `{.kubernetes.namespace_name}`)
- If `.level` is missing, defaults to "DEFAULT"
- `.level` values "warn" and "trace" are normalized to "WARNING" and "DEBUG" respectively

## Elasticsearch

**Required fields:**

| Field | Reason |
|-------|--------|
| `.kubernetes.event.metadata.uid` | Used in `elasticsearch.go` for Elasticsearch v6 document `_id` (if present) |

**Safe to prune:** Most fields. Elasticsearch stores the full JSON document, so pruning reduces index size but doesn't break functionality.

**Gotchas:**
- The `index` template (user-configured) may reference additional fields (e.g., `{.log_type}-{.kubernetes.namespace_name}`)
- For Elasticsearch v6, a UUID `_id` is generated if `.kubernetes.event.metadata.uid` is missing
- Version 7+ does not use document IDs

## Splunk

**Required fields:**

| Field | Reason |
|-------|--------|
| `.log_type` | Used in `splunk.go` `sourceTmpl` VRL to detect infrastructure vs. audit log types |
| `.log_source` | Used in `splunk.go` `sourceTmpl` VRL to detect container vs. node source, and as the audit source value |
| `.systemd.u.SYSLOG_IDENTIFIER` | Used in `splunk.go` `sourceTmpl` VRL as Splunk `source` for infrastructure node logs |
| `.kubernetes.namespace_name` | Used in `splunk.go` `sourceTmpl` VRL to construct Splunk `source` for container logs (joined with pod and container names) |
| `.kubernetes.pod_name` | Used in `splunk.go` `sourceTmpl` VRL to construct Splunk `source` for container logs |
| `.kubernetes.container_name` | Used in `splunk.go` `sourceTmpl` VRL to construct Splunk `source` for container logs |
| `.hostname` | Used in `splunk.go` as Splunk HEC `HostKey` (`._internal.hostname`) |
| `.timestamp` | Used in `splunk.go` as Splunk HEC `TimestampKey` (`._internal.timestamp`), parsed via `fixTimestampFormat` |

**Safe to prune:** Fields not listed above, unless referenced in user-configured `index`, `source`, `sourceType`, or `indexedFields` templates.

**Gotchas:**
- If `source`, `sourceType`, or `payloadKey` are user-configured, those templates may reference additional fields
- `indexedFields` configuration extracts nested fields to root level (field paths specified by user)
- If the `payloadKey` field is an object, `sourcetype` is set to "_json"; otherwise "generic_single_line"
- The default `source` construction joins Kubernetes metadata with underscores for container logs

## CloudWatch

**Required fields:**

| Field | Reason |
|-------|--------|
| `.log_type` | Used in `cloudwatch.go` `NormalizeStreamName` VRL to detect audit vs. infrastructure log types |
| `.log_source` | Used in `cloudwatch.go` `NormalizeStreamName` VRL to detect container vs. node source |
| `.hostname` | Used in `cloudwatch.go` `NormalizeStreamName` VRL to construct stream name for audit and infrastructure logs |
| `.kubernetes.namespace_name` | Used in `cloudwatch.go` `NormalizeStreamName` VRL to construct stream name for container logs |
| `.kubernetes.pod_name` | Used in `cloudwatch.go` `NormalizeStreamName` VRL to construct stream name for container logs |
| `.kubernetes.container_name` | Used in `cloudwatch.go` `NormalizeStreamName` VRL to construct stream name for container logs |

**Safe to prune:** All other fields. CloudWatch stream names are the primary routing mechanism; other fields are stored in the JSON payload.

**Gotchas:**
- The `groupName` template (user-configured) may reference additional fields
- Stream name defaults to "default" if metadata is missing
- Container stream names are constructed as `namespace_pod_container` (joined with underscores)
- Audit stream names use `hostname.log_source` format
- Infrastructure stream names use `hostname.default` format
- Node stream names use `hostname.journal.system` format

## Kafka

**Required fields:**

None beyond the global required fields (`.log_type`, `.log_source`, `.message`).

**Safe to prune:** All other fields. Kafka stores the full JSON document.

**Gotchas:**
- The `topic` template (user-configured) may reference additional fields (e.g., `{.log_type}`)
- If no topic is configured, defaults to "topic"
- Timestamp is encoded in RFC3339 format

## Syslog

**Required fields:**

| Field | Reason |
|-------|--------|
| `.log_type` | Used in `syslog.go` conditional defaults for infrastructure vs. audit log formatting |
| `.log_source` | Used in `syslog.go` conditional defaults to detect container vs. node source; used as audit `app_name` and `msg_id` defaults |
| `.hostname` | Implicitly used for syslog hostname field (not explicitly in code but part of syslog protocol) |
| `.systemd.u.SYSLOG_IDENTIFIER` | Used in `syslog.go` as default `app_name` for infrastructure node logs (RFC3164 and RFC5424) |
| `.systemd.t.PID` | Used in `syslog.go` as default `proc_id` for infrastructure node logs (RFC3164 and RFC5424) |
| `.kubernetes.namespace_name` | Used in `syslog.go` to construct default `app_name` for container logs (joined with pod and container names) |
| `.kubernetes.pod_name` | Used in `syslog.go` to construct default `app_name` for container logs |
| `.kubernetes.container_name` | Used in `syslog.go` to construct default `app_name` for container logs |
| `.kubernetes.pod_id` | Used in `syslog.go` as default `proc_id` for container logs |
| `.level` | Used in `syslog.go` as default `severity` for container logs |
| `.auditID` | Used in `syslog.go` as default `proc_id` for audit logs (RFC5424) |

**Safe to prune:** Fields not listed above, unless referenced in user-configured `payloadKey` or syslog field templates (`facility`, `severity`, `appName`, `procId`, `msgId`).

**Gotchas:**
- RFC3164 vs. RFC5424 have different field requirements and defaults
- If `enrichment` is set to `KubernetesMinimal`, Kubernetes metadata fields (`.kubernetes.namespace_name`, `.kubernetes.container_name`, `.kubernetes.pod_name`) are prepended to the message body
- Syslog severity and facility fields support both names ("info", "user") and numeric codes (6, 1), with normalization applied
- For container logs, RFC3164 `app_name` is sanitized (non-alphanumeric characters removed) and truncated to 32 characters
- If Kubernetes metadata is missing for container logs, defaults or errors are logged
- The `payloadKey` configuration determines what becomes the `.message` field; if not configured, the full log entry (excluding `_internal` and `_syslog`) is serialized

## HTTP

**Required fields:**

None beyond the global required fields (`.log_type`, `.log_source`, `.message`).

**Safe to prune:** All other fields. HTTP outputs forward the full JSON document.

**Gotchas:**
- HTTP is a generic sink; field requirements depend entirely on the receiving endpoint's expectations
- Users should consult the documentation of their HTTP endpoint to determine required fields

## S3

**Required fields:**

None beyond the global required fields (`.log_type`, `.log_source`, `.message`).

**Safe to prune:** All other fields. S3 stores the full JSON document.

**Gotchas:**
- The `keyPrefix` template (user-configured) may reference additional fields (e.g., `logs/{.log_type}/{@timestamp|date}/`)
- S3 object keys support timestamp patterns (`@timestamp|year`, `@timestamp|month`, etc.)

## OTLP

**Required fields:**

| Field | Reason |
|-------|--------|
| `.log_source` | Used in `otlp.go` `RouteBySource` to route logs to source-specific transforms; also set as `openshift.log.source` resource attribute in `transform.go` |
| `.log_type` | Set as `openshift.log.type` resource attribute in `transform.go` `BaseResourceAttributes` |
| `.hostname` | Set as `k8s.node.name` resource attribute in `transform.go` `BaseResourceAttributes` |
| `.openshift.cluster_id` | Set as `openshift.cluster.uid` resource attribute in `transform.go` `BaseResourceAttributes` |
| `.openshift.labels` | Iterated in `transform.go` `BaseResourceAttributes` to set `openshift.label.*` resource attributes |
| `."@timestamp"` | Used in `transform.go` `LogRecord` to set `timeUnixNano` via `parse_timestamp!(.@timestamp, ...)`. ViaQ normalization creates both `.timestamp` and `."@timestamp"` from `._internal.timestamp`; the OTLP transform consumes `@timestamp`. |
| `.level` | Used in `transform.go` `LogRecordSeverity` to set `severityText` |
| `.kubernetes.pod_name` | Used in `transform.go` `ContainerResourceAttributes` to set `k8s.pod.name` resource attribute for container logs |
| `.kubernetes.pod_id` | Used in `transform.go` `ContainerResourceAttributes` to set `k8s.pod.uid` resource attribute for container logs |
| `.kubernetes.container_name` | Used in `transform.go` `ContainerResourceAttributes` to set `k8s.container.name` resource attribute for container logs |
| `.kubernetes.namespace_name` | Used in `transform.go` `ContainerResourceAttributes` to set `k8s.namespace.name` resource attribute for container logs |
| `.kubernetes.labels` | Iterated in `transform.go` `ContainerResourceAttributes` to set `k8s.pod.label.*` resource attributes for container logs |
| `.kubernetes.container_iostream` | Used in `transform.go` `ContainerLogAttributes` to set `log.iostream` log attribute for container logs |
| `.systemd.t.CMDLINE` | Used in `transform.go` `NodeResourceAttributes` to set `process.command_line` resource attribute for node logs |
| `.systemd.t.COMM` | Used in `transform.go` `NodeResourceAttributes` to set `process.executable.name` resource attribute for node logs |
| `.systemd.t.EXE` | Used in `transform.go` `NodeResourceAttributes` to set `process.executable.path` resource attribute for node logs |
| `.systemd.t.PID` | Used in `transform.go` `NodeResourceAttributes` to set `process.pid` resource attribute for node logs |
| `.systemd.t.SYSTEMD_UNIT` | Used in `transform.go` `NodeResourceAttributes` to set `service.name` resource attribute for node logs |
| `._internal.message` | Used in `transform.go` `BodyFromInternal` as the log body for audit logs |
| `._internal.trace_id` | Used in `transform.go` `LogRecordTraceContext` to set `traceId` if present |
| `._internal.span_id` | Used in `transform.go` `LogRecordTraceContext` to set `spanId` if present |
| `._internal.trace_flags` | Used in `transform.go` `LogRecordTraceContext` to set `flags` if present |

**Safe to prune:** Fields not listed above. OTLP transforms extract specific fields into resource attributes and log records; unmapped fields are not included in the OTLP payload.

**Gotchas:**
- OTLP output generates extensive resource attributes and structured log records from log entry fields
- Different log sources (container, node, audit) have different field requirements
- For container logs, `.message` or `.structured` is used as the log body; for audit logs, `._internal.message` is used
- Trace context fields (`trace_id`, `span_id`, `trace_flags`) are optional but included if present
- Backward compatibility attributes are added (e.g., `log_type`, `kubernetes.pod_name`) in addition to OTel semantic convention attributes
- OVN audit logs parse the message to extract sequence number and component into log attributes
- Host audit logs parse the message to extract `auditd.type` and sequence number into log attributes

## Azure Monitor

**Required fields:**

None beyond the global required fields (`.log_type`, `.log_source`, `.message`).

**Safe to prune:** All other fields. Azure Monitor stores the full JSON document.

**Gotchas:**
- Azure Monitor is a relatively simple output with minimal field-specific logic
- The `logType` configuration (user-specified) determines the Azure Monitor log table name but doesn't reference log entry fields
