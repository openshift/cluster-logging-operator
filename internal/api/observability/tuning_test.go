package observability_test

import (
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	internalobs "github.com/openshift/cluster-logging-operator/internal/api/observability"
	"github.com/openshift/cluster-logging-operator/internal/utils"
	"k8s.io/apimachinery/pkg/api/resource"
)

var _ = Describe("Output", func() {

	const (
		compression = "foo"
	)
	var (
		baseSpec = &obs.BaseOutputTuningSpec{
			Delivery:         obs.DeliveryModeAtLeastOnce,
			MaxWrite:         utils.GetPtr(resource.MustParse("1250G")),
			MaxRetryDuration: utils.GetPtr(time.Second),
			MinRetryDuration: utils.GetPtr(3 * time.Second),
		}
		kafkaBaseSpec = &obs.BaseOutputTuningSpec{
			Delivery: obs.DeliveryModeAtLeastOnce,
			MaxWrite: utils.GetPtr(resource.MustParse("1250G")),
		}
	)

	DescribeTable("#NewTuning", func(spec obs.OutputSpec, expBase *obs.BaseOutputTuningSpec, expCompression string) {
		tuningSpec := internalobs.NewTuning(spec)
		if expBase != nil {
			Expect(tuningSpec.BaseOutputTuningSpec).To(Equal(*expBase))
		}
		Expect(tuningSpec.Compression).To(Equal(expCompression))
	},
		Entry("with AzureMonitor", obs.OutputSpec{
			Type: obs.OutputTypeAzureMonitor,
			AzureMonitor: &obs.AzureMonitor{
				Tuning: baseSpec,
			},
		}, baseSpec, ""),
		Entry("with GoogleCloudLogging", obs.OutputSpec{
			Type: obs.OutputTypeGoogleCloudLogging,
			GoogleCloudLogging: &obs.GoogleCloudLogging{
				Tuning: &obs.GoogleCloudLoggingTuningSpec{
					BaseOutputTuningSpec: *baseSpec,
				},
			},
		}, baseSpec, ""),
		Entry("with Cloudwatch", obs.OutputSpec{
			Type: obs.OutputTypeCloudwatch,
			Cloudwatch: &obs.Cloudwatch{
				Tuning: &obs.CloudwatchTuningSpec{
					BaseOutputTuningSpec: *baseSpec,
					Compression:          compression,
				},
			},
		}, baseSpec, compression),
		Entry("with Elasticsearch", obs.OutputSpec{
			Type: obs.OutputTypeElasticsearch,
			Elasticsearch: &obs.Elasticsearch{
				Tuning: &obs.ElasticsearchTuningSpec{
					BaseOutputTuningSpec: *baseSpec,
					Compression:          compression,
				},
			},
		}, baseSpec, compression),
		Entry("with HTTP", obs.OutputSpec{
			Type: obs.OutputTypeHTTP,
			HTTP: &obs.HTTP{
				Tuning: &obs.HTTPTuningSpec{
					BaseOutputTuningSpec: *baseSpec,
					Compression:          compression,
				},
			},
		}, baseSpec, compression),
		Entry("with OTLP", obs.OutputSpec{
			Type: obs.OutputTypeOTLP,
			OTLP: &obs.OTLP{
				Tuning: &obs.OTLPTuningSpec{
					BaseOutputTuningSpec: *baseSpec,
					Compression:          compression,
				},
			},
		}, baseSpec, compression),
		Entry("with Kafka", obs.OutputSpec{
			Type: obs.OutputTypeKafka,
			Kafka: &obs.Kafka{
				Tuning: &obs.KafkaTuningSpec{
					Delivery:    obs.DeliveryModeAtLeastOnce,
					MaxWrite:    utils.GetPtr(resource.MustParse("1250G")),
					Compression: compression,
				},
			},
		}, kafkaBaseSpec, compression),
		Entry("with Loki", obs.OutputSpec{
			Type: obs.OutputTypeLoki,
			Loki: &obs.Loki{
				Tuning: &obs.LokiTuningSpec{
					BaseOutputTuningSpec: *baseSpec,
					Compression:          compression,
				},
			},
		}, baseSpec, compression),
		Entry("with LokiStack", obs.OutputSpec{
			Type: obs.OutputTypeLokiStack,
			LokiStack: &obs.LokiStack{
				Tuning: &obs.LokiTuningSpec{
					BaseOutputTuningSpec: *baseSpec,
					Compression:          compression,
				},
			},
		}, baseSpec, compression),
		Entry("with Splunk", obs.OutputSpec{
			Type: obs.OutputTypeSplunk,
			Splunk: &obs.Splunk{
				Tuning: &obs.SplunkTuningSpec{
					BaseOutputTuningSpec: *baseSpec,
				},
			},
		}, baseSpec, ""),
	)
})
