package v1

// AddLogRecordTraceContexts is the VRL code for extracting trace context from log messages.
// It uses a three-step extraction strategy:
//  1. Check structured log fields directly
//  2. Parse the message as JSON
//  3. Fall back to regex matching with field name alternation per W3C Trace Context spec
//
// References:
//   - https://www.w3.org/TR/trace-context/
//   - https://opentelemetry.io/docs/specs/otel/compatibility/logging_trace_context/
const AddLogRecordTraceContexts = `
trace_context = {}
# 1. Try to extract trace context from structured log fields
if exists(._internal.structured) {
	if exists(._internal.structured.trace_id) {
		trace_context.trace_id = ._internal.structured.trace_id
	}
	if exists(._internal.structured.span_id) {
		trace_context.span_id = ._internal.structured.span_id
	}
	if exists(._internal.structured.trace_flags) {
		trace_context.trace_flags = ._internal.structured.trace_flags
	}
}

# 2. If not structured, try parsing the message as JSON
if !exists(._internal.structured) {
	parsed, err = parse_json(._internal.message)
	if err == null {
		if exists(parsed.trace_id) {
			trace_context.trace_id = parsed.trace_id
		}
		if exists(parsed.span_id) {
			trace_context.span_id = parsed.span_id
		}
		if exists(parsed.trace_flags) {
			trace_context.trace_flags = parsed.trace_flags
		}
	}
}

# 3. Fall back to regex for any fields still missing
if trace_context.trace_id == null {
	parsed, err = parse_regex(._internal.message, r'(?i)(trace_id|traceId|traceID|trace\-id|trace\.id)[=:]\s*["\']?(?<trace_id>[0-9a-f]{32})["\']?')
	if err == null && exists(parsed.trace_id) {
		trace_context.trace_id = parsed.trace_id
	}
}
if trace_context.span_id == null {
	parsed, err = parse_regex(._internal.message, r'(?i)(span_id|spanId|spanID|span\-id|span\.id)[=:]\s*["\']?(?<span_id>[0-9a-f]{16})["\']?')
	if err == null && exists(parsed.span_id) {
		trace_context.span_id = parsed.span_id
	}
}
if trace_context.trace_flags == null {
	parsed, err = parse_regex(._internal.message, r'(?i)(trace_flags|traceFlags|flags|trace\-flags|trace\.flags)[=:]\s*["\']?(?<trace_flags>[0-9a-f]{1,2})["\']?')
	if err == null && exists(parsed.trace_flags) {
		trace_context.trace_flags = parsed.trace_flags
	}
}

# 4. Validate and set each trace context field
if trace_context.trace_id != null {
	trace_id_str = downcase(to_string!(trace_context.trace_id))
	if match(trace_id_str, r'^[0-9a-f]{32}$') {
		._internal.trace_id = trace_id_str
	}
}
if trace_context.span_id != null {
	span_id_str = downcase(to_string!(trace_context.span_id))
	if match(span_id_str, r'^[0-9a-f]{16}$') {
		._internal.span_id = span_id_str
	}
}
if trace_context.trace_flags != null {
	trace_flags_str = downcase(to_string!(trace_context.trace_flags))
	if match(trace_flags_str, r'^0?[01]$') {
		._internal.trace_flags = trace_flags_str
	}
}
`
