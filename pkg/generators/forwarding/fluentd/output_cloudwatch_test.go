package fluentd

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"

	loggingv1 "github.com/openshift/cluster-logging-operator/pkg/apis/logging/v1"

	. "github.com/openshift/cluster-logging-operator/test/matchers"
)

var _ = Describe("Generating fluentd config", func() {
	var (
		forwarderSpec *loggingv1.ForwarderSpec
		generator     *ConfigGenerator
		output        = loggingv1.OutputSpec{
			Type: loggingv1.OutputTypeCloudwatch,
			Name: "my-cloudwatch",
			OutputTypeSpec: loggingv1.OutputTypeSpec{
				Cloudwatch: &loggingv1.Cloudwatch{
					Region:  "anumber1",
					GroupBy: loggingv1.LogGroupByNamespaceName,
				},
			},
			Secret: &loggingv1.OutputSecretSpec{
				Name: "my-secret",
			},
		}
		outputs = []loggingv1.OutputSpec{output}
		secrets = map[string]*corev1.Secret{
			output.Secret.Name: {
				Data: map[string][]byte{
					"aws_access_key_id":     nil,
					"aws_secret_access_key": nil,
				},
			},
		}
	)

	Context("for cloudwatch output ", func() {
		BeforeEach(func() {
			var err error
			generator, err = NewConfigGenerator(false, false, true)
			Expect(err).To(BeNil())
			Expect(generator).ToNot(BeNil())
		})

		Context("grouped by log type", func() {
			BeforeEach(func() {
				output.Cloudwatch.GroupBy = loggingv1.LogGroupByLogType
			})
			It("should provide a valid configuration", func() {
				expConf := `
			<label @MY_CLOUDWATCH>
				<filter kubernetes.**>
					@type record_transformer
					enable_ruby true
					<record>
						cw_group_name application
						cw_stream_name ${tag}
					</record>
				</filter>
				<filter journal **_default_** **_kube-*_** **_openshift-*_** **_openshift_**>
					@type record_transformer
					enable_ruby true
					<record>
						cw_group_name infrastructure
						cw_stream_name ${record['hostname']}.${tag}
					</record>
				</filter>
				<filter *audit.log>
					@type record_transformer
					enable_ruby true
					<record>
						cw_group_name audit
						cw_stream_name ${record['hostname']}.${tag}
					</record>
				</filter>
				<match **>
					@type cloudwatch_logs
					auto_create_stream true
					region anumber1
					log_group_name_key cw_group_name
					log_stream_name_key cw_stream_name
					remove_log_stream_name_key true
					remove_log_group_name_key true
					auto_create_stream true
					concurrency 2
					aws_key_id "#{open('/var/run/ocp-collector/secrets/my-secret/aws_access_key_id','r') do |f|f.read end}"
					aws_sec_key "#{open('/var/run/ocp-collector/secrets/my-secret/aws_secret_access_key','r') do |f|f.read end}"
					include_time_key true
					log_rejected_request true
				</match>
			</label>`
				results, err := generator.generateOutputLabelBlocks(outputs, secrets, forwarderSpec)
				Expect(err).To(BeNil())
				Expect(len(results)).To(Equal(1))
				Expect(results[0]).To(EqualTrimLines(expConf))
			})
		})
		Context("grouped by namespace", func() {
			BeforeEach(func() {
				output.Cloudwatch.GroupBy = loggingv1.LogGroupByNamespaceName
			})
			It("should provide a valid configuration", func() {
				expConf := `
			<label @MY_CLOUDWATCH>
				<filter kubernetes.**>
					@type record_transformer
					enable_ruby true
					<record>
						cw_group_name ${record['kubernetes']['namespace_name']}
						cw_stream_name ${tag}
					</record>
				</filter>
				<filter journal **_default_** **_kube-*_** **_openshift-*_** **_openshift_**>
					@type record_transformer
					enable_ruby true
					<record>
						cw_group_name infrastructure
						cw_stream_name ${record['hostname']}.${tag}
					</record>
				</filter>
				<filter *audit.log>
					@type record_transformer
					enable_ruby true
					<record>
						cw_group_name audit
						cw_stream_name ${record['hostname']}.${tag}
					</record>
				</filter>
				<match **>
					@type cloudwatch_logs
					auto_create_stream true
					region anumber1
					log_group_name_key cw_group_name
					log_stream_name_key cw_stream_name
					remove_log_stream_name_key true
					remove_log_group_name_key true
					auto_create_stream true
					concurrency 2
					aws_key_id "#{open('/var/run/ocp-collector/secrets/my-secret/aws_access_key_id','r') do |f|f.read end}"
					aws_sec_key "#{open('/var/run/ocp-collector/secrets/my-secret/aws_secret_access_key','r') do |f|f.read end}"
					include_time_key true
					log_rejected_request true
				</match>
			</label>`

				results, err := generator.generateOutputLabelBlocks(outputs, secrets, forwarderSpec)
				Expect(err).To(BeNil())
				Expect(len(results)).To(Equal(1))
				Expect(results[0]).To(EqualTrimLines(expConf))
			})
		})
		Context("grouped by namespace UUID", func() {
			BeforeEach(func() {
				output.Cloudwatch.GroupBy = loggingv1.LogGroupByNamespaceUUID
				prefix := "foo"
				output.Cloudwatch.GroupPrefix = &prefix
			})
			It("should provide a valid configuration", func() {
				expConf := `
			<label @MY_CLOUDWATCH>
				<filter kubernetes.**>
					@type record_transformer
					enable_ruby true
					<record>
						cw_group_name foo.${record['kubernetes']['namespace_id']}
						cw_stream_name ${tag}
					</record>
				</filter>
				<filter journal **_default_** **_kube-*_** **_openshift-*_** **_openshift_**>
					@type record_transformer
					enable_ruby true
					<record>
						cw_group_name foo.infrastructure
						cw_stream_name ${record['hostname']}.${tag}
					</record>
				</filter>
				<filter *audit.log>
					@type record_transformer
					enable_ruby true
					<record>
						cw_group_name foo.audit
						cw_stream_name ${record['hostname']}.${tag}
					</record>
				</filter>
				<match **>
					@type cloudwatch_logs
					auto_create_stream true
					region anumber1
					log_group_name_key cw_group_name
					log_stream_name_key cw_stream_name
					remove_log_stream_name_key true
					remove_log_group_name_key true
					auto_create_stream true
					concurrency 2
					aws_key_id "#{open('/var/run/ocp-collector/secrets/my-secret/aws_access_key_id','r') do |f|f.read end}"
					aws_sec_key "#{open('/var/run/ocp-collector/secrets/my-secret/aws_secret_access_key','r') do |f|f.read end}"
					include_time_key true
					log_rejected_request true
				</match>
			</label>`

				results, err := generator.generateOutputLabelBlocks(outputs, secrets, forwarderSpec)
				Expect(err).To(BeNil())
				Expect(len(results)).To(Equal(1))
				Expect(results[0]).To(EqualTrimLines(expConf))
			})
		})
	})
})
