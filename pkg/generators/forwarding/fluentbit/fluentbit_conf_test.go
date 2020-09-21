package fluentbit

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	logging "github.com/openshift/cluster-logging-operator/pkg/apis/logging/v1"
	. "github.com/openshift/cluster-logging-operator/test/matchers"
)

var _ = Describe("Generating fluentbit.conf", func() {
	var (
		forwarder     *logging.ClusterLogForwarderSpec
		forwarderSpec *logging.ForwarderSpec
		generator     *ConfigGenerator
	)
	BeforeEach(func() {
		var err error
		generator, err = NewConfigGenerator(false, false, true)
		Expect(err).To(BeNil())
		Expect(generator).ToNot(BeNil())
		forwarder = &logging.ClusterLogForwarderSpec{
			Outputs: []logging.OutputSpec{
				{
					Type:   logging.OutputTypeElasticsearch,
					Name:   "infra-es",
					URL:    "https://es.svc.infra.cluster:9999",
					Secret: &logging.OutputSecretSpec{Name: "my-infra-secret"},
				},
				{
					Type:   logging.OutputTypeElasticsearch,
					Name:   "apps-es-1",
					URL:    "https://es.svc.messaging.cluster.local:9654",
					Secret: &logging.OutputSecretSpec{Name: "my-es-secret"},
				},
				{
					Type: logging.OutputTypeElasticsearch,
					Name: "apps-es-2",
					URL:  "https://es.svc.messaging.cluster.local2:9654",
					Secret: &logging.OutputSecretSpec{
						Name: "my-other-secret",
					},
				},
				{
					Type: logging.OutputTypeElasticsearch,
					Name: "audit-es",
					URL:  "https://es.svc.audit.cluster:9654",
					Secret: &logging.OutputSecretSpec{
						Name: "my-audit-secret",
					},
				},
			},
			Pipelines: []logging.PipelineSpec{
				{
					Name:       "infra-pipeline",
					InputRefs:  []string{logging.InputNameInfrastructure},
					OutputRefs: []string{"infra-es"},
				},
				{
					Name:       "apps-pipeline",
					InputRefs:  []string{logging.InputNameApplication},
					OutputRefs: []string{"apps-es-1", "apps-es-2"},
				},
				{
					Name:       "audit-pipeline",
					InputRefs:  []string{logging.InputNameAudit},
					OutputRefs: []string{"audit-es"},
				},
			},
		}
	})

	It("should produce well formed configuration", func() {
		results, err := generator.Generate(forwarder, forwarderSpec)
		Expect(err).To(BeNil())
		Expect(results).To(EqualTrimLines(`
		## CLO GENERATED CONFIGURATION ###
		# This file is a generated fluentbit configuration
		# supplied in a configmap.
		[SERVICE]
			Log_Level ${LOG_LEVEL}
			HTTP_Server  On
			HTTP_Listen  ${POD_IP}
			HTTP_PORT    2020
			Parsers_file /etc/fluent-bit/parsers.conf
		[INPUT]
			Name systemd
			Path /var/log/journal
			Tag journal
			DB /var/lib/fluent-bit/journal.pos.db
			Read_From_Tail On
		[INPUT]
			Name tail
			Path /var/log/containers/*_openshift*_*.log, /var/log/containers/*_kube*_*.log, /var/log/containers/*_default_*.log
			Path_Key filename
			Parser containerd
			Exclude_Path /var/log/containers/*_openshift-logging*_*.log
			Tag kubernetes.*
			DB /var/lib/fluent-bit/infra-containers.pos.db
			Refresh_Interval 5
		[INPUT]
			Name tail
			Path /var/log/containers/*.log
			Path_Key filename
			Parser containerd
			Exclude_Path /var/log/containers/*_openshift*_*.log, /var/log/containers/*_kube*_*.log, /var/log/containers/*_default_*.log
			Tag kubernetes.*
			DB /var/lib/fluent-bit/app-containers.pos.db
			Refresh_Interval 5
		[INPUT]
			Name tail
			Path /var/log/audit/audit.log
			Path_Key filename
			Parser json
			Tag linux-audit.log
			DB /var/lib/fluent-bit/audit-linux.pos.db
			Refresh_Interval 5
		[INPUT]
			Name tail
			Path /var/log/kube-apiserver/audit.log
			Path_Key filename
			Parser json
			Tag k8s-audit.log
			DB /var/lib/fluent-bit/audit-k8s.pos.db
			Refresh_Interval 5
		[INPUT]
			Name tail
			Path /var/log/oauth-apiserver/audit.log, /var/log/openshift-apiserver/audit.log
			Path_Key filename
			Parser json
			Tag openshift-audit.log
			DB /var/lib/fluent-bit/audit-oauth.pos.db
			Refresh_Interval 5
		
		[FILTER]
			Name    lua
			Match   kubernetes.*
			script  /etc/fluent-bit/concat-crio.lua
			call reassemble_cri_logs
		
		[OUTPUT]
			Name forward
			Match *
`))
	})
})
