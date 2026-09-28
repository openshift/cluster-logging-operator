package outputs

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
)

var _ = Describe("#validateOutputTemplates (LOG-9752)", func() {

	It("should accept benign template fields", func() {
		out := obs.OutputSpec{
			Type:  obs.OutputTypeKafka,
			Kafka: &obs.Kafka{Topic: "topic-{.log_type||\"none\"}"},
		}
		Expect(validateOutputTemplates(out)).To(BeEmpty())
	})

	It("should reject a kafka topic containing the terminator", func() {
		out := obs.OutputSpec{
			Type:  obs.OutputTypeKafka,
			Kafka: &obs.Kafka{Topic: "x'''injection"},
		}
		msgs := validateOutputTemplates(out)
		Expect(msgs).ToNot(BeEmpty())
		Expect(msgs[0]).To(ContainSubstring("kafka.topic"))
	})

	It("should reject a splunk index containing a newline", func() {
		out := obs.OutputSpec{
			Type:   obs.OutputTypeSplunk,
			Splunk: &obs.Splunk{Index: "line1\nline2"},
		}
		msgs := validateOutputTemplates(out)
		Expect(msgs).ToNot(BeEmpty())
		Expect(msgs[0]).To(ContainSubstring("splunk.index"))
	})

	It("should reject a cloudwatch groupName containing the terminator", func() {
		out := obs.OutputSpec{
			Type:       obs.OutputTypeCloudwatch,
			Cloudwatch: &obs.Cloudwatch{GroupName: "grp'''x"},
		}
		msgs := validateOutputTemplates(out)
		Expect(msgs).ToNot(BeEmpty())
		Expect(msgs[0]).To(ContainSubstring("cloudwatch.groupName"))
	})
})
