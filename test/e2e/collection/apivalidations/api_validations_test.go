package apivalidations

import (
	"bytes"
	_ "embed"
	"fmt"
	"os/exec"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	framework "github.com/openshift/cluster-logging-operator/test/framework/e2e"
	"github.com/openshift/cluster-logging-operator/test/helpers/cmd"
)

var _ = Describe("", func() {

	const name = "clf-validation-test"
	var (
		e2e *framework.E2ETestFramework
	)
	AfterEach(func() {
		if e2e != nil {
			e2e.Cleanup()
		}
	})

	DescribeTable("Verifying declarative API validations", func(crFile string, assert func(string, error)) {

		e2e = framework.NewE2ETestFramework()
		crYaml, err := tomlContent.ReadFile(crFile)
		if err != nil {
			Fail(fmt.Sprintf("Error reading the file %q with exp config: %v", crFile, err))
		}
		deployNS := e2e.Test.NS.Name
		if _, err := e2e.BuildAuthorizationFor(deployNS, name).
			AllowClusterRole(framework.ClusterRoleCollectApplicationLogs).
			AllowClusterRole(framework.ClusterRoleCollectInfrastructureLogs).
			AllowClusterRole(framework.ClusterRoleCollectAuditLogs).Create(); err != nil {
			Fail(err.Error())
		}

		execCMD := exec.Command("sh", "-c", fmt.Sprintf("echo '%s' | oc -n %s create -f -", crYaml, deployNS))
		reader, err := cmd.NewReader(execCMD)
		Expect(err).ToNot(HaveOccurred())
		defer reader.Close()
		buffer := bytes.NewBuffer([]byte{})
		_, err = buffer.ReadFrom(reader)
		assert(buffer.String(), err)
	},
		Entry("should pass for LokiStack with empty tuning", "lokistack-empty-tuning.yaml", func(out string, err error) {
			Expect(err).ToNot(HaveOccurred())
		}),
		Entry("should fail for LokiStack with snappy compression", "lokistack-snappy-compression-otel.yaml", func(out string, err error) {
			Expect(err.Error()).To(MatchRegexp(".'snappy' compression cannot be used when data model is 'Otel'"))
		}),
		Entry("should pass for syslog with valid udp URL", "syslog_valid_url_udp.yaml", func(out string, err error) {
			Expect(err).ToNot(HaveOccurred())
		}),
		Entry("should pass for syslog with valid tls URL", "syslog_valid_url_tls.yaml", func(out string, err error) {
			Expect(err).ToNot(HaveOccurred())
		}),
		Entry("should pass for syslog with valid tcp URL", "syslog_valid_url_tcp.yaml", func(out string, err error) {
			Expect(err).ToNot(HaveOccurred())
		}),
		Entry("should pass for kafka with valid URL or brokers", "kafka_valid_url_and_brokers.yaml", func(out string, err error) {
			Expect(err).ToNot(HaveOccurred())
		}),
		Entry("should fail for kafka without URL or brokers", "kafka_no_url_or_brokers.yaml", func(out string, err error) {
			Expect(err.Error()).To(MatchRegexp(".*URL.*brokers.*required.*"))
		}),
		Entry("should fail for kafka with invalid URL", "kafka_invalid_url.yaml", func(out string, err error) {
			Expect(err.Error()).To(MatchRegexp("must be a valid URL with a tcp or tls scheme"))
		}),
		Entry("should fail for kafka invalid broker URL", "kafka_invalid_broker_url.yaml", func(out string, err error) {
			Expect(err).To(HaveOccurred())
			//occurrenceCount := strings.Count(err.Error(), "each broker must be a valid URL with a tcp or tls scheme")
			//Expect(occurrenceCount).To(Equal(2), "expect validation error appear twice")
			Expect(err.Error()).To(ContainSubstring("each broker must be a valid URL with a tcp or tls scheme"))
		}),
		Entry("LOG-5788: for multilineException filter should not fail", "log5788_mulitiline_ex_filter.yaml", func(out string, err error) {
			Expect(err).ToNot(HaveOccurred())
			Expect(out).To(MatchRegexp("clusterlogforwarder.*created"))
		}),
		Entry("LOG-5793: for lokiStack bearer token from SA should not fail", "log5793_bearer_token_from_sa.yaml", func(out string, err error) {
			Expect(err).ToNot(HaveOccurred())
			Expect(out).To(MatchRegexp("clusterlogforwarder.*created"))
		}),
		Entry("should fail with invalid name", "invalid_name.yaml", func(out string, err error) {
			Expect(err.Error()).To(MatchRegexp("Name.*valid DNS1035"))
		}),
		Entry("should pass for Cloudwatch with no URL", "cloudwatch-no-url.yaml", func(out string, err error) {
			Expect(err).ToNot(HaveOccurred())
		}),
		Entry("should fail for Cloudwatch with invalid URL", "cloudwatch-invalid-url.yaml", func(out string, err error) {
			Expect(err.Error()).To(MatchRegexp("invalid URL"))
		}),
		Entry("should pass for Cloudwatch with empty URL", "cloudwatch-empty-url.yaml", func(out string, err error) {
			Expect(err).ToNot(HaveOccurred())
		}),
		Entry("should fail for Cloudwatch with invalid characters in external_id", "cw-assume-role-ext-id.yaml", func(out string, err error) {
			Expect(err.Error()).To(MatchRegexp("Invalid value"))
		}),
		Entry("should fail for Cloudwatch with invalid external_id", "cw-assume-role-ext-id-bad.yaml", func(out string, err error) {
			Expect(err.Error()).To(MatchRegexp("should be at least 2 chars"))
		}),
		Entry("should pass for Cloudwatch with allowed external_id characters", "cw-assume-role-ext-id-chars.yaml", func(out string, err error) {
			Expect(err).ToNot(HaveOccurred())
		}),
		Entry("should pass for OTLP with any http or https URL", "otlp_valid_url.yaml", func(out string, err error) {
			Expect(err).ToNot(HaveOccurred())
		}),
		Entry("should fail for OTLP with non http URL", "otlp_valid_non_http.yaml", func(out string, err error) {
			Expect(err).To(HaveOccurred())
		}),
	)
})
