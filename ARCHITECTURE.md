# Cluster Logging Operator - Architecture

This document describes the internal architecture of the Cluster Logging Operator (CLO), including design decisions, key components, and important tradeoffs.

## Overview

The Cluster Logging Operator manages log collection and forwarding in OpenShift clusters using a declarative, Kubernetes-native approach. Users define what logs to collect and where to send them via the ClusterLogForwarder CRD, and the operator translates this into a working deployment of Vector collectors.

### Core Design Principles

1. **Declarative Configuration**: Users specify desired state via CRDs; the operator ensures the cluster matches that state
2. **Separation of Concerns**: Collector logic is separated from forwarding logic through Vector's transform and sink model
3. **Multi-tenancy**: Infrastructure, application, and audit logs are isolated to support per-tenant controls
4. **Template-Based Generation**: Complex configurations are generated from templates to ensure consistency and maintainability

## Architecture Diagram

```text
┌─────────────────────────────────────────────────────────────────────┐
│                         User (via kubectl)                          │
└────────────────────────────┬────────────────────────────────────────┘
                             │
                             ▼
┌─────────────────────────────────────────────────────────────────────┐
│  ClusterLogForwarder CRD (User Intent)                              │
│  ├─ Inputs (what logs to collect)                                   │
│  ├─ Outputs (where to send logs)                                    │
│  ├─ Pipelines (how to route logs)                                   │
│  └─ Filters (transformations to apply)                              │
└────────────────────────────┬────────────────────────────────────────┘
                             │
                             ▼
┌─────────────────────────────────────────────────────────────────────┐
│         Cluster Logging Operator Controller                         │
│  ┌──────────────────────────────────────────────────────────────┐   │
│  │ 1. Validate CRD                                              │   │
│  │ 2. Generate Vector Config                                    │   │
│  │ 3. Create/Update Secrets & ConfigMaps                        │   │
│  │ 4. Deploy Collector DaemonSets/Deployments                   │   │
│  └──────────────────────────────────────────────────────────────┘   │
└────────────────────────────┬────────────────────────────────────────┘
                             │
                             ▼
┌─────────────────────────────────────────────────────────────────────┐
│              Vector Collectors (on each node)                       │
│  ┌────────────────┐  ┌──────────────┐  ┌──────────────┐             │
│  │ Sources        │  │ Transforms   │  │ Sinks        │             │
│  │ - journald     │→ │ - normalize  │→ │ - Loki       │             │
│  │ - containers   │  │ - filter     │  │ - Splunk     │             │
│  │ - audit        │  │ - enrich     │  │ - CloudWatch │             │
│  └────────────────┘  └──────────────┘  └──────────────┘             │
└─────────────────────────────────────────────────────────────────────┘
```

## Core Components

### 1. Controllers (`internal/controller/`)

**ClusterLogForwarder Controller** (`observability/clusterlogforwarder_controller.go`)
- Watches ClusterLogForwarder resources for changes
- Validates CRD specifications
- Manages Secret and ConfigMap creation for collector credentials
- Orchestrates Vector deployment (DaemonSets/Deployments)
- Handles reconciliation and error reporting

**LogFileMetricsExporter Controller** (`logfilemetricsexporter/logfilemetricsexporter_controller.go`)
- Manages LogFileMetricsExporter resources
- Exposes metrics about log file processing
- Provides observability into the logging system

### 2. API Definitions (`api/observability/v1/`)

**ClusterLogForwarder Types** (`clusterlogforwarder_types.go`)
- Main CRD definition for user-facing configuration
- Defines the structure for inputs, outputs, pipelines, and filters
- Includes validation rules

**Output Types** (`output_types.go`)
- Specifications for each supported output type
- Examples: Kafka, Splunk, CloudWatch, Google Cloud Logging, Elasticsearch, Azure, etc.
- Handles output-specific configuration parameters

### 3. Configuration Generator (`internal/generator/`)

The configuration generator translates ClusterLogForwarder specs into Vector configurations.

**Key Design**: Template-based generation for consistency and maintainability

```text
     ClusterLogForwarder Spec
               │
               ▼
    ┌─────────────────────────┐
    │  Generator Module       │
    │  ├─ inputs.go           │
    │  ├─ outputs.go          │
    │  ├─ filters.go          │
    │  └─ pipelines.go        │
    └──────────┬──────────────┘
               │
               ▼
    Vector Configuration (TOML)
```

**Output Structure** (`internal/generator/vector/output/`)

Each output type is a separate package implementing the Element interface:
- `aws/` - AWS CloudWatch
- `azurelogsingestion/` - Azure Logs Ingestion
- `azuremonitor/` - Azure Monitor
- `elasticsearch/` - Elasticsearch
- `gcl/` - Google Cloud Logging
- `http/` - Generic HTTP
- `kafka/` - Kafka
- `loki/` - Grafana Loki
- `lokistack/` - LokiStack (managed)
- `otlp/` - OpenTelemetry Protocol
- `splunk/` - Splunk
- `syslog/` - Syslog

Each output module:
1. Defines Go types for configuration
2. Implements template-based Vector TOML generation
3. Handles output-specific validation and transformations

### 4. Data Flow Architecture

Logs flow through the Vector collector following this pattern:

```text
Sources → internal preprocessing → global transforms → output-specific transforms → sinks
```

**Minimal Required Attributes**:
- `log_type` - classification (application, infrastructure, audit)
- `log_source` - specific source (e.g., node, container, kubeAPI)
- `timestamp` - when the log was generated
- `message` - log content
- `level` - log severity
- `kubernetes_metadata` - pod, namespace, labels, etc.

### 5. Key Directories

- `api/` - API/CRD definitions
- `cmd/` - Operator entry point
- `config/` - Kubernetes manifests (RBAC, CRDs, ServiceAccount)
- `internal/controller/` - Reconciliation logic
- `internal/generator/` - Configuration generation system
- `internal/validations/` - Input validation logic
- `test/functional/` - Functional tests for outputs
- `test/e2e/` - End-to-end integration tests
- `hack/` - Build scripts and utilities
- `docs/` - Documentation

## Design Decisions

### 1. Vector as the Collector

**Decision**: Use Vector for log collection and forwarding

**Rationale**:
- High-performance, memory-efficient collector
- Extensive output support (100+ destinations)
- Strong community and maintenance
- ViaQ data model compatibility

**Tradeoff**: Operator must stay updated with Vector releases and API changes

### 2. Template-Based Configuration Generation

**Decision**: Generate Vector configs from Go templates rather than programmatic construction

**Rationale**:
- Maintains consistency across different output types
- Easier to review and modify configurations
- Simpler to add new outputs
- Closer to how users would write configurations manually

**Tradeoff**: Requires careful template management and testing

### 3. Separate Input, Output, and Transform Definitions

**Decision**: Keep inputs, outputs, and filters as distinct CRD fields with separate pipeline specification

**Rationale**:
- Allows flexible routing of logs from multiple inputs to multiple outputs
- Simplifies validation of each component independently
- Enables reuse of filter definitions

**Tradeoff**: More verbose CRD structure; requires careful validation to ensure pipeline validity

### 4. DaemonSet-Based Collection

**Decision**: Deploy Vector as a DaemonSet on each node for node-level logs, Deployments for centralized collection

**Rationale**:
- DaemonSets ensure collection from every node
- Reduces network hops for node-local logs
- Deployments handle cluster-wide logs (API audit, etc.)

**Tradeoff**: Multiple collection processes may have higher resource overhead than single centralized collector

## Important Tradeoffs Still in Effect

### 1. Complexity vs. Flexibility

**Trade**: Supporting complex log routing (multiple inputs → transformations → multiple outputs) increases operator complexity.

**Mitigation**: Validate pipelines thoroughly; provide clear error messages.

### 2. Performance vs. Compatibility

**Trade**: Supporting older OpenShift versions requires maintaining compatibility with older Kubernetes APIs.

**Mitigation**: Use operator-sdk abstractions; test against supported versions.

### 3. Security vs. Usability

**Trade**: Storing secrets securely requires additional complexity (Secret management, RBAC validation).

**Mitigation**: Use Kubernetes Secrets; validate RBAC permissions; document security best practices.

### 4. Configuration Customization vs. Maintainability

**Trade**: Allowing arbitrary Vector configuration increases flexibility but risks unsupported or broken configurations.

**Mitigation**: Provide a well-defined CRD with validation; discourage direct Vector config modification.

## Adding New Features

### Adding a New Output Type

See [docs/contributing/how-to-add-new-output.md](docs/contributing/how-to-add-new-output.md) for detailed steps.

Basic approach:
1. Add output type definition to `api/observability/v1/output_types.go`
2. Create generator package in `internal/generator/vector/output/[type]/`
3. Implement Element interface with Go templates for TOML generation
4. Register in `internal/generator/vector/outputs.go`
5. Add functional tests in `test/functional/outputs/`

### Adding a New Input Type

Similar process to outputs:
1. Update `api/observability/v1/clusterlogforwarder_types.go` with new input type
2. Create input generator in `internal/generator/vector/input/[type]/`
3. Add validation logic to controller
4. Test with functional tests

### Adding a New Filter

Filters are transformations applied to logs:
1. Add filter definition to `api/observability/v1/clusterlogforwarder_types.go`
2. Implement in `internal/generator/vector/filter/`
3. Update pipeline logic to apply filters correctly
4. Add tests

## Testing Strategy

### Unit Tests

Test individual components in isolation:
- Template generation correctness
- Configuration validation
- Data transformations

Location: Throughout codebase with `*_test.go` files

### Functional Tests

Test output connector integration:
- Actual log forwarding to supported destinations
- Connection handling and credential management
- Data format compliance

Location: `test/functional/outputs/`

### E2E Tests

Test full operator functionality:
- CRD application and reconciliation
- Collector deployment and configuration
- End-to-end log flow

Location: `test/e2e/`

## Dependency Points

### External Dependencies

1. **Vector** - The actual log collector (versioned in Dockerfile)
2. **operator-sdk** - Kubernetes operator framework
3. **controller-runtime** - Kubernetes controller libraries
4. **client-go** - Kubernetes client library
5. **loki-operator** - For managed LokiStack integration (optional)

### Internal Dependencies

- `api/` packages imported by `internal/controller/`
- `internal/generator/` imported by controllers for config generation
- Controllers depend on Kubernetes primitives via client-go

## Version Compatibility

- **Go**: 1.24+ (see `go.mod` for the exact version)
- **Vector**: See Dockerfile for the exact version

## Monitoring and Observability

### Metrics Exposed

- Reconciliation timing and success rates
- Vector deployment status
- Configuration generation statistics
- Resource usage (via Vector itself)

### Logs

- Controller reconciliation logs
- Configuration generation debug information
- Error reporting on CRD validation failures

### Status Conditions

ClusterLogForwarder Status includes:
- Validation status (Valid/Invalid)
- Deployment status (Deployed/Failing)
- Collector readiness status

## Security Considerations

1. **Secret Management**: Credentials stored in Kubernetes Secrets, not in configs
2. **RBAC**: Operator validates permissions before accessing external services
3. **Network**: TLS/mTLS configuration for secure communication with outputs
4. **Audit**: Audit log collection with special handling for sensitive data

