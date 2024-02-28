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

	DescribeTable("tuning", func(expValid bool, spec loggingv1.OutputSpec) {
		valid, msg := VerifyTuning(spec)
		Expect(valid).To(Equal(expValid), msg)
	},
		Entry("should succeed when tuning is not spec'd", true, loggingv1.OutputSpec{
			Type: loggingv1.OutputTypeElasticsearch,
		}),
		Entry("should succeed when is spec'd with valid fields", true, loggingv1.OutputSpec{
			Type: loggingv1.OutputTypeElasticsearch,
			Tuning: &loggingv1.OutputTuningSpec{
				Compression:      "gzip",
				Delivery:         "AtLeastOnce",
				MinRetryDuration: utils.GetPtr(time.Duration(10)),
				MaxRetryDuration: utils.GetPtr(time.Duration(20)),
				MaxWrite:         utils.GetPtr(resource.MustParse("10M")),
			},
		}),
		Entry("should pass for syslog when compression is empty", true, loggingv1.OutputSpec{
			Type: loggingv1.OutputTypeSyslog,
			Tuning: &loggingv1.OutputTuningSpec{
				Compression: "",
			},
		}),
		Entry("should pass for azure when compression is empty", true, loggingv1.OutputSpec{
			Type: loggingv1.OutputTypeAzureMonitor,
			Tuning: &loggingv1.OutputTuningSpec{
				Compression: "",
			},
		}),
		Entry("should fail for syslog when compression is spec'd", false, loggingv1.OutputSpec{
			Type: loggingv1.OutputTypeSyslog,
			Tuning: &loggingv1.OutputTuningSpec{
				Compression: "gzip",
			},
		}),
		Entry("should fail for azure when compression is spec'd", false, loggingv1.OutputSpec{
			Type: loggingv1.OutputTypeAzureMonitor,
			Tuning: &loggingv1.OutputTuningSpec{
				Compression: "gzip",
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
	)

})
