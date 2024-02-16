package fluentd

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	logging "github.com/openshift/cluster-logging-operator/api/logging/v1"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	"github.com/openshift/cluster-logging-operator/internal/generator/framework"
	"github.com/openshift/cluster-logging-operator/internal/tls"
	"github.com/openshift/cluster-logging-operator/test/helpers"
	corev1 "k8s.io/api/core/v1"
)

var _ = Describe("Testing Config Generation", func() {
	var f = func(clspec logging.CollectionSpec, secrets map[string]*corev1.Secret, clfspec logging.ClusterLogForwarderSpec, op framework.Options) []framework.Element {
		var tuning *logging.FluentdInFileSpec
		if clspec.Fluentd != nil && clspec.Fluentd.InFile != nil {
			tuning = clspec.Fluentd.InFile
		}
		return LogSources(&clfspec, tuning, constants.OpenshiftNS, op)
	}
	DescribeTable("Source(s)", helpers.TestGenerateConfWith(f),
		Entry("Only Application", helpers.ConfGenerateTest{
			CLFSpec: logging.ClusterLogForwarderSpec{
				Pipelines: []logging.PipelineSpec{
					{
						InputRefs: []string{
							logging.InputNameApplication,
						},
						OutputRefs: []string{logging.OutputNameDefault},
						Name:       "pipeline",
					},
				},
			},
			ExpectedConf: `
# Logs from containers (including openshift containers)
<source>
  @type tail
  @id container-input
  path "/var/log/pods/*/*/*.log"
  exclude_path ["/var/log/pods/openshift-logging_collector-*/*/*.log", "/var/log/pods/openshift-logging_logfilesmetricexporter-*/*/*.log", "/var/log/pods/openshift-logging_elasticsearch-*/*/*.log", "/var/log/pods/openshift-logging_kibana-*/*/*.log", "/var/log/pods/openshift-logging_*/loki*/*.log", "/var/log/pods/openshift-logging_*/gateway/*.log", "/var/log/pods/openshift-logging_*/opa/*.log", "/var/log/pods/*/*/*.gz", "/var/log/pods/*/*/*.tmp"]
  pos_file "/var/lib/fluentd/pos/es-containers.log.pos"
  follow_inodes true
  refresh_interval 5
  rotate_wait 5
  tag kubernetes.*
  read_from_head "true"
  skip_refresh_on_startup true
  @label @CONCAT
  <parse>
    @type regexp
    expression /^(?<@timestamp>[^\s]+) (?<stream>stdout|stderr) (?<logtag>[F|P]) (?<message>.*)$/
    time_key '@timestamp'
    keep_time_key true
  </parse>
</source>
`,
		}),
		Entry("Only Application with InTail tuning", helpers.ConfGenerateTest{
			CLSpec: logging.CollectionSpec{
				Fluentd: &logging.FluentdForwarderSpec{
					InFile: &logging.FluentdInFileSpec{
						ReadLinesLimit: 1500,
					},
				},
			},
			CLFSpec: logging.ClusterLogForwarderSpec{
				Pipelines: []logging.PipelineSpec{
					{
						InputRefs: []string{
							logging.InputNameApplication,
						},
						OutputRefs: []string{logging.OutputNameDefault},
						Name:       "pipeline",
					},
				},
			},
			ExpectedConf: `
# Logs from containers (including openshift containers)
<source>
  @type tail
  @id container-input
  path "/var/log/pods/*/*/*.log"
  exclude_path ["/var/log/pods/openshift-logging_collector-*/*/*.log", "/var/log/pods/openshift-logging_logfilesmetricexporter-*/*/*.log", "/var/log/pods/openshift-logging_elasticsearch-*/*/*.log", "/var/log/pods/openshift-logging_kibana-*/*/*.log", "/var/log/pods/openshift-logging_*/loki*/*.log", "/var/log/pods/openshift-logging_*/gateway/*.log", "/var/log/pods/openshift-logging_*/opa/*.log", "/var/log/pods/*/*/*.gz", "/var/log/pods/*/*/*.tmp"]
  pos_file "/var/lib/fluentd/pos/es-containers.log.pos"
  follow_inodes true
  refresh_interval 5
  rotate_wait 5
  tag kubernetes.*
  read_from_head "true"
  read_lines_limit 1500
  skip_refresh_on_startup true
  @label @CONCAT
  <parse>
    @type regexp
    expression /^(?<@timestamp>[^\s]+) (?<stream>stdout|stderr) (?<logtag>[F|P]) (?<message>.*)$/
    time_key '@timestamp'
    keep_time_key true
  </parse>
</source>
`,
		}),
		Entry("Only Infrastructure", helpers.ConfGenerateTest{
			CLFSpec: logging.ClusterLogForwarderSpec{
				Pipelines: []logging.PipelineSpec{
					{
						InputRefs: []string{
							logging.InputNameInfrastructure,
						},
						OutputRefs: []string{logging.OutputNameDefault},
						Name:       "pipeline",
					},
				},
			},
			ExpectedConf: `
# Logs from linux journal
<source>
  @type systemd
  @id systemd-input
  @label @INGRESS
  path '/var/log/journal'
  <storage>
    @type local
    persistent true
    # NOTE: if this does not end in .json, fluentd will think it
    # is the name of a directory - see fluentd storage_local.rb
    path '/var/lib/fluentd/pos/journal_pos.json'
  </storage>
  matches "#{ENV['JOURNAL_FILTERS_JSON'] || '[]'}"
  tag journal
  read_from_head "#{if (val = ENV.fetch('JOURNAL_READ_FROM_HEAD','')) && (val.length > 0); val; else 'false'; end}"
</source>

# Logs from containers (including openshift containers)
<source>
  @type tail
  @id container-input
  path "/var/log/pods/*/*/*.log"
  exclude_path ["/var/log/pods/openshift-logging_collector-*/*/*.log", "/var/log/pods/openshift-logging_logfilesmetricexporter-*/*/*.log", "/var/log/pods/openshift-logging_elasticsearch-*/*/*.log", "/var/log/pods/openshift-logging_kibana-*/*/*.log", "/var/log/pods/openshift-logging_*/loki*/*.log", "/var/log/pods/openshift-logging_*/gateway/*.log", "/var/log/pods/openshift-logging_*/opa/*.log", "/var/log/pods/*/*/*.gz", "/var/log/pods/*/*/*.tmp"]
  pos_file "/var/lib/fluentd/pos/es-containers.log.pos"
  follow_inodes true
  refresh_interval 5
  rotate_wait 5
  tag kubernetes.*
  read_from_head "true"
  skip_refresh_on_startup true
  @label @CONCAT
  <parse>
    @type regexp
    expression /^(?<@timestamp>[^\s]+) (?<stream>stdout|stderr) (?<logtag>[F|P]) (?<message>.*)$/
    time_key '@timestamp'
    keep_time_key true
  </parse>
</source>
`,
		}),
		Entry("Only Audit", helpers.ConfGenerateTest{
			CLFSpec: logging.ClusterLogForwarderSpec{
				Pipelines: []logging.PipelineSpec{
					{
						InputRefs: []string{
							logging.InputNameAudit,
						},
						OutputRefs: []string{logging.OutputNameDefault},
						Name:       "pipeline",
					},
				},
			},
			ExpectedConf: `
# linux audit logs
<source>
  @type tail
  @id audit-input
  @label @INGRESS
  path "/var/log/audit/audit.log"
  pos_file "/var/lib/fluentd/pos/audit.log.pos"
  follow_inodes true
  tag linux-audit.log
  <parse>
    @type viaq_host_audit
  </parse>
</source>

# k8s audit logs
<source>
  @type tail
  @id k8s-audit-input
  @label @INGRESS
  path "/var/log/kube-apiserver/audit.log"
  pos_file "/var/lib/fluentd/pos/kube-apiserver.audit.log.pos"
  follow_inodes true
  tag k8s-audit.log
  <parse>
    @type json
    time_key requestReceivedTimestamp
    # In case folks want to parse based on the requestReceivedTimestamp key
    keep_time_key true
    time_format %Y-%m-%dT%H:%M:%S.%N%z
  </parse>
</source>

# Openshift audit logs
<source>
  @type tail
  @id openshift-audit-input
  @label @INGRESS
  path /var/log/oauth-apiserver/audit.log,/var/log/openshift-apiserver/audit.log,/var/log/oauth-server/audit.log
  pos_file /var/lib/fluentd/pos/oauth-apiserver.audit.log
  follow_inodes true
  tag openshift-audit.log
  <parse>
    @type json
    time_key requestReceivedTimestamp
    # In case folks want to parse based on the requestReceivedTimestamp key
    keep_time_key true
    time_format %Y-%m-%dT%H:%M:%S.%N%z
  </parse>
</source>

# Openshift Virtual Network (OVN) audit logs
<source>
  @type tail
  @id ovn-audit-input
  @label @INGRESS
  path "/var/log/ovn/acl-audit-log.log"
  pos_file "/var/lib/fluentd/pos/acl-audit-log.pos"
  follow_inodes true
  tag ovn-audit.log
  refresh_interval 5
  rotate_wait 5
  read_from_head true
  <parse>
    @type none
  </parse>
</source>
`,
		}),
		Entry("Only Audit with InTail tuning", helpers.ConfGenerateTest{
			CLSpec: logging.CollectionSpec{
				Fluentd: &logging.FluentdForwarderSpec{
					InFile: &logging.FluentdInFileSpec{
						ReadLinesLimit: 1500,
					},
				},
			},
			CLFSpec: logging.ClusterLogForwarderSpec{
				Pipelines: []logging.PipelineSpec{
					{
						InputRefs: []string{
							logging.InputNameAudit,
						},
						OutputRefs: []string{logging.OutputNameDefault},
						Name:       "pipeline",
					},
				},
			},
			ExpectedConf: `
# linux audit logs
<source>
  @type tail
  @id audit-input
  @label @INGRESS
  path "/var/log/audit/audit.log"
  pos_file "/var/lib/fluentd/pos/audit.log.pos"
  follow_inodes true
  read_lines_limit 1500
  tag linux-audit.log
  <parse>
    @type viaq_host_audit
  </parse>
</source>

# k8s audit logs
<source>
  @type tail
  @id k8s-audit-input
  @label @INGRESS
  path "/var/log/kube-apiserver/audit.log"
  pos_file "/var/lib/fluentd/pos/kube-apiserver.audit.log.pos"
  follow_inodes true
  read_lines_limit 1500
  tag k8s-audit.log
  <parse>
    @type json
    time_key requestReceivedTimestamp
    # In case folks want to parse based on the requestReceivedTimestamp key
    keep_time_key true
    time_format %Y-%m-%dT%H:%M:%S.%N%z
  </parse>
</source>

# Openshift audit logs
<source>
  @type tail
  @id openshift-audit-input
  @label @INGRESS
  path /var/log/oauth-apiserver/audit.log,/var/log/openshift-apiserver/audit.log,/var/log/oauth-server/audit.log
  pos_file /var/lib/fluentd/pos/oauth-apiserver.audit.log
  follow_inodes true
  read_lines_limit 1500
  tag openshift-audit.log
  <parse>
    @type json
    time_key requestReceivedTimestamp
    # In case folks want to parse based on the requestReceivedTimestamp key
    keep_time_key true
    time_format %Y-%m-%dT%H:%M:%S.%N%z
  </parse>
</source>

# Openshift Virtual Network (OVN) audit logs
<source>
  @type tail
  @id ovn-audit-input
  @label @INGRESS
  path "/var/log/ovn/acl-audit-log.log"
  pos_file "/var/lib/fluentd/pos/acl-audit-log.pos"
  follow_inodes true
  read_lines_limit 1500
  tag ovn-audit.log
  refresh_interval 5
  rotate_wait 5
  read_from_head true
  <parse>
    @type none
  </parse>
</source>
`,
		}),
		Entry("All Log Sources", helpers.ConfGenerateTest{
			CLFSpec: logging.ClusterLogForwarderSpec{
				Pipelines: []logging.PipelineSpec{
					{
						InputRefs: []string{
							logging.InputNameApplication,
							logging.InputNameInfrastructure,
							logging.InputNameAudit,
						},
						OutputRefs: []string{logging.OutputNameDefault},
						Name:       "pipeline",
					},
				},
			},
			ExpectedConf: AllSources,
		}),
		Entry("Exclusion log path should include Elasticsearch and Loki", helpers.ConfGenerateTest{
			CLFSpec: logging.ClusterLogForwarderSpec{
				Pipelines: []logging.PipelineSpec{
					{
						InputRefs: []string{
							logging.InputNameApplication,
						},
						OutputRefs: []string{"es-1", "loki-2"},
						Name:       "pipeline",
					},
				},
				Outputs: []logging.OutputSpec{
					{
						Type: logging.OutputTypeElasticsearch,
						Name: "es-1",
						URL:  "https://es-1.svc.messaging.cluster.local:9200",
						Secret: &logging.OutputSecretSpec{
							Name: "es-1",
						},
					},
					{
						Type: logging.OutputTypeLoki,
						Name: "loki-2",
						URL:  "https://loki-2.svc.messaging.cluster.local:9200",
						Secret: &logging.OutputSecretSpec{
							Name: "loki-2",
						},
					},
				},
			},
			ExpectedConf: `
# Logs from containers (including openshift containers)
<source>
  @type tail
  @id container-input
  path "/var/log/pods/*/*/*.log"
  exclude_path ["/var/log/pods/openshift-logging_collector-*/*/*.log", "/var/log/pods/openshift-logging_logfilesmetricexporter-*/*/*.log", "/var/log/pods/openshift-logging_elasticsearch-*/*/*.log", "/var/log/pods/openshift-logging_kibana-*/*/*.log", "/var/log/pods/openshift-logging_*/loki*/*.log", "/var/log/pods/openshift-logging_*/gateway/*.log", "/var/log/pods/openshift-logging_*/opa/*.log", "/var/log/pods/*/*/*.gz", "/var/log/pods/*/*/*.tmp"]
  pos_file "/var/lib/fluentd/pos/es-containers.log.pos"
  follow_inodes true
  refresh_interval 5
  rotate_wait 5
  tag kubernetes.*
  read_from_head "true"
  skip_refresh_on_startup true
  @label @CONCAT
  <parse>
    @type regexp
    expression /^(?<@timestamp>[^\s]+) (?<stream>stdout|stderr) (?<logtag>[F|P]) (?<message>.*)$/
    time_key '@timestamp'
    keep_time_key true
  </parse>
</source>
`,
		}),
	)
})

const AllSources = `
# Logs from linux journal
<source>
  @type systemd
  @id systemd-input
  @label @INGRESS
  path '/var/log/journal'
  <storage>
    @type local
    persistent true
    # NOTE: if this does not end in .json, fluentd will think it
    # is the name of a directory - see fluentd storage_local.rb
    path '/var/lib/fluentd/pos/journal_pos.json'
  </storage>
  matches "#{ENV['JOURNAL_FILTERS_JSON'] || '[]'}"
  tag journal
  read_from_head "#{if (val = ENV.fetch('JOURNAL_READ_FROM_HEAD','')) && (val.length > 0); val; else 'false'; end}"
</source>

# Logs from containers (including openshift containers)
<source>
  @type tail
  @id container-input
  path "/var/log/pods/*/*/*.log"
  exclude_path ["/var/log/pods/openshift-logging_collector-*/*/*.log", "/var/log/pods/openshift-logging_logfilesmetricexporter-*/*/*.log", "/var/log/pods/openshift-logging_elasticsearch-*/*/*.log", "/var/log/pods/openshift-logging_kibana-*/*/*.log", "/var/log/pods/openshift-logging_*/loki*/*.log", "/var/log/pods/openshift-logging_*/gateway/*.log", "/var/log/pods/openshift-logging_*/opa/*.log", "/var/log/pods/*/*/*.gz", "/var/log/pods/*/*/*.tmp"]
  pos_file "/var/lib/fluentd/pos/es-containers.log.pos"
  follow_inodes true
  refresh_interval 5
  rotate_wait 5
  tag kubernetes.*
  read_from_head "true"
  skip_refresh_on_startup true
  @label @CONCAT
  <parse>
    @type regexp
    expression /^(?<@timestamp>[^\s]+) (?<stream>stdout|stderr) (?<logtag>[F|P]) (?<message>.*)$/
    time_key '@timestamp'
    keep_time_key true
  </parse>
</source>

# linux audit logs
<source>
  @type tail
  @id audit-input
  @label @INGRESS
  path "/var/log/audit/audit.log"
  pos_file "/var/lib/fluentd/pos/audit.log.pos"
  follow_inodes true
  tag linux-audit.log
  <parse>
    @type viaq_host_audit
  </parse>
</source>

# k8s audit logs
<source>
  @type tail
  @id k8s-audit-input
  @label @INGRESS
  path "/var/log/kube-apiserver/audit.log"
  pos_file "/var/lib/fluentd/pos/kube-apiserver.audit.log.pos"
  follow_inodes true
  tag k8s-audit.log
  <parse>
    @type json
    time_key requestReceivedTimestamp
    # In case folks want to parse based on the requestReceivedTimestamp key
    keep_time_key true
    time_format %Y-%m-%dT%H:%M:%S.%N%z
  </parse>
</source>

# Openshift audit logs
<source>
  @type tail
  @id openshift-audit-input
  @label @INGRESS
  path /var/log/oauth-apiserver/audit.log,/var/log/openshift-apiserver/audit.log,/var/log/oauth-server/audit.log
  pos_file /var/lib/fluentd/pos/oauth-apiserver.audit.log
  follow_inodes true
  tag openshift-audit.log
  <parse>
    @type json
    time_key requestReceivedTimestamp
    # In case folks want to parse based on the requestReceivedTimestamp key
    keep_time_key true
    time_format %Y-%m-%dT%H:%M:%S.%N%z
  </parse>
</source>

# Openshift Virtual Network (OVN) audit logs
<source>
  @type tail
  @id ovn-audit-input
  @label @INGRESS
  path "/var/log/ovn/acl-audit-log.log"
  pos_file "/var/lib/fluentd/pos/acl-audit-log.pos"
  follow_inodes true
  tag ovn-audit.log
  refresh_interval 5
  rotate_wait 5
  read_from_head true
  <parse>
    @type none
  </parse>
</source>
`

var _ = Describe("Testing Config Generation", func() {
	var f = func(clspec logging.CollectionSpec, secrets map[string]*corev1.Secret, clfspec logging.ClusterLogForwarderSpec, op framework.Options) []framework.Element {
		return MetricSources(&clfspec, op)
	}
	DescribeTable("Metric Source(s)", helpers.TestGenerateConfWith(f),
		Entry("Any Input", helpers.ConfGenerateTest{
			CLFSpec: logging.ClusterLogForwarderSpec{
				Pipelines: []logging.PipelineSpec{
					{
						InputRefs:  []string{},
						OutputRefs: []string{logging.OutputNameDefault},
						Name:       "pipeline",
					},
				},
			},
			Options: framework.Options{framework.ClusterTLSProfileSpec: tls.GetClusterTLSProfileSpec(nil)},
			ExpectedConf: `
# Prometheus Monitoring
<source>
  @type prometheus
  bind "#{ENV['PROM_BIND_IP']}"
  <transport tls>
    cert_path /etc/collector/metrics/tls.crt
    private_key_path /etc/collector/metrics/tls.key
    min_version TLS1_2
    max_version TLS1_3
    ciphers TLS_AES_128_GCM_SHA256:TLS_AES_256_GCM_SHA384:TLS_CHACHA20_POLY1305_SHA256:ECDHE-ECDSA-AES128-GCM-SHA256:ECDHE-RSA-AES128-GCM-SHA256:ECDHE-ECDSA-AES256-GCM-SHA384:ECDHE-RSA-AES256-GCM-SHA384:ECDHE-ECDSA-CHACHA20-POLY1305:ECDHE-RSA-CHACHA20-POLY1305:DHE-RSA-AES128-GCM-SHA256:DHE-RSA-AES256-GCM-SHA384
  </transport>
</source>

<source>
  @type prometheus_monitor
  <labels>
    hostname ${hostname}
  </labels>
</source>

# excluding prometheus_tail_monitor
# since it leaks namespace/pod info
# via file paths

# tail_monitor plugin which publishes log_collected_bytes_total
<source>
  @type collected_tail_monitor
  <labels>
    hostname ${hostname}
  </labels>
</source>

# This is considered experimental by the repo
<source>
  @type prometheus_output_monitor
  <labels>
    hostname ${hostname}
  </labels>
</source>`,
		}))
})
