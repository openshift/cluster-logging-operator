package output_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	logging "github.com/openshift/cluster-logging-operator/apis/logging/v1"
	. "github.com/openshift/cluster-logging-operator/internal/generator"
	"github.com/openshift/cluster-logging-operator/internal/generator/fluentd/helpers"
	"github.com/openshift/cluster-logging-operator/internal/generator/fluentd/output"
	. "github.com/openshift/cluster-logging-operator/test/helpers"
	corev1 "k8s.io/api/core/v1"
)

var _ = Describe("Generate fluentd conf", func() {
	var f = func(clspec logging.ClusterLoggingSpec, secrets map[string]*corev1.Secret, clfspec logging.ClusterLogForwarderSpec, op Options) []Element {
		es := make([][]Element, len(clfspec.Outputs))
		for i := range clfspec.Outputs {
			storeID := helpers.StoreID("", clfspec.Outputs[i].Name, "")
			es[i] = output.Buffer([]string{"time", "tag"}, clspec.Forwarder.Fluentd.Buffer, storeID, &clfspec.Outputs[i])
		}
		return MergeElements(es...)
	}
	DescribeTable("Buffers", TestGenerateConfWith(f),
		Entry("With tuning parameters", ConfGenerateTest{
			CLFSpec: logging.ClusterLogForwarderSpec{
				Outputs: []logging.OutputSpec{
					{
						Name: "es-1",
					},
				},
			},
			CLSpec: logging.ClusterLoggingSpec{
				Forwarder: &logging.ForwarderSpec{
					Fluentd: &logging.FluentdForwarderSpec{
						Buffer: &logging.FluentdBufferSpec{
							ChunkLimitSize:   "8m",
							TotalLimitSize:   "800000000",
							OverflowAction:   "throw_exception",
							FlushThreadCount: 128,
							FlushMode:        "interval",
							FlushInterval:    "25s",
							RetryWait:        "20s",
							RetryType:        "periodic",
							RetryMaxInterval: "300s",
							RetryTimeout:     "60h",
						},
					},
				},
			},
			ExpectedConf: `
<buffer time, tag>
  @type file
  path '/var/lib/fluentd/es_1'
  flush_mode interval
  flush_interval 25s
  flush_thread_count 128
  retry_type periodic
  retry_wait 20s
  retry_max_interval 300s
  retry_timeout 60h
  queued_chunks_limit_size "#{ENV['BUFFER_QUEUE_LIMIT'] || '32'}"
  total_limit_size 800000000
  chunk_limit_size 8m
  overflow_action throw_exception
  disable_chunk_backup true
</buffer>`,
		}),
		Entry("when tuning flush_mode other then interval", ConfGenerateTest{
			CLFSpec: logging.ClusterLogForwarderSpec{
				Outputs: []logging.OutputSpec{
					{
						Name: "es-1",
					},
				},
			},
			CLSpec: logging.ClusterLoggingSpec{
				Forwarder: &logging.ForwarderSpec{
					Fluentd: &logging.FluentdForwarderSpec{
						Buffer: &logging.FluentdBufferSpec{
							FlushMode:     "lazy",
							FlushInterval: "25s",
						},
					},
				},
			},
			ExpectedConf: `
<buffer time, tag>
  @type file
  path '/var/lib/fluentd/es_1'
  flush_mode lazy
  flush_thread_count 2
  retry_type exponential_backoff
  retry_wait 1s
  retry_max_interval 60s
  retry_timeout 60m
  queued_chunks_limit_size "#{ENV['BUFFER_QUEUE_LIMIT'] || '32'}"
  total_limit_size "#{ENV['TOTAL_LIMIT_SIZE_PER_BUFFER'] || '8589934592'}"
  chunk_limit_size "#{ENV['BUFFER_SIZE_LIMIT'] || '8m'}"
  overflow_action block
  disable_chunk_backup true
</buffer>`,
		}),
	)
})

var _ = Describe("Generate fluentd conf", func() {
	var f = func(clspec logging.ClusterLoggingSpec, secrets map[string]*corev1.Secret, clfspec logging.ClusterLogForwarderSpec, op Options) []Element {
		es := make([][]Element, len(clfspec.Outputs))
		for i := range clfspec.Outputs {
			storeID := helpers.StoreID("retry_", clfspec.Outputs[i].Name, "")
			es[i] = output.Buffer([]string{}, clspec.Forwarder.Fluentd.Buffer, storeID, &clfspec.Outputs[i])
		}
		return MergeElements(es...)
	}
	DescribeTable("Retry Buffers", TestGenerateConfWith(f),
		Entry("With no tuning parameters", ConfGenerateTest{
			CLFSpec: logging.ClusterLogForwarderSpec{
				Pipelines: []logging.PipelineSpec{
					{
						InputRefs:  []string{logging.InputNameApplication},
						OutputRefs: []string{logging.OutputNameDefault},
						Name:       "defaultoutput",
					},
				},
				Outputs: []logging.OutputSpec{
					{
						Name: "es-2",
					},
				},
			},
			CLSpec: logging.ClusterLoggingSpec{
				Forwarder: &logging.ForwarderSpec{
					Fluentd: &logging.FluentdForwarderSpec{
						Buffer: nil,
					},
				},
			},
			ExpectedConf: `
<buffer>
  @type file
  path '/var/lib/fluentd/retry_es_2'
  flush_mode interval
  flush_interval 1s
  flush_thread_count 2
  retry_type exponential_backoff
  retry_wait 1s
  retry_max_interval 60s
  retry_timeout 60m
  queued_chunks_limit_size "#{ENV['BUFFER_QUEUE_LIMIT'] || '32'}"
  total_limit_size "#{ENV['TOTAL_LIMIT_SIZE_PER_BUFFER'] || '8589934592'}"
  chunk_limit_size "#{ENV['BUFFER_SIZE_LIMIT'] || '8m'}"
  overflow_action block
  disable_chunk_backup true
</buffer>`,
		}))
})
