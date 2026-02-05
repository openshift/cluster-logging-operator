package v1

import (
	"fmt"
	"regexp"
	"strings"

	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
)

const (
	// Default field names per W3C Trace Context spec
	DefaultTraceIdFieldName    = "trace_id"
	DefaultSpanIdFieldName     = "span_id"
	DefaultTraceFlagsFieldName = "trace_flags"

	// Value patterns per W3C Trace Context spec
	// https://www.w3.org/TR/trace-context/
	// - trace_id: 32 lowercase hex characters (128-bit)
	// - span_id: 16 lowercase hex characters (64-bit)
	// - trace_flags: 1-2 lowercase hex characters (8-bit)
	TraceIdValuePattern    = `[0-9a-fA-F]{1,32}`
	SpanIdValuePattern     = `[0-9a-fA-F]{1,16}`
	TraceFlagsValuePattern = `[0-9a-fA-F]{1,2}`
)

// buildPattern creates a regex pattern using field name alternation with a single capture group.
// Example: (trace_id|traceId|traceID)[=:]\s*["']?(?<trace_id>[0-9a-fA-F]{32})["']?
func buildPattern(captureName string, defaultFieldName string, additionalNames []string, valuePattern string) string {
	allFieldNames := append([]string{defaultFieldName}, additionalNames...)

	// Escape special regex characters in field names (e.g., dots in "trace.id")
	escapedNames := make([]string, len(allFieldNames))
	for i, name := range allFieldNames {
		escapedNames[i] = regexp.QuoteMeta(name)
	}

	// Build field name alternation (e.g., trace_id|traceId|traceID)
	var fieldNamePattern string
	if len(escapedNames) == 1 {
		fieldNamePattern = escapedNames[0]
	} else {
		fieldNamePattern = "(" + strings.Join(escapedNames, "|") + ")"
	}

	// Build full pattern with single capture group
	// Matches: field=value, field="value", field='value', field: value
	return fmt.Sprintf(`%s[=:]\s*["\']?(?<%s>%s)["\']?`, fieldNamePattern, captureName, valuePattern)
}

// GenerateAddLogRecordTraceContexts generates VRL code for extracting trace context from log messages.
// If custom patterns are defined, only those are used. Otherwise, field name and value matching per W3C spec is used.
// https://opentelemetry.io/docs/specs/otel/compatibility/logging_trace_context/
func GenerateAddLogRecordTraceContexts(traceContext *obs.OtlpTraceContextSpec) string {
	var additionalTraceIds, additionalSpanIds, additionalFlags, customPatterns []string
	if traceContext != nil {
		additionalTraceIds = traceContext.AdditionalTraceIdFieldNames
		additionalSpanIds = traceContext.AdditionalSpanIdFieldNames
		additionalFlags = traceContext.AdditionalTraceFlagsFieldNames
		customPatterns = traceContext.CustomPatterns
	}

	var sb strings.Builder

	// If custom patterns are defined, use only those (skip defaults and custom field name matching)
	if len(customPatterns) > 0 {
		sb.WriteString(`
# Extract trace context using custom patterns
`)
		for i, pattern := range customPatterns {
			sb.WriteString(fmt.Sprintf(`
# Custom trace context pattern %d

parsed, err = parse_regex(._internal.message, r'%s')
if err == null {
	if exists(parsed.trace_id) {
		# Buffer trace_id with leading zeros to 32 hex chars per W3C spec
		buffered_trace_id = "00000000000000000000000000000000" + downcase(string!(parsed.trace_id))
		._internal.trace_id = slice!(buffered_trace_id, -32)
	}
	if exists(parsed.span_id) {
		# Buffer span_id with leading zeros to 16 hex chars per W3C spec
		buffered_span_id = "0000000000000000" + downcase(string!(parsed.span_id))
		._internal.span_id = slice!(buffered_span_id, -16)
	}
	if exists(parsed.trace_flags) {
		._internal.trace_flags = parsed.trace_flags
	}
}

`, i+1, pattern))
		}
		return sb.String()
	}

	// Default: use field name matching but keep value matching per W3C spec
	traceIdRegex := buildPattern(DefaultTraceIdFieldName, DefaultTraceIdFieldName, additionalTraceIds, TraceIdValuePattern)
	spanIdRegex := buildPattern(DefaultSpanIdFieldName, DefaultSpanIdFieldName, additionalSpanIds, SpanIdValuePattern)
	flagsRegex := buildPattern(DefaultTraceFlagsFieldName, DefaultTraceFlagsFieldName, additionalFlags, TraceFlagsValuePattern)

	sb.WriteString(fmt.Sprintf(`
# Extract trace context using field name matching
# Extract trace_id padded to 32 hex chars per W3C spec
parsed, err = parse_regex(._internal.message, r'%s')
if err == null && exists(parsed.trace_id) {
  # Buffer trace_id with leading zeros to 32 hex chars per W3C spec
  buffered_trace_id = "00000000000000000000000000000000" + downcase(string!(parsed.trace_id))
  ._internal.trace_id = slice!(buffered_trace_id, -32)
}

# Extract span_id padded to 16 hex chars per W3C spec
parsed, err = parse_regex(._internal.message, r'%s')
if err == null && exists(parsed.span_id) {
  # Buffer span_id with leading zeros to 16 hex chars per W3C spec
  buffered_span_id = "0000000000000000" + downcase(string!(parsed.span_id))
  ._internal.span_id = slice!(buffered_span_id, -16)
}

# Extract trace_flags
parsed, err = parse_regex(._internal.message, r'%s')
if err == null && exists(parsed.trace_flags) {
  ._internal.trace_flags = parsed.trace_flags
}
`, traceIdRegex, spanIdRegex, flagsRegex))

	return sb.String()
}
