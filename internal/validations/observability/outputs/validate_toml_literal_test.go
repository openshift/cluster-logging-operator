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

	It("should reject a splunk source containing the terminator", func() {
		out := obs.OutputSpec{
			Type:   obs.OutputTypeSplunk,
			Splunk: &obs.Splunk{Source: "src'''x"},
		}
		msgs := validateOutputTemplates(out)
		Expect(msgs).ToNot(BeEmpty())
		Expect(msgs[0]).To(ContainSubstring("splunk.source"))
	})

	It("should reject a splunk sourceType containing the terminator", func() {
		out := obs.OutputSpec{
			Type:   obs.OutputTypeSplunk,
			Splunk: &obs.Splunk{SourceType: "st'''x"},
		}
		msgs := validateOutputTemplates(out)
		Expect(msgs).ToNot(BeEmpty())
		Expect(msgs[0]).To(ContainSubstring("splunk.sourceType"))
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

	It("should reject a splunk payloadKey containing the terminator", func() {
		out := obs.OutputSpec{
			Type:   obs.OutputTypeSplunk,
			Splunk: &obs.Splunk{PayloadKey: obs.FieldPath(`.foo."x'''y"`)},
		}
		msgs := validateOutputTemplates(out)
		Expect(msgs).ToNot(BeEmpty())
		Expect(msgs[0]).To(ContainSubstring("splunk.payloadKey"))
	})

	It("should reject a splunk indexedFields entry containing the terminator", func() {
		out := obs.OutputSpec{
			Type:   obs.OutputTypeSplunk,
			Splunk: &obs.Splunk{IndexedFields: []obs.FieldPath{`.good`, `.foo."x'''y"`}},
		}
		msgs := validateOutputTemplates(out)
		Expect(msgs).ToNot(BeEmpty())
		Expect(msgs[0]).To(ContainSubstring("splunk.indexedFields[1]"))
	})

	It("should reject an s3 keyPrefix containing the terminator", func() {
		out := obs.OutputSpec{
			Type: obs.OutputTypeS3,
			S3:   &obs.S3{KeyPrefix: "pre'''x"},
		}
		msgs := validateOutputTemplates(out)
		Expect(msgs).ToNot(BeEmpty())
		Expect(msgs[0]).To(ContainSubstring("s3.keyPrefix"))
	})

	It("should reject a syslog facility containing the terminator", func() {
		out := obs.OutputSpec{
			Type:   obs.OutputTypeSyslog,
			Syslog: &obs.Syslog{Facility: "fac'''x"},
		}
		msgs := validateOutputTemplates(out)
		Expect(msgs).ToNot(BeEmpty())
		Expect(msgs[0]).To(ContainSubstring("syslog.facility"))
	})

	It("should reject a syslog appName containing a newline", func() {
		out := obs.OutputSpec{
			Type:   obs.OutputTypeSyslog,
			Syslog: &obs.Syslog{AppName: "app\nname"},
		}
		msgs := validateOutputTemplates(out)
		Expect(msgs).ToNot(BeEmpty())
		Expect(msgs[0]).To(ContainSubstring("syslog.appName"))
	})
})
