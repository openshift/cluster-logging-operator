package fluentd

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	logging "github.com/openshift/cluster-logging-operator/pkg/apis/logging/v1"
	test "github.com/openshift/cluster-logging-operator/test"
	"k8s.io/apimachinery/pkg/util/sets"
)

var _ = Describe("generating source", func() {

	var (
		generator *ConfigGenerator
		err       error
		results   []string
	)

	BeforeEach(func() {
		generator, err = NewConfigGenerator()
		Expect(err).To(BeNil())
	})

	Context("for only logs.app source", func() {
		BeforeEach(func() {
			results, err = generator.generateSource(sets.NewString(string(logging.LogSourceTypeApp)))
			Expect(err).To(BeNil())
			Expect(len(results) == 1).To(BeTrue())
		})

		It("should produce a container config with no exclusions", func() {
			test.Expect(results[0]).ToEqual(`# container logs
		  <source>
			@type tail
			@id container-input
			path "/var/log/containers/*.log"
			pos_file "/var/log/es-containers.log.pos"
			refresh_interval 5
			rotate_wait 5
			tag kubernetes.*
			read_from_head "true"
			exclude_path []
			@label @CONCAT
			<parse>
			  @type multi_format
			  <pattern>
				format json
				time_format \'%Y-%m-%dT%H:%M:%S.%N%Z\'
				keep_time_key true
			  </pattern>
			  <pattern>
				format regexp
				expression /^(?<time>.+) (?<stream>stdout|stderr)( (?<logtag>.))? (?<log>.*)$/
				time_format \'%Y-%m-%dT%H:%M:%S.%N%:z\'
				keep_time_key true
			  </pattern>
			</parse>
		  </source>
		  `)
		})
	})

	Context("for only logs.infra source", func() {
		BeforeEach(func() {
			results, err = generator.generateSource(sets.NewString(string(logging.LogSourceTypeInfra)))
			Expect(err).To(BeNil())
			Expect(len(results) == 1).To(BeTrue())
		})

		It("should produce a journal config", func() {
			test.Expect(results[0]).ToEqual(`
			#journal logs to gather node
			<source>
				@type systemd
				@id systemd-input
				@label @INGRESS
				path "#{if (val = ENV.fetch('JOURNAL_SOURCE','')) && (val.length > 0); val; else '/run/log/journal'; end}"
				<storage>
				@type local
				persistent true
				# NOTE: if this does not end in .json, fluentd will think it
				# is the name of a directory - see fluentd storage_local.rb
				path "#{ENV['JOURNAL_POS_FILE'] || '/var/log/journal_pos.json'}"
				</storage>
				matches "#{ENV['JOURNAL_FILTERS_JSON'] || '[]'}"
				tag journal
				read_from_head "#{if (val = ENV.fetch('JOURNAL_READ_FROM_HEAD','')) && (val.length > 0); val; else 'false'; end}"
			</source>
		  `)
		})
	})

	Context("for all log sources", func() {

		BeforeEach(func() {
			results, err = generator.generateSource(sets.NewString(string(logging.LogSourceTypeApp), string(logging.LogSourceTypeInfra)))
			Expect(err).To(BeNil())
			Expect(len(results) == 2).To(BeTrue())
		})
		Context("for journal input", func() {

			It("should produce a config with no exclusions", func() {
				test.Expect(results[0]).ToEqual(`
			#journal logs to gather node
			<source>
				@type systemd
				@id systemd-input
				@label @INGRESS
				path "#{if (val = ENV.fetch('JOURNAL_SOURCE','')) && (val.length > 0); val; else '/run/log/journal'; end}"
				<storage>
				@type local
				persistent true
				# NOTE: if this does not end in .json, fluentd will think it
				# is the name of a directory - see fluentd storage_local.rb
				path "#{ENV['JOURNAL_POS_FILE'] || '/var/log/journal_pos.json'}"
				</storage>
				matches "#{ENV['JOURNAL_FILTERS_JSON'] || '[]'}"
				tag journal
				read_from_head "#{if (val = ENV.fetch('JOURNAL_READ_FROM_HEAD','')) && (val.length > 0); val; else 'false'; end}"
			</source>`)
			})
		})

		Context("for container inputs", func() {

			It("should produce a config with no exclusions", func() {
				test.Expect(results[1]).ToEqual(`# container logs
			  <source>
				@type tail
				@id container-input
				path "/var/log/containers/*.log"
				pos_file "/var/log/es-containers.log.pos"
				refresh_interval 5
				rotate_wait 5
				tag kubernetes.*
				read_from_head "true"
				exclude_path []
				@label @CONCAT
				<parse>
				  @type multi_format
				  <pattern>
					format json
					time_format \'%Y-%m-%dT%H:%M:%S.%N%Z\'
					keep_time_key true
				  </pattern>
				  <pattern>
					format regexp
					expression /^(?<time>.+) (?<stream>stdout|stderr)( (?<logtag>.))? (?<log>.*)$/
					time_format \'%Y-%m-%dT%H:%M:%S.%N%:z\'
					keep_time_key true
				  </pattern>
				</parse>
			  </source>
			  `)
			})
		})
	})

})
