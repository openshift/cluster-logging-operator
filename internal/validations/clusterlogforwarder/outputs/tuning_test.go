package outputs

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	loggingv1 "github.com/openshift/cluster-logging-operator/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/internal/utils"
	"k8s.io/apimachinery/pkg/api/resource"
	"time"
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
				MinRetryDuration: utils.GetPtr(10 * time.Second),
				MaxRetryDuration: utils.GetPtr(20 * time.Second),
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
				MaxRetryDuration: utils.GetPtr(1 * time.Second),
			},
		}),
		Entry("should fail for kafka when MinRetryDuration is spec'd", false, loggingv1.OutputSpec{
			Type: loggingv1.OutputTypeKafka,
			Tuning: &loggingv1.OutputTuningSpec{
				MinRetryDuration: utils.GetPtr(1 * time.Second),
			},
		}),
	)

})
