package splunk

import (
	"fmt"
	"strings"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	internalobs "github.com/openshift/cluster-logging-operator/internal/api/observability"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	"github.com/openshift/cluster-logging-operator/internal/runtime"
	"github.com/openshift/cluster-logging-operator/test/framework/functional"
	"github.com/openshift/cluster-logging-operator/test/helpers/splunk"
	obstestruntime "github.com/openshift/cluster-logging-operator/test/runtime/observability"
	v1 "k8s.io/api/core/v1"
)

var _ = Describe("Forwarding to Splunk with authorization failures", func() {
	var (
		framework    *functional.CollectorFunctionalFramework
		secret       *v1.Secret
		hecSecretKey = *internalobs.NewSecretReference(constants.SplunkHECTokenKey, SplunkSecretName)
	)

	BeforeEach(func() {
		framework = functional.NewCollectorFunctionalFramework()
		// Create a secret with an invalid/wrong HEC token that Splunk will reject with 403
		// Note: The Splunk server itself is configured with functional.HecToken (valid token)
		// but the collector will use this invalid token from the secret
		secret = runtime.NewSecret(framework.Namespace, SplunkSecretName,
			map[string][]byte{
				"hecToken": []byte("invalid-token-12345"),
			},
		)
	})

	AfterEach(func() {
		framework.Cleanup()
	})

	It("should log forbidden error and report 403 in vector_http_client_responses_total metric when HEC token is invalid", func() {
		// Configure ClusterLogForwarder to send to Splunk with invalid token
		obstestruntime.NewClusterLogForwarderBuilder(framework.Forwarder).
			FromInput(obs.InputTypeApplication).
			ToSplunkOutput(hecSecretKey, func(output *obs.OutputSpec) {
				output.Splunk.Index = "main"
			})

		framework.Secrets = append(framework.Secrets, secret)

		// Remove auth from prometheus metrics endpoint so we can query metrics without auth
		framework.VisitConfig = func(conf string) string {
			// Remove the [sinks.prometheus_output.auth] section
			lines := strings.Split(conf, "\n")
			var result []string
			skipNext := false
			for _, line := range lines {
				// Skip the [sinks.prometheus_output.auth] section and its properties
				if strings.Contains(line, "[sinks.prometheus_output.auth]") {
					skipNext = true
					continue
				}
				if skipNext {
					// Skip lines that are part of the auth section (strategy, path, verb)
					trimmed := strings.TrimSpace(line)
					if trimmed == `strategy = "sar"` || trimmed == `path = "/metrics"` || trimmed == `verb = "get"` {
						continue
					}
					// Stop skipping when we hit a non-auth line or another section
					if !strings.HasPrefix(trimmed, "strategy") && !strings.HasPrefix(trimmed, "path") && !strings.HasPrefix(trimmed, "verb") {
						skipNext = false
					}
				}
				if !skipNext {
					result = append(result, line)
				}
			}
			return strings.Join(result, "\n")
		}

		// Deploy with real Splunk container
		// The Splunk server will be configured with functional.HecToken (the valid token)
		// but the collector will use "invalid-token-12345", causing 403 responses
		Expect(framework.Deploy()).To(BeNil())

		// Wait for Splunk to be ready
		splunk.WaitOnSplunk(framework)

		// Write application logs
		timestamp := "2020-11-04T18:13:59.061892+00:00"
		applicationLogLine := functional.NewCRIOLogMessage(timestamp, "This is my test message", false)
		Expect(framework.WriteMessagesToApplicationLog(applicationLogLine, 2)).To(BeNil())

		// Wait for logs to be attempted to be sent and retried
		time.Sleep(15 * time.Second)

		// Read collector logs and verify forbidden error is logged
		collectorLog, err := framework.ReadCollectorLogs()
		Expect(err).To(BeNil(), "Expected no errors reading collector logs")
		Expect(collectorLog).ToNot(BeEmpty(), "Expected collector logs to not be empty")

		// Verify that the collector logs contain a forbidden error message
		// Vector logs HTTP errors with the status code
		Expect(collectorLog).To(
			Or(
				ContainSubstring("forbidden"),
				ContainSubstring("Forbidden"),
				ContainSubstring("status: 403"),
				ContainSubstring("status=403"),
				And(
					ContainSubstring("error"),
					ContainSubstring("403"),
				),
			),
			"Expected collector logs to contain forbidden error or 403 status",
		)

		// Query metrics from the collector (no auth required since we removed it)
		metricsURL := fmt.Sprintf("https://%s.%s:24231/metrics", framework.Name, framework.Namespace)
		curlCmd := fmt.Sprintf(`curl -ksv %s`, metricsURL)
		metrics, err := framework.RunCommand(constants.CollectorName, "sh", "-c", curlCmd)
		Expect(err).To(BeNil(), "Expected no errors querying metrics")
		Expect(metrics).ToNot(BeEmpty(), "Expected metrics to not be empty")

		// Verify vector_http_client_responses_total metric exists
		Expect(metrics).To(ContainSubstring("vector_http_client_responses_total"), "Expected to find vector_http_client_responses_total metric")

		// Look for the metric line with status="403"
		metricLines := strings.Split(metrics, "\n")
		found403Metric := false
		for _, line := range metricLines {
			// Skip comment lines
			if strings.HasPrefix(strings.TrimSpace(line), "#") {
				continue
			}
			// Look for vector_http_client_responses_total with status="403"
			if strings.Contains(line, "vector_http_client_responses_total") &&
				strings.Contains(line, `status="403"`) {
				found403Metric = true
				// Log the metric line for debugging
				break
			}
		}
		Expect(found403Metric).To(BeTrue(), "Expected to find vector_http_client_responses_total metric with status=\"403\" label")
	})
})
