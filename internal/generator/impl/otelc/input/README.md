# OpenTelemetry Collector Input Package

This package contains input-specific configurations for OpenTelemetry Collector receivers, using the operator types defined in `internal/generator/impl/otelc/api/receivers/operators`.

## Purpose

The `input` package provides high-level, opinionated receiver pipelines for common log input sources. These functions create complete operator chains that:
1. Parse input-specific log formats
2. Extract metadata
3. Map to standardized, non-deprecated attributes
4. Transform data for compatibility with observability backends

## Structure

- `factory.go` - Entry point: `input.New(input, resNames, secrets, op)` maps an adapted forwarder input to its receivers (currently container sources only)

### Container Logs (`container/` subpackage)

Functions for parsing CRI-O container logs from Kubernetes pods, mapping them to the attributes defined in the [Red Hat Observability Data Model](https://github.com/rhobs/observability-data-model/blob/main/cluster-logging.md).

**Available Functions:**

- **`container.New(spec obs.InputSpec)`** - Creates a `FileLog` receiver for a container input source, wired with the CRI-O operator pipeline

- **`container.NewOperators(clusterUID, nodeName, logSource, logType string)`** - The complete CRI-O parser pipeline
  - Parses CRI-O log format: `<timestamp> <stream> <tags> <log>`
  - Extracts Kubernetes metadata from file path
  - Maps to non-deprecated attributes (k8s.*, log.iostream)
  - Adds `k8s.node.name` and the OpenShift attributes when the corresponding arguments are non-empty:
    - `openshift.cluster.uid`
    - `openshift.log.source`
    - `openshift.log.type`

## Attribute Mapping

### Non-Deprecated Attributes

The operator pipelines map CRI-O container logs to these **non-deprecated** attributes:

#### Resource Attributes (same for all records in a stream)
- `k8s.namespace.name` ← extracted from file path
- `k8s.pod.name` ← extracted from file path
- `k8s.pod.uid` ← extracted from file path
- `k8s.container.name` ← extracted from file path
- `k8s.container.restart_count` ← extracted from file path
- `k8s.node.name` ← provided via parameter (optional)
- `openshift.cluster.uid` ← provided via parameter (optional)
- `openshift.log.source` ← provided via parameter (optional)
- `openshift.log.type` ← provided via parameter (optional)

#### Log Attributes (can vary per record)
- `log.iostream` ← parsed from CRI-O stream field (stdout/stderr)

### Deprecated Attributes (NOT used)

These attributes from the legacy ViaQ data model are **not** used:
- ❌ `kubernetes.namespace_name` → use `k8s.namespace.name`
- ❌ `kubernetes.pod_name` → use `k8s.pod.name`
- ❌ `kubernetes.container_name` → use `k8s.container.name`
- ❌ `kubernetes.host` → use `k8s.node.name`
- ❌ `openshift.cluster_id` → use `openshift.cluster.uid`
- ❌ `log_type` → use `openshift.log.type`
- ❌ `log_source` → use `openshift.log.source`

## Usage Example

```go
package main

import (
    "github.com/openshift/cluster-logging-operator/internal/generator/impl/otelc/api/receivers"
    "github.com/openshift/cluster-logging-operator/internal/generator/impl/otelc/input/container"
    "gopkg.in/yaml.v3"
)

func main() {
    // Create FileLog receiver (name, include patterns...)
    fileLog := receivers.NewFileLog("", "/var/log/pods/*/*/*.log")
    fileLog.StartAt = "end"
    fileLog.IncludeFilePath = true

    // Add the CRI-O operator pipeline
    fileLog.Operators = container.NewOperators(
        "my-cluster-uuid",
        "my-node-name",
        "container",
        "application",
    )

    // Marshal to YAML
    data, _ := yaml.Marshal(fileLog)
    println(string(data))
}
```

### Output

```yaml
include:
  - /var/log/pods/*/*/*.log
start_at: end
include_file_path: true
operators:
  - type: regex_parser
    id: parser-crio
    output: extract_metadata_from_filepath
    regex: ^(?P<time>[^ Z]+) (?P<stream>stdout|stderr) (?P<logtag>[^ ]*) ?(?P<log>.*)$
    timestamp:
      parse_from: attributes.time
      layout_type: gotime
      layout: "2006-01-02T15:04:05.999999999Z07:00"
  # ... (full pipeline as defined in container/operators.go)
```

## Testing

Run the test suite:

```bash
go test -v github.com/openshift/cluster-logging-operator/internal/generator/impl/otelc/input/...
```

## References

- [Red Hat Observability Data Model - Cluster Logging](https://github.com/rhobs/observability-data-model/blob/main/cluster-logging.md)
- [OpenTelemetry Collector FileLog Receiver](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/receiver/filelogreceiver)
- [CRI-O Log Format](https://github.com/cri-o/cri-o/blob/main/docs/crio.8.md#logging)
