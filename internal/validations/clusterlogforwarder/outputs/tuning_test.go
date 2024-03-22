package outputs

import (
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	loggingv1 "github.com/openshift/cluster-logging-operator/api/logging/v1"
	"github.com/openshift/cluster-logging-operator/internal/utils"
	"k8s.io/apimachinery/pkg/api/resource"
)

var _ = Describe("Validate ", func() {

	Context("tuning", func() {
		DescribeTable("all fields except compression", func(expValid bool, spec loggingv1.OutputSpec) {
			valid, msg := VerifyTuning(spec)
			Expect(valid).To(Equal(expValid), msg)
		},
			Entry("should succeed when tuning is not spec'd", true, loggingv1.OutputSpec{
				Type: loggingv1.OutputTypeElasticsearch,
			}),
			Entry("should succeed when is spec'd with valid fields", true, loggingv1.OutputSpec{
				Type: loggingv1.OutputTypeElasticsearch,
				Tuning: &loggingv1.OutputTuningSpec{
					Delivery:         "AtLeastOnce",
					MinRetryDuration: utils.GetPtr(time.Duration(10)),
					MaxRetryDuration: utils.GetPtr(time.Duration(20)),
					MaxWrite:         utils.GetPtr(resource.MustParse("10M")),
				},
			}),
			Entry("should fail for kafka when MaxRetryDuration is spec'd", false, loggingv1.OutputSpec{
				Type: loggingv1.OutputTypeKafka,
				Tuning: &loggingv1.OutputTuningSpec{
					MaxRetryDuration: utils.GetPtr(time.Duration(1)),
				},
			}),
			Entry("should fail for kafka when MinRetryDuration is spec'd", false, loggingv1.OutputSpec{
				Type: loggingv1.OutputTypeKafka,
				Tuning: &loggingv1.OutputTuningSpec{
					MinRetryDuration: utils.GetPtr(time.Duration(1)),
				},
			}),
			Entry("should fail for syslog when maxWrite is spec'd", false, loggingv1.OutputSpec{
				Type: loggingv1.OutputTypeSyslog,
				Tuning: &loggingv1.OutputTuningSpec{
					MaxWrite: utils.GetPtr(resource.MustParse("10M")),
				},
			}),
		)

		Context("with compression", func() {
			DescribeTable("outputs with no compression support", func(outputType, compression string, expMsg string) {
				msg := verifyCompression(outputType, compression)
				Expect(msg).To(Equal(expMsg))
			},
				Entry("should fail for syslog when compression is spec'd", loggingv1.OutputTypeSyslog, "gzip", compressionNotSupportedForType),
				Entry("should fail for azure when compression is spec'd", loggingv1.OutputTypeAzureMonitor, "gzip", compressionNotSupportedForType),
				Entry("should fail for gcp when compression is spec'd", loggingv1.OutputTypeGoogleCloudLogging, "gzip", compressionNotSupportedForType),
				Entry("should pass for syslog when compression is empty", loggingv1.OutputTypeSyslog, "", ""),
				Entry("should pass for azure when compression is empty", loggingv1.OutputTypeAzureMonitor, "", ""),
				Entry("should pass for gcp when compression is empty", loggingv1.OutputTypeGoogleCloudLogging, "", ""),
			)

			DescribeTable("kafka", func(compression string, expMsg string) {
				msg := verifyCompression(loggingv1.OutputTypeKafka, compression)
				Expect(msg).To(Equal(expMsg))
			},
				// gzip should be supported but there is an issue with librdkafka
				Entry("should fail when gzip spec'd as compression", "gzip", compressionNotSupportedForType),
				Entry("should fail when zlib spec'd as compression", "zlib", compressionNotSupportedForType),
				Entry("should pass when lz4 spec'd as compression", "lz4", ""),
				Entry("should pass when snappy spec'd as compression", "snappy", ""),
				Entry("should pass when zstd spec'd as compression", "zstd", ""),
				Entry("should pass when no compression is spec'd", "", ""),
			)

			DescribeTable("elasticsearch", func(compression string, expMsg string) {
				msg := verifyCompression(loggingv1.OutputTypeElasticsearch, compression)
				Expect(msg).To(Equal(expMsg))
			},
				Entry("should fail when lz4 spec'd as compression", "lz4", compressionNotSupportedForType),
				Entry("should fail when snappy spec'd as compression", "snappy", compressionNotSupportedForType),
				Entry("should fail when zstd spec'd as compression", "zstd", compressionNotSupportedForType),
				Entry("should pass when zlib spec'd as compression", "zlib", ""),
				Entry("should pass when gzip spec'd as compression", "gzip", ""),
				Entry("should pass when no compression is spec'd", "", ""),
			)

			DescribeTable("cloudwatch", func(compression string, expMsg string) {
				msg := verifyCompression(loggingv1.OutputTypeCloudwatch, compression)
				Expect(msg).To(Equal(expMsg))
			},
				Entry("should fail when lz4 spec'd as compression", "lz4", compressionNotSupportedForType),
				Entry("should pass when snappy spec'd as compression", "snappy", ""),
				Entry("should pass when zstd spec'd as compression", "zstd", ""),
				Entry("should pass when zlib spec'd as compression", "zlib", ""),
				Entry("should pass when gzip spec'd as compression", "gzip", ""),
				Entry("should pass when no compression is spec'd", "", ""),
			)

			DescribeTable("splunk", func(compression string, expMsg string) {
				msg := verifyCompression(loggingv1.OutputTypeSplunk, compression)
				Expect(msg).To(Equal(expMsg))
			},
				Entry("should fail when lz4 spec'd as compression", "lz4", compressionNotSupportedForType),
				Entry("should fail when snappy spec'd as compression", "snappy", compressionNotSupportedForType),
				Entry("should fail when zstd spec'd as compression", "zstd", compressionNotSupportedForType),
				Entry("should fail when zlib spec'd as compression", "zlib", compressionNotSupportedForType),
				Entry("should pass when gzip spec'd as compression", "gzip", ""),
				Entry("should pass when no compression is spec'd", "", ""),
			)

			DescribeTable("http", func(compression string, expMsg string) {
				msg := verifyCompression(loggingv1.OutputTypeHttp, compression)
				Expect(msg).To(Equal(expMsg))
			},
				Entry("should fail when lz4 spec'd as compression", "lz4", compressionNotSupportedForType),
				Entry("should pass when snappy spec'd as compression", "snappy", ""),
				Entry("should pass when zstd spec'd as compression", "zstd", ""),
				Entry("should pass when zlib spec'd as compression", "zlib", ""),
				Entry("should pass when gzip spec'd as compression", "gzip", ""),
				Entry("should pass when no compression is spec'd", "", ""),
			)

			DescribeTable("loki", func(compression string, expMsg string) {
				msg := verifyCompression(loggingv1.OutputTypeLoki, compression)
				Expect(msg).To(Equal(expMsg))
			},
				Entry("should fail when lz4 spec'd as compression", "lz4", compressionNotSupportedForType),
				Entry("should fail when zstd spec'd as compression", "zstd", compressionNotSupportedForType),
				Entry("should fail when zlib spec'd as compression", "zlib", compressionNotSupportedForType),
				Entry("should pass when snappy spec'd as compression", "snappy", ""),
				Entry("should pass when gzip spec'd as compression", "gzip", ""),
				Entry("should pass when no compression is spec'd", "", ""),
			)

		})
	})
})
