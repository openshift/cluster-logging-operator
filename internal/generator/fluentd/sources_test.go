package fluentd

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	logging "github.com/openshift/cluster-logging-operator/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/internal/generator"
	corev1 "k8s.io/api/core/v1"
)

var _ = Describe("Testing Config Generation", func() {
	var f = func(clspec logging.ClusterLoggingSpec, secrets map[string]*corev1.Secret, clfspec logging.ClusterLogForwarderSpec, op generator.Options) []generator.Element {
		return LogSources(&clfspec, op)
	}
	DescribeTable("Source(s)", generator.TestGenerateConfWith(f),
		Entry("Only Application", generator.ConfGenerateTest{
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
  path "/var/log/pods/**/*.log"
  exclude_path ["/var/log/pods/openshift-logging_collector-*/*/*.log", "/var/log/pods/openshift-logging_elasticsearch-*/*/*.log", "/var/log/pods/openshift-logging_kibana-*/*/*.log"]
  pos_file "/var/lib/fluentd/pos/es-containers.log.pos"
  refresh_interval 5
  rotate_wait 5
  tag kubernetes.*
  read_from_head "true"
  skip_refresh_on_startup true
  @label @MEASURE
  <parse>
    @type multi_format
    <pattern>
      format json
      time_format '%Y-%m-%dT%H:%M:%S.%N%Z'
      keep_time_key true
    </pattern>
    <pattern>
      format regexp
      expression /^(?<time>[^\s]+) (?<stream>stdout|stderr)( (?<logtag>.))? (?<log>.*)$/
      time_format '%Y-%m-%dT%H:%M:%S.%N%:z'
      keep_time_key true
    </pattern>
  </parse>
</source>
`,
		}),
		Entry("Only Infrastructure", generator.ConfGenerateTest{
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
  @label @MEASURE
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
  path "/var/log/pods/**/*.log"
  exclude_path ["/var/log/pods/openshift-logging_collector-*/*/*.log", "/var/log/pods/openshift-logging_elasticsearch-*/*/*.log", "/var/log/pods/openshift-logging_kibana-*/*/*.log"]
  pos_file "/var/lib/fluentd/pos/es-containers.log.pos"
  refresh_interval 5
  rotate_wait 5
  tag kubernetes.*
  read_from_head "true"
  skip_refresh_on_startup true
  @label @MEASURE
  <parse>
    @type multi_format
    <pattern>
      format json
      time_format '%Y-%m-%dT%H:%M:%S.%N%Z'
      keep_time_key true
    </pattern>
    <pattern>
      format regexp
      expression /^(?<time>[^\s]+) (?<stream>stdout|stderr)( (?<logtag>.))? (?<log>.*)$/
      time_format '%Y-%m-%dT%H:%M:%S.%N%:z'
      keep_time_key true
    </pattern>
  </parse>
</source>
`,
		}),
		Entry("Only Audit", generator.ConfGenerateTest{
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
  @label @MEASURE
  path "/var/log/audit/audit.log"
  pos_file "/var/lib/fluentd/pos/audit.log.pos"
  tag linux-audit.log
  <parse>
    @type viaq_host_audit
  </parse>
</source>

# k8s audit logs
<source>
  @type tail
  @id k8s-audit-input
  @label @MEASURE
  path "/var/log/kube-apiserver/audit.log"
  pos_file "/var/lib/fluentd/pos/kube-apiserver.audit.log.pos"
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
  @label @MEASURE
  path /var/log/oauth-apiserver/audit.log,/var/log/openshift-apiserver/audit.log
  pos_file /var/lib/fluentd/pos/oauth-apiserver.audit.log
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
  @label @MEASURE
  path "/var/log/ovn/acl-audit-log.log"
  pos_file "/var/lib/fluentd/pos/acl-audit-log.pos"
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
		Entry("All Log Sources", generator.ConfGenerateTest{
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
	)
})

const AllSources = `
# Logs from linux journal
<source>
  @type systemd
  @id systemd-input
  @label @MEASURE
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
  path "/var/log/pods/**/*.log"
  exclude_path ["/var/log/pods/openshift-logging_collector-*/*/*.log", "/var/log/pods/openshift-logging_elasticsearch-*/*/*.log", "/var/log/pods/openshift-logging_kibana-*/*/*.log"]
  pos_file "/var/lib/fluentd/pos/es-containers.log.pos"
  refresh_interval 5
  rotate_wait 5
  tag kubernetes.*
  read_from_head "true"
  skip_refresh_on_startup true
  @label @MEASURE
  <parse>
    @type multi_format
    <pattern>
      format json
      time_format '%Y-%m-%dT%H:%M:%S.%N%Z'
      keep_time_key true
    </pattern>
    <pattern>
      format regexp
      expression /^(?<time>[^\s]+) (?<stream>stdout|stderr)( (?<logtag>.))? (?<log>.*)$/
      time_format '%Y-%m-%dT%H:%M:%S.%N%:z'
      keep_time_key true
    </pattern>
  </parse>
</source>

# linux audit logs
<source>
  @type tail
  @id audit-input
  @label @MEASURE
  path "/var/log/audit/audit.log"
  pos_file "/var/lib/fluentd/pos/audit.log.pos"
  tag linux-audit.log
  <parse>
    @type viaq_host_audit
  </parse>
</source>

# k8s audit logs
<source>
  @type tail
  @id k8s-audit-input
  @label @MEASURE
  path "/var/log/kube-apiserver/audit.log"
  pos_file "/var/lib/fluentd/pos/kube-apiserver.audit.log.pos"
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
  @label @MEASURE
  path /var/log/oauth-apiserver/audit.log,/var/log/openshift-apiserver/audit.log
  pos_file /var/lib/fluentd/pos/oauth-apiserver.audit.log
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
  @label @MEASURE
  path "/var/log/ovn/acl-audit-log.log"
  pos_file "/var/lib/fluentd/pos/acl-audit-log.pos"
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
	var f = func(clspec logging.ClusterLoggingSpec, secrets map[string]*corev1.Secret, clfspec logging.ClusterLogForwarderSpec, op generator.Options) []generator.Element {
		return MetricSources(&clfspec, op)
	}
	DescribeTable("Metric Source(s)", generator.TestGenerateConfWith(f),
		Entry("Any Input", generator.ConfGenerateTest{
			CLFSpec: logging.ClusterLogForwarderSpec{
				Pipelines: []logging.PipelineSpec{
					{
						InputRefs:  []string{},
						OutputRefs: []string{logging.OutputNameDefault},
						Name:       "pipeline",
					},
				},
			},
			ExpectedConf: `
# Prometheus Monitoring
<source>
  @type prometheus
  bind "#{ENV['POD_IP']}"
  <ssl>
    enable true
    certificate_path "#{ENV['METRICS_CERT'] || '/etc/fluent/metrics/tls.crt'}"
    private_key_path "#{ENV['METRICS_KEY'] || '/etc/fluent/metrics/tls.key'}"
  </ssl>
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
