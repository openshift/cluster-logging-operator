package fluentd

import (
	"github.com/openshift/cluster-logging-operator/internal/constants"
	"github.com/openshift/cluster-logging-operator/internal/logstore/lokistack"
	"github.com/openshift/cluster-logging-operator/test/helpers"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	logging "github.com/openshift/cluster-logging-operator/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/internal/generator"
	corev1 "k8s.io/api/core/v1"
)

var _ = Describe("[internal][generator][fluentd] Generating outputs", func() {
	var f = func(clspec logging.CollectionSpec, secrets map[string]*corev1.Secret, clfspec logging.ClusterLogForwarderSpec, op generator.Options) []generator.Element {
		return Outputs(nil, secrets, &clfspec, op)
	}
	DescribeTable("using #Outputs", helpers.TestGenerateConfWith(f),
		Entry("should honor global minTLSVersion & ciphers with loki as the default logstore regardless of the feature gate setting", helpers.ConfGenerateTest{
			CLFSpec: logging.ClusterLogForwarderSpec{
				Pipelines: []logging.PipelineSpec{
					{Name: logging.InputNameApplication, OutputRefs: []string{lokistack.FormatOutputNameFromInput(logging.InputNameApplication)}},
				},
				Outputs: []logging.OutputSpec{
					{
						Type: logging.OutputTypeLoki,
						Name: lokistack.FormatOutputNameFromInput(logging.InputNameApplication),
						URL:  "https://lokistack-dev-gateway-http.openshift-logging.svc:8080/api/logs/v1/application",
					},
				},
			},
			Secrets: map[string]*corev1.Secret{
				constants.LogCollectorToken: {
					Data: map[string][]byte{
						"token": []byte("token-for-loki"),
					},
				},
			},
			Options: generator.Options{},
			ExpectedConf: `
    # Ship logs to specific outputs
    <label @DEFAULT_LOKI_APPS>
      #dedot namespace_labels and rebuild message field if present
      <filter **>
        @type record_modifier
        <record>
          _dummy_ ${if m=record.dig("kubernetes","namespace_labels");record["kubernetes"]["namespace_labels"]={}.tap{|n|m.each{|k,v|n[k.gsub(/[.\/]/,'_')]=v}};end}
          _dummy2_ ${if m=record.dig("kubernetes","labels");record["kubernetes"]["labels"]={}.tap{|n|m.each{|k,v|n[k.gsub(/[.\/]/,'_')]=v}};end}
          _dummy3_ ${if m=record.dig("kubernetes","flat_labels");record["kubernetes"]["flat_labels"]=[].tap{|n|m.each_with_index{|s, i|n[i] = s.gsub(/[.\/]/,'_')}};end}
        </record>
        remove_keys _dummy_, _dummy2_, _dummy3_
      </filter>
      
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
        @id default_loki_apps
        line_format json
        url https://lokistack-dev-gateway-http.openshift-logging.svc:8080/api/logs/v1/application
        min_version TLS1_2
        ciphers TLS_AES_128_GCM_SHA256:TLS_AES_256_GCM_SHA384:TLS_CHACHA20_POLY1305_SHA256:ECDHE-ECDSA-AES128-GCM-SHA256:ECDHE-RSA-AES128-GCM-SHA256:ECDHE-ECDSA-AES256-GCM-SHA384:ECDHE-RSA-AES256-GCM-SHA384:ECDHE-ECDSA-CHACHA20-POLY1305:ECDHE-RSA-CHACHA20-POLY1305:DHE-RSA-AES128-GCM-SHA256:DHE-RSA-AES256-GCM-SHA384
        ca_cert /var/run/secrets/kubernetes.io/serviceaccount/service-ca.crt
        bearer_token_file /var/run/secrets/kubernetes.io/serviceaccount/token
        <label>
          kubernetes_container_name _kubernetes_container_name
          kubernetes_host _kubernetes_host
          kubernetes_namespace_name _kubernetes_namespace_name
          kubernetes_pod_name _kubernetes_pod_name
          log_type _log_type
        </label>
        <buffer>
          @type file
          path '/var/lib/fluentd/default_loki_apps'
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
