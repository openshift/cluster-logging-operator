# Cluster Logging Operator

The Cluster Logging Operator (CLO) is a Kubernetes operator for configuring log collection and forwarding on OpenShift clusters. It manages the deployment of Vector as the log collector and supports complex log routing through inputs, outputs, filters, and pipelines defined via the ClusterLogForwarder CRD.

## What This Repository Does

This repository contains the source code and configuration for the Cluster Logging Operator, which:

- Collects logs from multiple sources (application, infrastructure, and audit logs)
- Transforms and filters logs using configurable pipelines
- Forwards logs to various destinations (Loki, Splunk, AWS CloudWatch, Google Cloud Logging, Elasticsearch, Kafka, Azure, and more)
- Manages log collection through Kubernetes Custom Resources
- Provides metrics and observability for the logging infrastructure

## Building and Running Locally

### Prerequisites

- Go (see `go.mod` for the required version)
- Podman or Docker
- Kubernetes cluster (or local cluster like Kind, minikube, or Code Ready Containers)
- kubeconfig configured for your target cluster

### Quick Start

```bash
# Install development tools
make tools

# Build the operator binary
make build

# Run the operator locally (requires kubeconfig)
make run

# Or deploy to a cluster
make deploy
```

For a full list of development commands and workflows, see [CONTRIBUTING.md](CONTRIBUTING.md).

## Directory Structure

- `api/` - API definitions and CRD schemas
- `cmd/` - Main entry point for the operator
- `config/` - Kubernetes configuration manifests
- `internal/controller/` - Reconciliation logic
- `internal/generator/` - Configuration generation system for Vector
- `test/` - All testing infrastructure (unit, functional, e2e)
- `hack/` - Development and build scripts
- `docs/` - Detailed documentation including administration and contribution guides

## Architecture

The operator uses a multi-layered architecture:

- **Controllers**: Manage ClusterLogForwarder and LogFileMetricsExporter resources
- **Configuration Generator**: Translates CRD specs into Vector collector configurations
- **Vector Integration**: Uses Vector as the log collector/forwarder
- **Kubernetes Integration**: Deploys and manages collector components via DaemonSets and Deployments

For detailed architecture information, see [ARCHITECTURE.md](ARCHITECTURE.md).

## Documentation

- **[CONTRIBUTING.md](CONTRIBUTING.md)** - How to submit changes and development guidelines
- **[ARCHITECTURE.md](ARCHITECTURE.md)** - Design decisions, dependency points, and architecture details
- **[AGENTS.md](AGENTS.md)** - AI agent instructions and conventions
- **[docs/](docs/)** - Detailed developer documentation including:
  - [Administration Guides](docs/administration/)
  - [Architecture Overview](docs/architecture/)
  - [Contributing Guidelines](docs/contributing/)
  - [Feature Documentation](docs/features/)
  - [API Reference](docs/reference/)

For official OpenShift Logging documentation, see the [OpenShift Container Platform documentation](https://docs.redhat.com/en/documentation/openshift_container_platform/latest/html/logging).

## Contributing

We welcome contributions! Please see [CONTRIBUTING.md](CONTRIBUTING.md) for information on how to:
- Submit changes via pull requests
- Follow coding conventions
- Run tests before submitting
- Participate in code reviews

## License

This repository is licensed under the Apache License 2.0. See [LICENSE](LICENSE) file for details.

## Getting Help

- Open an issue in this repository for bugs, feature requests, or documentation problems
- Check existing [issues](https://redhat.atlassian.net/browse/LOG)
- Visit the [OpenShift Logging documentation](https://docs.redhat.com/en/documentation/openshift_container_platform/latest/html/logging) for user-facing information
