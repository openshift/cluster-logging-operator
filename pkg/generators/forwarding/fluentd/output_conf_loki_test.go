package fluentd

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	v1 "github.com/openshift/cluster-logging-operator/pkg/apis/logging/v1"
	. "github.com/openshift/cluster-logging-operator/test"
)

var _ = Describe("Generating external loki server output store config block", func() {
	var (
		err       error
		outputs   []v1.OutputSpec
		generator *ConfigGenerator
	)
	BeforeEach(func() {
		generator, err = NewConfigGenerator(false, false, false)
		Expect(err).To(BeNil())
	})

	Context("for a loki endpoint", func() {
		Context("with insecure communication", func() {
			lokiConf := `
        <label @LOKI_RECEIVER>
          <match **>
            @type loki
            url https://logs-us-west1.grafana.net
            tenant a-tenant
            line_format json
            extract_kubernetes_labels true
            remove_keys docker, kubernetes, pipeline_metadata, hostname, index_name
            <label>
              # docker
              container_id $.docker.container_id

              # kubernetes
              container_name $.kubernetes.container_name
              container_image $.kubernetes.container_image
              container_image_id $.kubernetes.container_image_id
              kubernetes_host $.kubernetes.host
              master_url $.kubernetes.master_url
              pod_name $.kubernetes.pod_name
              pod_id $.kubernetes.pod_id
              namespace_name $.kubernetes.namespace_name
              namespace_id $.kubernetes.namespace_id

              # general
              index_name $.viaq_index_name
              hostname $.hostname
            </label>
            <buffer>
              @type file
              path '/var/lib/fluentd/loki_receiver'
              flush_interval 1s
              flush_thread_count 2
              flush_at_shutdown false
              retry_max_interval 300
              retry_forever true
              queued_chunks_limit_size "#{ENV['BUFFER_QUEUE_LIMIT'] || '32' }"
              chunk_limit_size "#{ENV['BUFFER_SIZE_LIMIT'] || '8m' }"
              overflow_action "#{ENV['BUFFER_QUEUE_FULL_ACTION'] || 'block'}"
            </buffer>
          </match>
        </label>`

			BeforeEach(func() {
				outputs = []v1.OutputSpec{
					{
						Type: v1.OutputTypeLoki,
						Name: "loki-receiver",
						URL:  "https://logs-us-west1.grafana.net/a-tenant",
					},
				}
			})

			It("should produce well formed output label config", func() {
				results, err := generator.generateOutputLabelBlocks(outputs)
				Expect(err).To(BeNil())
				Expect(len(results)).To(Equal(1))
				Expect(results[0]).To(EqualTrimLines(lokiConf))
			})
		})

		Context("with secured communication", func() {
			lokiConf := `
        <label @LOKI_RECEIVER>
          <match **>
            @type loki
            url https://logs-us-west1.grafana.net
            tenant a-tenant
            line_format json
            extract_kubernetes_labels true
            remove_keys docker, kubernetes, pipeline_metadata, hostname, index_name
            ca_cert '/var/run/ocp-collector/secrets/a-secret-ref/ca-bundle.crt'
            cert '/var/run/ocp-collector/secrets/a-secret-ref/tls.crt'
            key '/var/run/ocp-collector/secrets/a-secret-ref/tls.key'
            <label>
              # docker
              container_id $.docker.container_id

              # kubernetes
              container_name $.kubernetes.container_name
              container_image $.kubernetes.container_image
              container_image_id $.kubernetes.container_image_id
              kubernetes_host $.kubernetes.host
              master_url $.kubernetes.master_url
              pod_name $.kubernetes.pod_name
              pod_id $.kubernetes.pod_id
              namespace_name $.kubernetes.namespace_name
              namespace_id $.kubernetes.namespace_id

              # general
              index_name $.viaq_index_name
              hostname $.hostname
            </label>
            <buffer>
              @type file
              path '/var/lib/fluentd/loki_receiver'
              flush_interval 1s
              flush_thread_count 2
              flush_at_shutdown false
              retry_max_interval 300
              retry_forever true
              queued_chunks_limit_size "#{ENV['BUFFER_QUEUE_LIMIT'] || '32' }"
              chunk_limit_size "#{ENV['BUFFER_SIZE_LIMIT'] || '8m' }"
              overflow_action "#{ENV['BUFFER_QUEUE_FULL_ACTION'] || 'block'}"
            </buffer>
          </match>
        </label>`

			BeforeEach(func() {
				outputs = []v1.OutputSpec{
					{
						Type: v1.OutputTypeLoki,
						Name: "loki-receiver",
						URL:  "https://logs-us-west1.grafana.net/a-tenant",
						Secret: &v1.OutputSecretSpec{
							Name: "a-secret-ref",
						},
					},
				}
			})

			It("should produce well formed output label config", func() {
				results, err := generator.generateOutputLabelBlocks(outputs)
				Expect(err).To(BeNil())
				Expect(len(results)).To(Equal(1))
				Expect(results[0]).To(EqualTrimLines(lokiConf))
			})
		})

		Context("with output spec", func() {
			lokiConf := `
        <label @LOKI_RECEIVER>
          <match **>
            @type loki
            url https://logs-us-west1.grafana.net
            tenant custom-tenant
            line_format json
            extract_kubernetes_labels true
            remove_keys docker, kubernetes, pipeline_metadata, hostname, index_name
            ca_cert '/var/run/ocp-collector/secrets/a-secret-ref/ca-bundle.crt'
            cert '/var/run/ocp-collector/secrets/a-secret-ref/tls.crt'
            key '/var/run/ocp-collector/secrets/a-secret-ref/tls.key'
            <label>
              # docker
              container_id $.docker.container_id

              # kubernetes
              container_name $.kubernetes.container_name
              container_image $.kubernetes.container_image
              container_image_id $.kubernetes.container_image_id
              kubernetes_host $.kubernetes.host
              master_url $.kubernetes.master_url
              pod_name $.kubernetes.pod_name
              pod_id $.kubernetes.pod_id
              namespace_name $.kubernetes.namespace_name
              namespace_id $.kubernetes.namespace_id

              # general
              index_name $.viaq_index_name
              hostname $.hostname
          </label>
          <buffer>
              @type file
              path '/var/lib/fluentd/loki_receiver'
              flush_interval 1s
              flush_thread_count 2
              flush_at_shutdown false
              retry_max_interval 300
              retry_forever true
              queued_chunks_limit_size "#{ENV['BUFFER_QUEUE_LIMIT'] || '32' }"
              chunk_limit_size "#{ENV['BUFFER_SIZE_LIMIT'] || '8m' }"
              overflow_action "#{ENV['BUFFER_QUEUE_FULL_ACTION'] || 'block'}"
          </buffer>
        </match>
      </label>`

			BeforeEach(func() {
				outputs = []v1.OutputSpec{
					{
						Type: v1.OutputTypeLoki,
						Name: "loki-receiver",
						URL:  "https://logs-us-west1.grafana.net/a-tenant",
						OutputTypeSpec: v1.OutputTypeSpec{
							Loki: &v1.Loki{
								TenantID: "custom-tenant",
							},
						},
						Secret: &v1.OutputSecretSpec{
							Name: "a-secret-ref",
						},
					},
				}
			})

			It("should produce well formed output label config", func() {
				results, err := generator.generateOutputLabelBlocks(outputs)
				Expect(err).To(BeNil())
				Expect(len(results)).To(Equal(1))
				Expect(results[0]).To(EqualTrimLines(lokiConf))
			})
		})
	})
})
