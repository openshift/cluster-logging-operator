package v1

import (
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
)

var _ = Describe("buildPattern", func() {
	It("should build a simple pattern with single field name", func() {
		pattern := buildPattern("trace_id", "trace_id", nil, TraceIdValuePattern)
		Expect(pattern).To(Equal(`trace_id[=:]\s*["\']?(?<trace_id>[0-9a-fA-F]{1,32})["\']?`))
	})

	It("should build a pattern with field name alternation", func() {
		pattern := buildPattern("trace_id", "trace_id", []string{"traceId", "traceID"}, TraceIdValuePattern)
		Expect(pattern).To(Equal(`(trace_id|traceId|traceID)[=:]\s*["\']?(?<trace_id>[0-9a-fA-F]{1,32})["\']?`))
	})

	It("should escape special regex characters in field names", func() {
		pattern := buildPattern("trace_id", "trace.id", nil, TraceIdValuePattern)
		Expect(pattern).To(Equal(`trace\.id[=:]\s*["\']?(?<trace_id>[0-9a-fA-F]{1,32})["\']?`))
	})

	It("should build span_id pattern correctly", func() {
		pattern := buildPattern("span_id", "span_id", []string{"spanId"}, SpanIdValuePattern)
		Expect(pattern).To(Equal(`(span_id|spanId)[=:]\s*["\']?(?<span_id>[0-9a-fA-F]{1,16})["\']?`))
	})

	It("should build trace_flags pattern correctly", func() {
		pattern := buildPattern("trace_flags", "trace_flags", nil, TraceFlagsValuePattern)
		Expect(pattern).To(Equal(`trace_flags[=:]\s*["\']?(?<trace_flags>[0-9a-fA-F]{1,2})["\']?`))
	})
})

var _ = Describe("GenerateAddLogRecordTraceContexts", func() {
	Context("with nil trace context spec", func() {
		It("should generate default field name matching VRL", func() {
			vrl := GenerateAddLogRecordTraceContexts(nil)

			// Should contain default trace_id pattern
			Expect(vrl).To(ContainSubstring(`trace_id[=:]\s*["\']?(?<trace_id>[0-9a-fA-F]{1,32})["\']?`))
			// Should contain default span_id pattern
			Expect(vrl).To(ContainSubstring(`span_id[=:]\s*["\']?(?<span_id>[0-9a-fA-F]{1,16})["\']?`))
			// Should contain default trace_flags pattern
			Expect(vrl).To(ContainSubstring(`trace_flags[=:]\s*["\']?(?<trace_flags>[0-9a-fA-F]{1,2})["\']?`))
			// Should contain padding logic comments
			Expect(vrl).To(ContainSubstring("# Buffer trace_id with leading zeros to 32 hex chars per W3C spec"))
			Expect(vrl).To(ContainSubstring("# Buffer span_id with leading zeros to 16 hex chars per W3C spec"))
		})
	})

	Context("with additional field names", func() {
		It("should generate VRL with field name alternation for trace_id", func() {
			spec := &obs.OtlpTraceContextSpec{
				AdditionalTraceIdFieldNames: []string{"traceId", "traceID"},
			}
			vrl := GenerateAddLogRecordTraceContexts(spec)

			// Should contain alternation pattern
			Expect(vrl).To(ContainSubstring(`(trace_id|traceId|traceID)[=:]\s*["\']?(?<trace_id>[0-9a-fA-F]{1,32})["\']?`))
		})

		It("should generate VRL with field name alternation for span_id", func() {
			spec := &obs.OtlpTraceContextSpec{
				AdditionalSpanIdFieldNames: []string{"spanId", "spanID"},
			}
			vrl := GenerateAddLogRecordTraceContexts(spec)

			// Should contain alternation pattern
			Expect(vrl).To(ContainSubstring(`(span_id|spanId|spanID)[=:]\s*["\']?(?<span_id>[0-9a-fA-F]{1,16})["\']?`))
		})

		It("should generate VRL with field name alternation for trace_flags", func() {
			spec := &obs.OtlpTraceContextSpec{
				AdditionalTraceFlagsFieldNames: []string{"traceFlags", "flags"},
			}
			vrl := GenerateAddLogRecordTraceContexts(spec)

			// Should contain alternation pattern
			Expect(vrl).To(ContainSubstring(`(trace_flags|traceFlags|flags)[=:]\s*["\']?(?<trace_flags>[0-9a-fA-F]{1,2})["\']?`))
		})

		It("should generate VRL with all additional field names", func() {
			spec := &obs.OtlpTraceContextSpec{
				AdditionalTraceIdFieldNames:    []string{"traceId"},
				AdditionalSpanIdFieldNames:     []string{"spanId"},
				AdditionalTraceFlagsFieldNames: []string{"flags"},
			}
			vrl := GenerateAddLogRecordTraceContexts(spec)

			Expect(vrl).To(ContainSubstring(`(trace_id|traceId)`))
			Expect(vrl).To(ContainSubstring(`(span_id|spanId)`))
			Expect(vrl).To(ContainSubstring(`(trace_flags|flags)`))
		})
	})

	Context("with custom patterns", func() {
		It("should use only custom patterns when defined", func() {
			spec := &obs.OtlpTraceContextSpec{
				CustomPatterns: []string{
					`traceparent[ ]*[=:][ ]*["\']?00-(?<trace_id>[0-9a-fA-F]{32})-(?<span_id>[0-9a-fA-F]{16})-(?<trace_flags>[0-9a-fA-F]{2})["\']?`,
				},
			}
			vrl := GenerateAddLogRecordTraceContexts(spec)

			// Should contain custom pattern comment
			Expect(vrl).To(ContainSubstring("# Extract trace context using custom patterns"))
			Expect(vrl).To(ContainSubstring("# Custom trace context pattern 1"))
			// Should contain the custom traceparent pattern
			Expect(vrl).To(ContainSubstring("traceparent"))
			// Should NOT contain default field name patterns
			Expect(vrl).NotTo(ContainSubstring("# Extract trace context using field name matching"))
		})

		It("should skip default patterns when custom patterns are defined with additional field names", func() {
			spec := &obs.OtlpTraceContextSpec{
				AdditionalTraceIdFieldNames: []string{"traceId"},
				CustomPatterns: []string{
					`traceparent[ ]*[=:][ ]*["\']?00-(?<trace_id>[0-9a-fA-F]{32})-(?<span_id>[0-9a-fA-F]{16})-(?<trace_flags>[0-9a-fA-F]{2})["\']?`,
				},
			}
			vrl := GenerateAddLogRecordTraceContexts(spec)

			// Should use custom patterns
			Expect(vrl).To(ContainSubstring("# Extract trace context using custom patterns"))
			// Should NOT use additional field names since custom patterns take precedence
			Expect(vrl).NotTo(ContainSubstring(`(trace_id|traceId)`))
		})

		It("should support multiple custom patterns", func() {
			spec := &obs.OtlpTraceContextSpec{
				CustomPatterns: []string{
					`pattern_one: (?<trace_id>[0-9a-fA-F]{32})`,
					`pattern_two: (?<span_id>[0-9a-fA-F]{16})`,
				},
			}
			vrl := GenerateAddLogRecordTraceContexts(spec)

			Expect(vrl).To(ContainSubstring("# Custom trace context pattern 1"))
			Expect(vrl).To(ContainSubstring("# Custom trace context pattern 2"))
			Expect(vrl).To(ContainSubstring("pattern_one"))
			Expect(vrl).To(ContainSubstring("pattern_two"))
		})

		It("should include padding logic in custom pattern handling", func() {
			spec := &obs.OtlpTraceContextSpec{
				CustomPatterns: []string{
					`my_trace=(?<trace_id>[0-9a-fA-F]{1,32})`,
				},
			}
			vrl := GenerateAddLogRecordTraceContexts(spec)

			// Should contain buffering/padding logic
			Expect(vrl).To(ContainSubstring(`buffered_trace_id = "00000000000000000000000000000000" + downcase(string!(parsed.trace_id))`))
			Expect(vrl).To(ContainSubstring(`._internal.trace_id = slice!(buffered_trace_id, -32)`))
		})
	})

	Context("VRL structure validation", func() {
		It("should generate valid VRL with error handling", func() {
			vrl := GenerateAddLogRecordTraceContexts(nil)

			// Should use parse_regex with error handling
			Expect(vrl).To(ContainSubstring("parsed, err = parse_regex"))
			Expect(vrl).To(ContainSubstring("if err == null"))
		})

		It("should store extracted values in ._internal namespace", func() {
			vrl := GenerateAddLogRecordTraceContexts(nil)

			Expect(vrl).To(ContainSubstring("._internal.trace_id"))
			Expect(vrl).To(ContainSubstring("._internal.span_id"))
			Expect(vrl).To(ContainSubstring("._internal.trace_flags"))
		})

		It("should downcase trace_id and span_id values", func() {
			vrl := GenerateAddLogRecordTraceContexts(nil)

			// Count occurrences of downcase for trace_id and span_id
			traceIdDowncase := strings.Count(vrl, "downcase(string!(parsed.trace_id))")
			spanIdDowncase := strings.Count(vrl, "downcase(string!(parsed.span_id))")

			Expect(traceIdDowncase).To(Equal(1))
			Expect(spanIdDowncase).To(Equal(1))
		})
	})
})
