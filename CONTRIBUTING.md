# Contributing to Cluster Logging Operator

Thank you for your interest in contributing! This document provides guidelines for submitting changes to the Cluster Logging Operator project.

## Getting Started

1. **Fork and Clone**
   ```bash
   git clone https://github.com/yourusername/cluster-logging-operator.git
   cd cluster-logging-operator
   ```

2. **Install Dependencies**
   ```bash
   make tools
   ```

3. **Create a Branch**
   ```bash
   git checkout -b feature/your-feature-name
   ```

## Development Workflow

### Before Making Changes

- Check existing [issues](https://github.com/openshift/cluster-logging-operator/issues) to avoid duplicate work
- For significant changes, open an issue to discuss the approach first
- Review the [ARCHITECTURE.md](ARCHITECTURE.md) to understand the design

### Making Changes

1. **Write Your Code**
   - Follow Go conventions and style guidelines
   - Keep changes focused and minimal
   - Add tests for new functionality

2. **Run Tests Locally**
   ```bash
   # Unit tests
   make test-unit

   # All validation
   make check

   # Functional tests (requires cluster)
   make test-functional

   # E2E tests (requires cluster)
   make test-e2e-local
   ```

3. **Code Quality Checks**
   ```bash
   # Run pre-commit validation
   make pre-commit

   # Run linter
   make lint
   ```

## Submitting a Pull Request

### PR Requirements

1. **PR Title and Description**
   - Use clear, descriptive titles
   - Start with the Jira issue number if applicable (e.g., "LOG-1234: Add feature X")
   - Provide context on what the change does and why it's needed

2. **Code Changes**
   - Keep PRs focused on a single feature or bug fix
   - Update related documentation
   - Include tests for new functionality
   - Ensure all tests pass locally before pushing

3. **Commit Messages**
   - Use clear, descriptive commit messages
   - Reference related issues when applicable
   - Keep commits atomic and logical

### CI/CD Pipeline

The repository uses automated CI/CD:
- **Unit Tests** - Run on every PR
- **Linting** - Code style validation
- **Build Validation** - Ensures code compiles
- **Functional Tests** - Tests component integration
- **E2E Tests** - Tests full operator functionality in a cluster

All checks must pass before merging. Don't be discouraged if CI catches issues—this is normal. Push fixes and the CI will re-run automatically.

### Review Process

- A maintainer will review your PR
- Address feedback and push updates as needed
- Be patient—reviews may take time due to other priorities
- For urgent changes, mention it in the PR description

## Testing Requirements

### Unit Tests

Unit tests are required for:
- New API types and validations
- Configuration generation logic
- Controller reconciliation logic

Run with:
```bash
make test-unit
```

### Functional Tests

Functional tests verify output connector integration:
```bash
make test-functional
```

### E2E Tests

End-to-end tests require a Kubernetes cluster:
```bash
make test-e2e-local
```

## Coding Conventions

### Go Style

- Follow [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- Use `gofmt` for formatting
- Run `make lint` before committing

### API Changes

When modifying CRD types in `api/`:
1. Update the struct tags and comments
2. Run `make generate` to update generated code
3. Add tests for validation logic
4. Update relevant documentation

### Configuration Generation

When adding new output types:
1. Add type definition to `api/observability/v1/output_types.go`
2. Create generator in `internal/generator/vector/output/[type]/`
3. Implement the Element interface with Go templates
4. Add entry point to `internal/generator/vector/outputs.go`
5. Add functional tests in `test/functional/outputs/`

See [docs/contributing/how-to-add-new-output.md](docs/contributing/how-to-add-new-output.md) for detailed examples.

## Adding New Output Types

Comprehensive guide: [How to Add a New Output Type](docs/contributing/how-to-add-new-output.md)

Quick summary:
1. Add API type definitions
2. Implement configuration generation with templates
3. Add functional tests to verify connectivity and log forwarding
4. Update documentation

## Documentation

### When to Update Docs

- New features should include documentation
- API changes need corresponding doc updates
- Complex logic should have implementation notes
- Architecture changes should be documented in ARCHITECTURE.md

### Documentation Locations

- **User Guides**: `/docs/administration/`
- **Developer Guides**: `/docs/contributing/`
- **Architecture**: `/docs/architecture/`
- **Feature Docs**: `/docs/features/`
- **API Reference**: `/docs/reference/`

## Project Structure

For navigation tips, see [ARCHITECTURE.md](ARCHITECTURE.md#key-directories).

## Common Issues

### Tests Failing

1. **Unit tests fail**: Check for linting errors with `make lint`
2. **Functional tests fail**: Verify cluster connectivity and required permissions
3. **E2E tests fail**: Ensure clean cluster state with `make undeploy-all`

### Build Issues

- Run `make clean` then `make build` to start fresh
- Ensure Go version is 1.24 or later
- Check that all dependencies are installed: `make tools`

## Development Tools

### IDE Setup

This project works with:
- GoLand / IntelliJ IDEA with Go plugin
- VS Code with Go extension
- Vim/Neovim with gopls

### Debugging

Run the operator in debug mode:
```bash
make run-debug
```

This starts the operator under the `dlv` debugger.

## Reporting Issues

When reporting bugs:
1. Search existing issues first
2. Provide:
   - Steps to reproduce
   - Expected vs actual behavior
   - OpenShift version
   - CLO version
   - Relevant logs or error messages

## Code Review Guidelines

See [docs/contributing/REVIEW.adoc](docs/contributing/REVIEW.adoc) for detailed review guidelines.

## Questions?

- Check [docs/](docs/) for detailed information
- Open a [discussion](https://github.com/openshift/cluster-logging-operator/discussions)
- Create an [issue](https://github.com/openshift/cluster-logging-operator/issues) with a question label

## License

By contributing, you agree that your contributions will be licensed under the Apache License 2.0.
