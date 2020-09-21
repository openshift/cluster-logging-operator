package fluentbit

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	logging "github.com/openshift/cluster-logging-operator/pkg/apis/logging/v1"
	. "github.com/openshift/cluster-logging-operator/test/matchers"
	"k8s.io/apimachinery/pkg/util/sets"
)

var _ = Describe("generating source", func() {

	var (
		generator *ConfigGenerator
		err       error
		results   []string
	)

	BeforeEach(func() {
		generator, err = NewConfigGenerator(false, false, true)
		Expect(err).To(BeNil())
	})

	Context("for only application source", func() {
		BeforeEach(func() {
			results, err = generator.generateSource(sets.NewString(logging.InputNameApplication))
			Expect(err).To(BeNil())
			Expect(len(results) == 1).To(BeTrue())
		})

		It("should produce a container config", func() {
			Expect(results[0]).To(EqualTrimLines(`
			[INPUT]
				Name tail
				Path /var/log/containers/*.log
				Path_Key filename
				Parser containerd
				Exclude_Path /var/log/containers/*_openshift*_*.log, /var/log/containers/*_kube*_*.log, /var/log/containers/*_default_*.log
				Tag kubernetes.*
				DB /var/lib/fluent-bit/app-containers.pos.db
				Refresh_Interval 5
		  `))
		})
	})

	Context("for only infra", func() {
		BeforeEach(func() {
			results, err = generator.generateSource(sets.NewString(logging.InputNameInfrastructure))
			Expect(err).To(BeNil())
			Expect(len(results) == 2).To(BeTrue())
		})
		It("should produce a container config", func() {
			Expect(results[1]).To(EqualTrimLines(`
			[INPUT]
				Name tail
				Path /var/log/containers/*_openshift*_*.log, /var/log/containers/*_kube*_*.log, /var/log/containers/*_default_*.log
				Path_Key filename
				Parser containerd
				Exclude_Path /var/log/containers/*_openshift-logging*_*.log
				Tag kubernetes.*
				DB /var/lib/fluent-bit/infra-containers.pos.db
				Refresh_Interval 5
		  `))
		})
		It("should produce a journal config", func() {
			Expect(results[0]).To(EqualTrimLines(`
			[INPUT]
				Name systemd
				Path /var/log/journal
				Tag journal
				DB /var/lib/fluent-bit/journal.pos.db
				Read_From_Tail On
		  `))
		})
	})

	Context("for only audit", func() {
		BeforeEach(func() {
			results, err = generator.generateSource(sets.NewString(logging.InputNameAudit))
			Expect(err).To(BeNil())
			Expect(len(results)).To(Equal(1))
		})

		It("should produce configs for the audit logs", func() {
			Expect(results[0]).To(EqualTrimLines(`
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
		  `))
		})
	})

	Context("for all log sources", func() {

		BeforeEach(func() {
			results, err = generator.generateSource(sets.NewString(logging.InputNameApplication, logging.InputNameInfrastructure, logging.InputNameAudit))
			Expect(err).To(BeNil())
		})

		It("should produce all source config", func() {
			Expect(len(results)).To(Equal(4))
		})
	})

})
