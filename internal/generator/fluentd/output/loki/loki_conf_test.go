package loki

import (
	"github.com/openshift/cluster-logging-operator/test/helpers"
	"sort"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	logging "github.com/openshift/cluster-logging-operator/apis/logging/v1"
	v1 "github.com/openshift/cluster-logging-operator/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/internal/generator"
	corev1 "k8s.io/api/core/v1"
)

var _ = Describe("outputLabelConf", func() {
	var (
		loki *logging.Loki
	)
	BeforeEach(func() {
		loki = &logging.Loki{}
	})
	Context("#lokiLabelKeys when LabelKeys", func() {
		Context("are not spec'd", func() {
			It("should provide a default set of labels including the required ones", func() {
				exp := append(defaultLabelKeys, requiredLabelKeys...)
				sort.Strings(exp)
				Expect(lokiLabelKeys(loki)).To(BeEquivalentTo(exp))
			})
		})
		Context("are spec'd", func() {
			It("should use the ones provided and add the required ones", func() {
				loki.LabelKeys = []string{"foo"}
				exp := append(loki.LabelKeys, requiredLabelKeys...)
				Expect(lokiLabelKeys(loki)).To(BeEquivalentTo(exp))
			})
		})

	})
})

var _ = Describe("Generate fluentd config", func() {
	var f = func(clspec logging.CollectionSpec, secrets map[string]*corev1.Secret, clfspec logging.ClusterLogForwarderSpec, op generator.Options) []generator.Element {
		var bufspec *logging.FluentdBufferSpec = nil
		if clspec.Fluentd != nil &&
			clspec.Fluentd.Buffer != nil {
			bufspec = clspec.Fluentd.Buffer
		}
		return Conf(bufspec, secrets[clfspec.Outputs[0].Name], clfspec.Outputs[0], generator.NoOptions)
	}
	DescribeTable("for Loki output", helpers.TestGenerateConfWith(f),
		Entry("with default labels", helpers.ConfGenerateTest{
			CLFSpec: logging.ClusterLogForwarderSpec{
				Outputs: []logging.OutputSpec{
					{
						Type: logging.OutputTypeLoki,
						Name: "loki-receiver",
						URL:  "https://logs-us-west1.grafana.net",
						Secret: &logging.OutputSecretSpec{
							Name: "es-1",
						},
					},
				},
			},
			Secrets: map[string]*corev1.Secret{
				"loki-receiver": {
					Data: map[string][]byte{
						"username": []byte("junk"),
						"password": []byte("junk"),
					},
				},
			},
			ExpectedConf: `
<label @LOKI_RECEIVER>
  <filter **>
    @type record_modifier
    <record>
      _kubernetes_container_name ${record.dig("kubernetes","container_name")}
      _kubernetes_host "#{ENV['NODE_NAME']}"
      _kubernetes_namespace_name ${record.dig("kubernetes","namespace_name")}
      _kubernetes_pod_name ${record.dig("kubernetes","pod_name")}
      _log_type ${record.dig("log_type")}
    </record>
  </filter>
  
  <match **>
    @type loki
    @id loki_receiver
    line_format json
    url https://logs-us-west1.grafana.net
    username "#{File.read('/var/run/ocp-collector/secrets/es-1/username') rescue nil}"
    password "#{File.read('/var/run/ocp-collector/secrets/es-1/password') rescue nil}"
    <label>
      kubernetes_container_name _kubernetes_container_name
      kubernetes_host _kubernetes_host
      kubernetes_namespace_name _kubernetes_namespace_name
      kubernetes_pod_name _kubernetes_pod_name
      log_type _log_type
    </label>
    <buffer>
      @type file
      path '/var/lib/fluentd/loki_receiver'
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
    </buffer>
  </match>
</label>
`,
		}),
		Entry("with custom labels", helpers.ConfGenerateTest{
			CLFSpec: logging.ClusterLogForwarderSpec{
				Outputs: []logging.OutputSpec{
					{
						Type: logging.OutputTypeLoki,
						Name: "loki-receiver",
						URL:  "https://logs-us-west1.grafana.net",
						Secret: &logging.OutputSecretSpec{
							Name: "es-1",
						},
						OutputTypeSpec: v1.OutputTypeSpec{Loki: &v1.Loki{
							LabelKeys: []string{"kubernetes.labels.app", "kubernetes.container_name"},
						}},
					},
				},
			},
			Secrets: map[string]*corev1.Secret{
				"loki-receiver": {
					Data: map[string][]byte{
						"username": []byte("junk"),
						"password": []byte("junk"),
					},
				},
			},
			ExpectedConf: `
<label @LOKI_RECEIVER>
  <filter **>
    @type record_modifier
    <record>
      _kubernetes_container_name ${record.dig("kubernetes","container_name")}
      _kubernetes_host "#{ENV['NODE_NAME']}"
      _kubernetes_labels_app ${record.dig("kubernetes","labels","app")}
    </record>
  </filter>
  
  <match **>
    @type loki
    @id loki_receiver
    line_format json
    url https://logs-us-west1.grafana.net
    username "#{File.read('/var/run/ocp-collector/secrets/es-1/username') rescue nil}"
    password "#{File.read('/var/run/ocp-collector/secrets/es-1/password') rescue nil}"
    <label>
      kubernetes_container_name _kubernetes_container_name
      kubernetes_host _kubernetes_host
      kubernetes_labels_app _kubernetes_labels_app
    </label>
    <buffer>
      @type file
      path '/var/lib/fluentd/loki_receiver'
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
    </buffer>
  </match>
</label>
`,
		}),
	)
})

func TestFluendConfGenerator(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Fluend Conf Generation")
}
