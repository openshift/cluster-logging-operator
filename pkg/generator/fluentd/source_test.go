package fluentd

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	logging "github.com/openshift/cluster-logging-operator/pkg/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/pkg/generator"
	. "github.com/openshift/cluster-logging-operator/test/matchers"
)

var _ = Describe("generating source", func() {

	var (
		g       generator.Generator
		err     error
		results string
		op      generator.Options
		fwd     logging.ClusterLogForwarderSpec
		e       []generator.Element
	)

	BeforeEach(func() {
		g = generator.MakeGenerator()
		fwd = logging.ClusterLogForwarderSpec{
			Pipelines: []logging.PipelineSpec{
				{
					Name:       "pipeline",
					OutputRefs: []string{logging.OutputNameDefault},
				},
			},
		}
		op = generator.NoOptions
	})

	Context("for only logs.app source", func() {
		BeforeEach(func() {
			fwd.Pipelines[0].InputRefs = []string{logging.InputNameApplication}
			e = LogSources(&fwd, op)
			Expect(len(e)).To(Equal(1))
			results, err = g.GenerateConf(e...)
			Expect(err).To(BeNil())
		})

		It("should produce a container config", func() {
			Expect(results).To(EqualTrimLines(`# Logs from containers (including openshift containers)
    <source>
     @type tail
     @id container-input
     path "/var/log/containers/*.log"
     exclude_path ["/var/log/containers/fluentd-*_openshift-logging_*.log", "/var/log/containers/elasticsearch-*_openshift-logging_*.log", "/var/log/containers/kibana-*_openshift-logging_*.log"]
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
    `))
		})
	})

	Context("for only logs.infra source", func() {
		BeforeEach(func() {
			fwd.Pipelines[0].InputRefs = []string{logging.InputNameInfrastructure}
			e = LogSources(&fwd, op)
			Expect(len(e) == 2).To(BeTrue())
		})

		/*
		   "infrastructure" logs include
		    - journal logs
		    - container logs from **_default_** **_kube-*_** **_openshift-*_** **_openshift_** namespaces
		*/

		It("should produce a journal config", func() {
			results, err := g.GenerateConf(e[0])
			Expect(err).To(BeNil())
			Expect(results).To(EqualTrimLines(`
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
    `))
		})
		It("should produce a source container config", func() {
			results, err := g.GenerateConf(e[1])
			Expect(err).To(BeNil())
			Expect(results).To(EqualTrimLines(`
   # Logs from containers (including openshift containers)
   <source>
    @type tail
    @id container-input
    path "/var/log/containers/*.log"
    exclude_path ["/var/log/containers/fluentd-*_openshift-logging_*.log", "/var/log/containers/elasticsearch-*_openshift-logging_*.log", "/var/log/containers/kibana-*_openshift-logging_*.log"]
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
    `))
		})
	})

	Context("for only logs.audit source", func() {
		BeforeEach(func() {
			fwd.Pipelines[0].InputRefs = []string{logging.InputNameAudit}
			e = LogSources(&fwd, op)
			Expect(len(e)).To(Equal(4))
		})

		It("should produce configs for the audit logs", func() {
			results, err := g.GenerateConf(e[0])
			Expect(err).To(BeNil())
			Expect(results).To(EqualTrimLines(`
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
    `))
			results, err = g.GenerateConf(e[1])
			Expect(err).To(BeNil())
			Expect(results).To(EqualTrimLines(`
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
    `))
			results, err = g.GenerateConf(e[2])
			Expect(err).To(BeNil())
			Expect(results).To(EqualTrimLines(`
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
    `))
			results, err = g.GenerateConf(e[3])
			Expect(err).To(BeNil())
			Expect(results).To(EqualTrimLines(`
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
    `))
		})
	})

	Context("for all log sources", func() {

		BeforeEach(func() {
			fwd.Pipelines[0].InputRefs = []string{logging.InputNameApplication, logging.InputNameInfrastructure, logging.InputNameAudit}
			e = LogSources(&fwd, op)
			Expect(len(e)).To(Equal(6))
		})
		Context("for journal input", func() {

			It("should produce a config with no exclusions", func() {
				results, err := g.GenerateConf(e[0])
				Expect(err).To(BeNil())
				Expect(results).To(EqualTrimLines(`
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
   </source>`))
			})
		})

		Context("for container inputs", func() {

			It("should produce a config", func() {
				results, err := g.GenerateConf(e[1])
				Expect(err).To(BeNil())
				Expect(results).To(EqualTrimLines(`# Logs from containers (including openshift containers)
     <source>
    @type tail
    @id container-input
    path "/var/log/containers/*.log"
    exclude_path ["/var/log/containers/fluentd-*_openshift-logging_*.log", "/var/log/containers/elasticsearch-*_openshift-logging_*.log", "/var/log/containers/kibana-*_openshift-logging_*.log"]
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
     `))
			})
		})

		Context("for audit inputs", func() {

			It("should produce a config with no exclusions", func() {
				results, err := g.GenerateConf(e[2])
				Expect(err).To(BeNil())
				Expect(results).To(EqualTrimLines(`
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
      `))
				results, err = g.GenerateConf(e[3])
				Expect(err).To(BeNil())
				Expect(results).To(EqualTrimLines(`
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
      `))
				results, err = g.GenerateConf(e[4])
				Expect(err).To(BeNil())
				Expect(results).To(EqualTrimLines(`
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
      `))
				results, err = g.GenerateConf(e[5])
				Expect(err).To(BeNil())
				Expect(results).To(EqualTrimLines(`
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
      `))
			})
		})
	})

})
