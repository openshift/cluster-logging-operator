# Developer guide

This file provides guidance to human developers and AI agents (such as Claude Code claude.ai/code)
when working with code in this repository.

## Overview

The Cluster Logging Operator (CLO) is a Kubernetes operator for configuring log collection and forwarding on OpenShift clusters. It manages the deployment of Vector as the log collector and supports complex log routing through inputs, outputs, filters, and pipelines defined via the ClusterLogForwarder CRD.

## Common Development Commands

### Building and Running
- `make build` - Build the operator binary
- `make build-debug` - Build with debug flags
- `make run` - Run the operator locally (applies CRDs and runs with local kubeconfig)
- `make deploy` - Deploy operator to cluster (includes building image, deploying catalog, and installing)

### Testing
- `make test-unit` - Run unit tests (includes API and internal package tests)
- `make test-functional` - Run functional tests against collector components
- `make test-e2e` - Run end-to-end tests (requires cluster)
- `make test-e2e-local` - Run e2e tests with locally built image
- `make test-cluster` - Run cluster-level tests
- `go test ./path/to/specific/package` - Run tests for a specific package
- `env $(make -s test-env) go test ./my/packages` - Run tests with proper environment setup

### Development Workflow
- `make pre-commit` - Run full validation before committing (clean, bundle, check, docs)
- `make check` - Health check including code generation, build, lint, and unit tests
- `make lint` - Run golangci-lint
- `make clean` - Clean build artifacts

### Environment Setup
- `make test-env` - Print environment variables needed for tests
- `make tools` - Install required development tools

## Architecture

### Core Components

**Controllers**: Located in `internal/controller/`
- `observability/clusterlogforwarder_controller.go` - Main reconciliation logic for ClusterLogForwarder CRD
- `logfilemetricsexporter/logfilemetricsexporter_controller.go` - LogFileMetricsExporter controller
- Handles resource initialization, secret/configmap mapping, validation, and deployment orchestration

**API Definitions**: Located in `api/observability/v1/`
- `clusterlogforwarder_types.go` - ClusterLogForwarder CRD schema
- `output_types.go` - Output type specifications for various log destinations
- Supports complex log routing via inputs, outputs, pipelines, and filters

**Configuration Generation**: Located in `internal/generator/`
- Translates ClusterLogForwarder specs into Vector collector configurations
- Template-based system for generating configuration fragments
- Uses internal data model to support multiple output formats while maintaining consistency

### Data Flow Architecture

Logs flow through the collector in this pattern:
```
collect from source → move to ._internal → transform → apply output datamodel → apply sink changes → send
```

The minimal attributes required for processing:
- log_type, log_source, timestamp, message, level, kubernetes_metadata

### Log Collection

**Supported Log Types**:
- Application logs (from regular pods)
- Infrastructure logs (from system pods and nodes)
- Audit logs (special node logs with legal/security implications)

**Current Implementation**: Vector as the log collector/forwarder

## Adding New Output Types

When adding support for a new log destination:

1. **Update API types**:
   - Add output type name to validation comment in `clusterlogforwarder_types.go:OutputSpec.Type`
   - Add constant and struct (if needed) to `output_types.go:OutputTypeSpec`

2. **Implement configuration generation**:
   - Create output package in `internal/generator/vector/output/[your_output]/`
   - Implement types satisfying the Element interface with Go templates
   - Add entry point function to `internal/generator/vector/outputs.go`
   - Add unit tests for template validation

3. **Add functional tests**:
   - Create `test/functional/outputs/[your_output]_test.go`
   - Verify output can connect and forward logs

Reference existing outputs like Kafka or CloudWatch as examples.

## Key Directories

- `api/` - API definitions and CRD schemas
- `cmd/` - Main entry point for the operator
- `config/` - Kubernetes configuration manifests
- `internal/controller/` - Reconciliation logic
- `internal/generator/` - Configuration generation system
- `test/` - All testing infrastructure (unit, functional, e2e)
- `hack/` - Development and build scripts
- `docs/` - Documentation including administration and contributing guides

## Development Notes

- Uses Operator SDK framework with controller-runtime
- Kustomize for Kubernetes manifest management
- Go modules for dependency management
- Supports Hosted Control Plane (HCP) compatibility
- Integrates with loki-operator for log storage
- Provides Grafana dashboard generation and Prometheus metrics

## Testing Strategy

The project uses a multi-layered testing approach:
- **Unit tests**: Test individual components and template generation
- **Functional tests**: Test collector component integration
- **E2E tests**: Test full operator functionality in real cluster environments
- **Benchmarker tests**: Performance validation for Vector collector

## Environment Variables

Key environment variables for development:
- `LOG_LEVEL` - Controls logging verbosity (default: 9)
- `KUBECONFIG` - Path to kubeconfig file (default: ~/.kube/config)
- `NAMESPACE` - Target namespace for deployments (default: openshift-logging)
- `IMAGE_TAG` - Container image tag for local builds
