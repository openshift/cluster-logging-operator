package drop

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	"github.com/openshift/cluster-logging-operator/test/matchers"
)

var _ = Describe("drop filter", func() {

	Context("#normalizeOlderThan", func() {
		DescribeTable("normalizes the cutoff to UTC for a Vector timestamp literal", func(value, expected string) {
			cutoff, err := normalizeOlderThan(value)
			Expect(err).NotTo(HaveOccurred())
			Expect(cutoff).To(Equal(expected))
		},
			Entry("date becomes UTC midnight", "2026-09-16", "2026-09-16T00:00:00Z"),
			Entry("UTC timestamp", "2026-09-16T00:15:30Z", "2026-09-16T00:15:30Z"),
			Entry("negative offset", "2026-09-16T00:15:30-04:00", "2026-09-16T04:15:30Z"),
			Entry("positive offset", "2026-09-16T12:15:30+05:30", "2026-09-16T06:45:30Z"),
			Entry("maximum offset components", "2026-09-16T23:59:00+23:59", "2026-09-16T00:00:00Z"),
			Entry("offset crosses the date boundary", "2026-09-16T00:15:30+05:30", "2026-09-15T18:45:30Z"),
			Entry("fractional seconds", "2026-09-16T00:15:30.123456789-04:00", "2026-09-16T04:15:30.123456789Z"),
		)

		DescribeTable("returns an error for an unparseable cutoff", func(value string) {
			cutoff, err := normalizeOlderThan(value)
			Expect(err).To(HaveOccurred())
			Expect(cutoff).To(BeEmpty())
		},
			Entry("empty value", ""),
			Entry("missing offset", "2026-09-16T00:15:30"),
			Entry("impossible date", "2026-02-30"),
		)
	})

	Context("#VRL", func() {
		DescribeTable("requires both an older timestamp and a matching field regardless of condition order", func(conditions []obs.DropCondition, expected string) {
			vrl, err := NewFilter([]obs.DropTest{{DropConditions: conditions}}).VRL()

			Expect(err).NotTo(HaveOccurred())
			Expect(vrl).To(Equal(expected))
		},
			Entry("olderThan before field", []obs.DropCondition{
				{OlderThan: "2026-07-01"},
				{Field: ".message", Matches: "historical"},
			}, `!((((parse_timestamp(to_string(.timestamp) ?? "", "%+") < t'2026-07-01T00:00:00Z') ?? false) && match(to_string(._internal.message) ?? "", r'historical')))`),
			Entry("field before olderThan", []obs.DropCondition{
				{Field: ".message", Matches: "historical"},
				{OlderThan: "2026-07-01"},
			}, `!((match(to_string(._internal.message) ?? "", r'historical') && ((parse_timestamp(to_string(.timestamp) ?? "", "%+") < t'2026-07-01T00:00:00Z') ?? false)))`),
		)

		It("should normalize olderThan dates and offsets into strict timestamp predicates", func() {
			spec := []obs.DropTest{
				{
					DropConditions: []obs.DropCondition{{OlderThan: "2026-07-01"}},
				},
				{
					DropConditions: []obs.DropCondition{{OlderThan: "2026-07-01T00:00:00-04:00"}},
				},
			}

			Expect(NewFilter(spec).VRL()).To(matchers.EqualTrimLines(`
!((((parse_timestamp(to_string(.timestamp) ?? "", "%+") < t'2026-07-01T00:00:00Z') ?? false)) || (((parse_timestamp(to_string(.timestamp) ?? "", "%+") < t'2026-07-01T04:00:00Z') ?? false)))
`))
		})

		It("should AND conditions in a test and OR separate tests", func() {
			spec := []obs.DropTest{
				{
					DropConditions: []obs.DropCondition{
						{OlderThan: "2026-07-01"},
						{OlderThan: "2026-07-01T00:00:00-04:00"},
					},
				},
				{
					DropConditions: []obs.DropCondition{{Field: ".log_type", Matches: "application"}},
				},
			}

			Expect(NewFilter(spec).VRL()).To(matchers.EqualTrimLines(`
!((((parse_timestamp(to_string(.timestamp) ?? "", "%+") < t'2026-07-01T00:00:00Z') ?? false) && ((parse_timestamp(to_string(.timestamp) ?? "", "%+") < t'2026-07-01T04:00:00Z') ?? false)) || (match(to_string(._internal.log_type) ?? "", r'application')))
`))
		})

		It("should return an error without a partial transform for an invalid olderThan cutoff", func() {
			spec := []obs.DropTest{
				{
					DropConditions: []obs.DropCondition{{OlderThan: "2026-07-01T00:00:00"}},
				},
			}

			vrl, err := NewFilter(spec).VRL()
			Expect(err).To(HaveOccurred())
			Expect(vrl).To(BeEmpty())
		})

		It("should reject matches containing single quotes", func() {
			spec := []obs.DropTest{
				{
					DropConditions: []obs.DropCondition{
						{
							Field:   ".kubernetes.namespace_name",
							Matches: "foo'bar",
						},
					},
				},
			}
			_, err := NewFilter(spec).VRL()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("single quotes"))
		})

		It("should reject notMatches containing single quotes", func() {
			spec := []obs.DropTest{
				{
					DropConditions: []obs.DropCondition{
						{
							Field:      ".kubernetes.namespace_name",
							NotMatches: "x'''[sources.evil]",
						},
					},
				},
			}
			_, err := NewFilter(spec).VRL()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("single quotes"))
		})

		It("should generate valid VRL for dropping", func() {
			spec := []obs.DropTest{
				{
					DropConditions: []obs.DropCondition{
						{
							Field:   ".kubernetes.namespace_name",
							Matches: "busybox",
						},
						{
							Field:      ".level",
							NotMatches: "d.+",
						},
					},
				},
				{
					DropConditions: []obs.DropCondition{
						{
							Field:   ".log_type",
							Matches: "application",
						},
					},
				},
				{
					DropConditions: []obs.DropCondition{
						{
							Field:   ".kubernetes.container_name",
							Matches: "error|warning",
						},
						{
							Field:      ".kubernetes.labels.test",
							NotMatches: "foo",
						},
					},
				},
			}
			Expect(NewFilter(spec).VRL()).To(matchers.EqualTrimLines(`
!((match(to_string(._internal.kubernetes.namespace_name) ?? "", r'busybox') && !match(to_string(._internal.level) ?? "", r'd.+')) || (match(to_string(._internal.log_type) ?? "", r'application')) || (match(to_string(._internal.kubernetes.container_name) ?? "", r'error|warning') && !match(to_string(._internal.kubernetes.labels.test) ?? "", r'foo')))
`))
		})
	})

})
