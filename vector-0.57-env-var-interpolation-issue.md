# Vector v0.57.0 disables ${VAR} env-var interpolation by default, breaking OPENSHIFT_CLUSTER_ID / VECTOR_SELF_NODE_NAME resolution

**Issue Type:** Bug
**Priority:** Blocker (for the v0.57 upgrade epic)
**Affects Version:** Vector v0.57.0 (branch `test-vector-v0.57.0`)

## Problem

Vector 0.57.0 changed environment-variable interpolation behavior: https://vector.dev/highlights/2026-07-14-0-57-0-upgrade-guide/#env-var-interpolation

`${VAR}`-style substitution in config files is now **disabled by default** — a global, process-wide change, not scoped to a sink or component. The old opt-out mechanism (`--disable-env-var-interpolation` / `VECTOR_DISABLE_ENV_VAR_INTERPOLATION`) was removed entirely, not just flipped. To restore interpolation, Vector now requires `--dangerously-allow-env-var-interpolation` on the command line, or `VECTOR_DANGEROUSLY_ALLOW_ENV_VAR_INTERPOLATION=true`.

This is a **separate v0.57 change from template confinement** (see the other ticket) — different mechanism entirely: this is a config-load-time text substitution pass over `${VAR}` syntax, not Vector's own `{{ }}` template/confinement system.

## How this surfaced

Functional test failure:
```text
[FAILED] [Functional][OutputConditions][Syslog] RFC5424: should allow combination of
static + dynamic setting of appname, procid, messageid from record

Expected <string>: foo-${OPENSHIFT_CLUSTER_ID:-}
to equal <string>: foo-functional
```
The `${OPENSHIFT_CLUSTER_ID:-}` env-var reference was staying literal in the rendered output instead of resolving to the actual cluster ID.

## Root cause (verified against the real v0.57.0 binary)

CLO uses `${VAR}` interpolation syntax in exactly two places in generated configs:

| Location | Usage |
|---|---|
| `internal/generator/vector/input/internal.go:30` (`setClusterID`) | `._internal.openshift = { "cluster_id": "${OPENSHIFT_CLUSTER_ID:-}"}` — embedded in a VRL `remap` source string, present in every input pipeline |
| `internal/generator/vector/output/loki/loki.go:171` | `"${VECTOR_SELF_NODE_NAME}"` — used for Loki's `k8s_node_name`/`kubernetes_host` labels |

Reproduced directly with the real v0.57.0 binary:
```bash
# without --dangerously-allow-env-var-interpolation
out: "${OPENSHIFT_CLUSTER_ID:-}"

# with --dangerously-allow-env-var-interpolation
out: "functional"
```

## Risk assessment

Both env vars are CLO-controlled and non-secret — set directly on the collector container spec, not derived from customer/user input:
- `OPENSHIFT_CLUSTER_ID` — set in `internal/collector/collector.go:239`
- `VECTOR_SELF_NODE_NAME` — set via downward API in `internal/collector/vector/visitors.go:16`

CLO doesn't rely on `${VAR}` syntax for any secrets or credentials — those already go through a separate, proper secrets mechanism (`api.NewDirectorySecret`, `internal/generator/vector/conf/global.go`). Since this flag is process-wide (unlike the confinement escape hatch, which is per-sink), there's no per-instance scoping question here — enabling it unconditionally on the collector process is low-risk given CLO's actual usage.

## Fix

Set the environment variable `VECTOR_DANGEROUSLY_ALLOW_ENV_VAR_INTERPOLATION=true` in the collector container specification.

1. **Production Collector:** In the Kubernetes manifest generation (likely around `internal/collector/collector.go` or `internal/collector/vector/visitors.go`), add the environment variable to the Vector container spec:
   ```go
   corev1.EnvVar{
       Name:  "VECTOR_DANGEROUSLY_ALLOW_ENV_VAR_INTERPOLATION",
       Value: "true",
   }
   ```
2. **Functional Tests:** Ensure the functional test harness also applies this configuration. This can be done by adding `--dangerously-allow-env-var-interpolation` to `internal/collector/vector/run-vector.sh` (shared by `vector.RunVectorScript`) or by exporting the environment variable in the test context.

Setting the environment variable on the container spec is cleaner than modifying the shell entrypoint and aligns better with Kubernetes operator patterns.

## Alternatives considered

The upgrade guide recommends, for cases involving actual secrets, switching to Vector's secrets backend instead of env-var interpolation. Not applicable here — neither `OPENSHIFT_CLUSTER_ID` nor `VECTOR_SELF_NODE_NAME` is a secret, and CLO's real secrets already use a separate mechanism. Pre-processing the config with `envsubst` (the guide's other suggestion) isn't viable either, since CLO uses the `${VAR:-default}` extended syntax, which `envsubst` doesn't support.

## Suggested labels

`vector-upgrade`, `regression`
