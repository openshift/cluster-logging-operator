---
name: clo-otel-migration
description: Working on proposal to migrate CLO from Vector collector to OpenTelemetry collector - milestone 1 replaces Vector binary with OTEL collector
metadata:
  type: project
---

User is working on a migration proposal to replace Vector with the OpenTelemetry collector in the Cluster Logging Operator (CLO). 

**Why:** CLO currently uses Vector (Datadog) as the log collection/forwarding engine. The goal is to migrate to the OTEL collector to align with OpenTelemetry standards and potentially consolidate with the OpenTelemetry Operator.

**How to apply:** When working on CLO code or the migration proposal, understand that:
- Milestone 1: Replace Vector binary with OTEL collector in CLO-managed pods, generate OTEL collector YAML config instead of Vector TOML
- Milestone 2: Use OTEL collector CR instead of CLO-managed deployment
- Milestone 3: Full migration to OpenTelemetry Operator
- The ClusterLogForwarder CRD API stays the same in milestone 1
- Key files: api/observability/v1/ (CRD types), internal/generator/vector/ (Vector config generation)
