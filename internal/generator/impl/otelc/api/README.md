# OpenTelemetry Collector API

This package provides Go structs for generating OpenTelemetry Collector configurations with YAML-only serialization, similar to the vector API in `internal/generator/vector/api`.

## Structure

### API Package
- `config.go` - Top-level `Config` and `Service` types; an collector configuration is assembled by adding receivers, exporters, extensions, and pipelines
- `types/` - Core types and interfaces shared by receivers, exporters, extensions, and pipelines
- `receivers/` - Receiver implementations (currently: FileLog)
  - `receivers/operators/` - Generic operator types for log processing pipelines
- `receivers.go` - Collection and unmarshaling logic for receivers
- `exporters/` - Exporter implementations (currently: OTLP HTTP)
- `exporters.go` - Collection and unmarshaling logic for exporters
- `extensions/` - Extension implementations (currently: bearertokenauth)
- `extensions.go` - Collection logic for extensions
- `pipelines.go` - `Pipeline`/`Pipelines` types connecting receivers to exporters
- `unmarshal_helpers.go` - Helpers for custom YAML unmarshaling

### Related Packages
- `internal/generator/impl/otelc/` - The otelc generator as a whole
  - `factory.go` - Entry point translating a `ClusterLogForwarder` spec into an `api.Config`
  - `input/` - Input-specific receiver pipelines (e.g., CRI-O container logs)
    - See [input package README](../input/README.md) for details
  - `output/` - Output-specific exporter pipelines (currently: Lokistack via OTLP HTTP)
  - `helpers/tls/` - TLS client configuration helpers

## Supported Receivers

### FileLog Receiver

The FileLog receiver reads logs from files on the filesystem. It is based on the [opentelemetry-collector-contrib filelog receiver](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/receiver/filelogreceiver).

#### Features

- File selection with include/exclude patterns
- Multiple reading behaviors (start position, polling interval, encoding)
- Multiline log support
- File metadata attributes
- Retry on failure
- Operator-based log processing
- Custom attributes and resource labels

## Supported Exporters

### OTLP HTTP Exporter

The OTLP HTTP exporter (`OtlpHttp`) sends logs, metrics, and traces via HTTP using the OTLP protocol. This is the recommended way to send logs to Grafana Loki v3+ using native OTLP ingestion.

#### Features

- HTTP/HTTPS endpoint configuration
- TLS/SSL support
- Custom headers (e.g., tenant ID)
- Compression (gzip, none)
- Encoding (proto, json)
- Queue and retry configuration
- Connection pool settings

#### Use with Grafana Loki

For Loki v3+, use the OTLP endpoint: `http://loki:3100/otlp`

Multi-tenancy can be configured using the `X-Scope-OrgID` header.

## Usage Examples

### Receivers Example

```go
package main

import (
	"fmt"
	"gopkg.in/yaml.v3"

	"github.com/openshift/cluster-logging-operator/internal/generator/impl/otelc/api"
	"github.com/openshift/cluster-logging-operator/internal/generator/impl/otelc/api/receivers"
	"github.com/openshift/cluster-logging-operator/internal/generator/impl/otelc/api/receivers/operators"
)

func main() {
	// Create a new FileLog receiver (name, include patterns...)
	fileLog := receivers.NewFileLog("", "/var/log/app/*.log", "/var/log/service/*.log")
	fileLog.StartAt = "beginning"
	fileLog.Encoding = "utf-8"
	fileLog.MaxConcurrentFiles = 256
	fileLog.Attributes = map[string]interface{}{
		"log.type":    "application",
		"environment": "production",
	}

	// Add multiline support
	fileLog.Multiline = &receivers.Multiline{
		LineStartPattern: "^\\d{4}-\\d{2}-\\d{2}",
	}

	// Add operators for log processing
	fileLog.Operators = []operators.Operator{
		{
			Type: operators.OperatorTypeJSONParser,
			ID:   "json_parse",
		},
	}

	// Create receivers collection
	receiversMap := api.Receivers{
		"file_log/app": fileLog,
	}

	// Marshal to YAML
	data, err := yaml.Marshal(receiversMap)
	if err != nil {
		panic(err)
	}

	fmt.Println(string(data))
}
```

### Output

```yaml
file_log/app:
  include:
    - /var/log/app/*.log
    - /var/log/service/*.log
  start_at: beginning
  encoding: utf-8
  max_concurrent_files: 256
  attributes:
    environment: production
    log.type: application
  multiline:
    line_start_pattern: ^\d{4}-\d{2}-\d{2}
  operators:
    - type: json_parser
      id: json_parse
```

### Exporters Example

```go
package main

import (
	"fmt"
	"gopkg.in/yaml.v3"

	"github.com/openshift/cluster-logging-operator/internal/generator/impl/otelc/api"
	"github.com/openshift/cluster-logging-operator/internal/generator/impl/otelc/api/exporters"
	"github.com/openshift/cluster-logging-operator/internal/generator/impl/otelc/api/types"
)

func main() {
	// Create OTLP HTTP exporter for Loki (name, endpoint)
	otlphttp := exporters.NewOtlpHttp("", "http://loki:3100/otlp")
	otlphttp.Encoding = "proto"
	otlphttp.Compression = "gzip"
	otlphttp.Headers = map[string]string{
		"X-Scope-OrgID": "tenant1",
	}
	otlphttp.TLS = &types.TlsClientConfig{
		Insecure: true,
	}
	otlphttp.SendingQueue = &types.QueueSettings{
		Enabled:      true,
		NumConsumers: 5,
		QueueSize:    1000,
	}
	otlphttp.RetryOnFailure = &types.RetrySettings{
		Enabled:         true,
		InitialInterval: "5s",
		MaxInterval:     "30s",
		MaxElapsedTime:  "5m",
	}

	// Create exporters collection
	exportersMap := api.Exporters{
		"otlphttp/loki": otlphttp,
	}

	// Marshal to YAML
	data, err := yaml.Marshal(exportersMap)
	if err != nil {
		panic(err)
	}

	fmt.Println(string(data))
}
```

#### Output

```yaml
otlphttp/loki:
  endpoint: http://loki:3100/otlp
  tls:
    insecure: true
  compression: gzip
  encoding: proto
  headers:
    X-Scope-OrgID: tenant1
  sending_queue:
    enabled: true
    num_consumers: 5
    queue_size: 1000
  retry_on_failure:
    enabled: true
    initial_interval: 5s
    max_interval: 30s
    max_elapsed_time: 5m
```

## YAML-Only Serialization

Unlike the vector API which supports both TOML and YAML, this API only uses YAML serialization tags. This is because OpenTelemetry Collector uses YAML for its configuration files.

All structs use only `yaml:` tags with appropriate options:
- `yaml:"field_name"` - Required field
- `yaml:"field_name,omitempty"` - Optional field (omitted if zero value)
- Fields without tags are not serialized

## Testing

Run the tests:

```bash
cd internal/generator/impl/otelc/api
go test -v
```

## References

### Receivers
- [OpenTelemetry Collector Contrib - FileLog Receiver](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/receiver/filelogreceiver)

### Exporters
- [OpenTelemetry Collector - OTLP HTTP Exporter](https://github.com/open-telemetry/opentelemetry-collector/tree/main/exporter/otlphttpexporter)
- [Grafana Loki - OTLP Ingestion](https://grafana.com/docs/loki/latest/send-data/otel/)

### General
- [OpenTelemetry Collector Configuration](https://opentelemetry.io/docs/collector/configuration/)
