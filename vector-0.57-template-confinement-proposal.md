# Vector v0.57.0 template confinement breaks S3/Kafka/ES/CloudWatch/Splunk/Loki/GCL output generation

**Issue Type:** Bug (blocks upgrade) — consider splitting into sub-tasks per the implementation steps below
**Priority:** Blocker (for the v0.57 upgrade epic)
**Affects Version:** Vector v0.57.0 (branch `test-vector-v0.57.0`)

## Problem

Vector 0.57.0 introduced **template confinement**. Sink routing templates (object keys, topics, indices, file paths, headers, table/stream names) must now resolve within a boundary derived from a literal string prefix in the template. Templates with no literal prefix are rejected at startup. This is a security fix: fully dynamic templates like `"{{ host }}/"` let event-field values determine write locations with no constraint (e.g. a crafted value like `"../../../etc/shadow"` could redirect a write outside the intended location).

Escape hatch: `dangerously_allow_unconfined_template_resolution: true`, set **per-sink**, restores pre-0.57 behavior for that sink.

**Impact if unaddressed:** every CLO output using S3, Kafka, Elasticsearch, CloudWatch, Splunk, Loki, or GCL fails config validation at collector startup on upgrade to v0.57 → crash loop, total outage for those pipelines, with no customer config change required to trigger it.

## Current state (verified against the real v0.57.0 binary via `vector validate`)

10 template sites across 7 output sinks fail, all for the same reason: they reference an event field with zero literal text in front of it.

| Sink | Field | Template |
|---|---|---|
| S3 | key_prefix | `{{ _internal.<id> }}` |
| Kafka | topic | `{{ _internal.<id> }}` |
| Elasticsearch | index | `{{ _internal.<id> }}` |
| CloudWatch | group_name | `{{ _internal.cw_group_name }}` |
| CloudWatch | stream_name | `{{ stream_name }}` |
| Splunk | index | `{{ ._internal.<id> }}` |
| Splunk | source / sourcetype | `{{ ._internal.splunk.source/sourcetype }}` |
| Loki | tenant / labels | `{{ _internal.<tenant> }}`, `{{kubernetes.container_name}}`, etc. |
| GCL | log_id | `{{ _internal.<id> }}` |
| GCL | resource.node_name | `{{hostname}}` |

These split into two groups:

*   **Category A — driven by user-configurable CR templates** (S3 `KeyPrefix`, Kafka `Topic`, ES `Index`, CloudWatch `GroupName`, Splunk `Index`, Loki `TenantKey`, GCL `LogId`). All route through a shared helper, `NewTemplateRemap()`, which compiles the customer's CR-supplied template string into VRL and computes the entire final value into an internal field. The sink then points at that field with no literal.
*   **Category B — CLO-generated defaults, not user-configurable** (CloudWatch `stream_name`, Splunk `source`, GCL `node_name`, Loki's default labels). These are fixed, multi-branch VRL logic built from Kubernetes metadata — there is no literal anywhere to hoist.

Also noted in passing: Splunk's default `sourcetype` is always the hardcoded constant `"_json"`. It's routed through a template/remap for no reason; this is cheap to clean up alongside the main fix.

## Risk analysis & runtime behavior

Vector's runtime confinement is generic across sink types and enforcement on violation is **drop-only**. There is no reroute, fallback, or dead-letter option on confinement failure. This means doing nothing guarantees a crash loop, but relying entirely on Vector's native confinement risks silent data loss (dropped logs) when an edge case is hit — a worse failure mode than usual for a logging product, since dropped events are exactly the audit/compliance trail customers rely on.

Verified this directly against the real v0.57.0 binary (not just `vector validate`, which only checks syntax):

- Kafka `topic = "app-{{ _internal.evil }}"` with `_internal.evil = "../../secret-tenant-topic"` (literal prefix present) → **event dropped**: `rendered value "app-../../secret-tenant-topic" contains a ".." path segment`, `error_type=confinement_failed`. No config option exists to reroute or fall back instead of dropping (checked the binary's strings; the `remap` transform's `reroute_dropped` is unrelated — it covers VRL script errors, not sink-level confinement).
- Same style of template with a *legitimate* value, `"team-a/service-b"` → delivered successfully as a nested path. Confinement only blocks upward traversal (`..`), not plain `/` nesting.

CLO currently has a `replace(value, r'[\./]', "_")` pattern already in use elsewhere, but it only sanitizes **field key names** (label/indexed-field flattening), not the **values** flowing into routing templates. Traced the actual value computation (shared by all Category A sinks) — it's a plain string concatenation with no sanitization. So today, raw values — dots, slashes, and all — flow directly into the routing templates.

## Proposal

**(e) Narrow value sanitization — primary defense to prevent data loss.**
Extend `TransformUserTemplateToVRL` (`template.go`) so every dynamic segment substituted into a Category A template has `..` path segments and a leading `/` stripped/neutralized before assignment to `_internal.<field>`. This is narrower than a blanket dot/slash strip, allowing legitimate dots (IPs, `v1.2.3`) and legitimate nesting (`team-a/service-b`) to pass through untouched — confirmed both patterns are safe to allow per the runtime test above. Because Vector's only response to a violation is dropping the event, preventing the violation upstream is the only way to guarantee no log loss from this failure mode.

**(a) Category A literal hoisting — secondary backstop.**
When the CR template has a literal prefix (e.g. `keyPrefix: "app-{.kubernetes.namespace_name}"`), move that literal out of the VRL computation and into the sink-level Vector template itself: `key_prefix = "app-{{ _internal.key_prefix }}"`. Verified this produces a byte-identical rendered value to the current approach and passes `vector validate`. With (e) in place, this shouldn't trigger a drop in practice, but acts as a second layer of defense in case the sanitization in (e) has a bug or misses a case.

**(b) Category A, zero-literal case — fallback mechanism.**
For CR templates with no literal at all (e.g., `{.log_type||.log_source||"missing"}/` — an example straight from our own `S3.KeyPrefix` CRD godoc, not hypothetical), there's no literal for Vector to derive a confinement base from, so we fall back to `dangerously_allow_unconfined_template_resolution: true` per-instance. Because we're sanitizing values upstream via (e), this is safe from the same path/key-traversal injection class this ticket addresses — it does not, by itself, protect against unrelated concerns like cardinality-driven resource exhaustion from an unbounded dynamic field. This state must be surfaced to users via a status condition on the output CR.

**(c) Category B — scoped escape hatch + sanitization.**
Apply `dangerously_allow_unconfined_template_resolution: true` to the 4 specific Category B fields. Apply (e)'s narrow sanitization to these fields as well to future-proof against edge cases, even though these fields are largely constrained by Kubernetes DNS-1123 validation.

**(d) Splunk sourcetype cleanup.**
Drop the template/remap round-trip entirely and hardcode `sourcetype = "_json"` in Go when no custom `SourceType` is configured.

## Resolved questions (from spike review)

1.  **Sanitizing Category B:** Yes, apply the narrow `..` and leading `/` sanitization to Category B. It is cheap and provides consistency.
2.  **Revisiting Category B default naming:** Breaking default naming schemes impacts customer dashboards and alarms. This will be spun out into a separate epic for a future release and will not block the v0.57.0 upgrade.
3.  **Auditing zero-literal templates:** A telemetry audit of customer configurations should be performed. If usage is extremely low, we can deprecate zero-literal templates in a future API version. *Needs confirmation: does existing telemetry actually capture template shape at this granularity, or does this require new instrumentation first?*
4.  **Alerting on confinement failures:** Yes. Even with upstream sanitization, we need observability. We will add the Vector `confinement_failed` metric to standard Grafana dashboards and Prometheus alerts.

## Implementation steps

1.  **Implement primary defense (e):** Add narrow `..`/leading-`/` neutralization in `template.go` (`TransformUserTemplateToVRL`). *Requirement: Include unit tests explicitly testing IP addresses (`192.168.1.1`), versions (`v1.2.3`), and valid nested paths to ensure they are not mangled.*
2.  **Implement secondary backstop (a):** Wire the literal-prefix hoist through the 7 Category A sinks.
3.  **Implement Category B fixes (c):** Apply scoped `dangerously_allow_unconfined_template_resolution: true` on the 4 Category B fields, AND apply the (e) sanitization to them.
4.  **Implement Splunk cleanup (d):** Simplify the default sourcetype.
5.  **Add visibility (b):** Implement a status condition on the output CR (e.g., `UnconfinedTemplateResolution: True`) with a warning message when the zero-literal fallback is used.
6.  **Add functional tests:** Validate generated configs against the real Vector v0.57.0 binary to catch confinement regressions, including a test that a malicious `..`-bearing value never triggers a dropped event.
7.  **Add observability:** Expose the `confinement_failed` and `security_confinement_disabled` metrics on default Grafana dashboards and add a Prometheus alert for `confinement_failed > 0`.

## Suggested labels

`vector-upgrade`, `security`, `regression-risk`
