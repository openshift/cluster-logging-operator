package fluent

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	logging "github.com/openshift/cluster-logging-operator/apis/logging/v1"
	framework "github.com/openshift/cluster-logging-operator/test/framework/e2e"
	"github.com/openshift/cluster-logging-operator/test/helpers"
	testruntime "github.com/openshift/cluster-logging-operator/test/runtime"
	apps "k8s.io/api/apps/v1"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

const (
	fluentConf = `
<system>
  log_level info
</system>
<source>
  @type http
  port 24224
  bind 0.0.0.0
  body_size_limit 32m
  keepalive_timeout 10s
  # Headers are capitalized, and added with prefix "HTTP_"
  add_http_headers true
  add_remote_addr true
  <parse>
    @type json
  </parse>
  <transport tls>
	  ca_path /etc/fluentd/secrets/ca-bundle.crt
	  cert_path /etc/fluentd/secrets/tls.crt
	  private_key_path /etc/fluentd/secrets/tls.key
  </transport>
</source>

<match logs.app>
  @type file
  append true
  path /tmp/app.logs
  symlink_path /tmp/app-logs
</match>
<match logs.infra>
  @type file
  append true
  path /tmp/infra.logs
  symlink_path /tmp/infra-logs
</match>
<match logs.audit>
  @type file
  append true
  path /tmp/audit.logs
  symlink_path /tmp/audit-logs
</match>
<match **>
	@type stdout
</match>
`
)

var _ = Describe("[ClusterLogForwarder] Forwards logs", func() {
	var (
		err              error
		e2e              = framework.NewE2ETestFramework()
		wd, _            = os.Getwd()
		rootDir          = fmt.Sprintf("%s/../../../../", wd)
		forwarder        *logging.ClusterLogForwarder
		logGenNS         string
		fluentDeployment *apps.Deployment
		headers          = map[string]string{"h1": "v1", "h2": "v2"}
		fwdSpec          = logging.ClusterLogForwarderSpec{
			Outputs: []logging.OutputSpec{
				{
					Name: "httpout-app",
					Type: "http",
					// Receiving fluentd instance will receive these logs under tag logs.app
					URL: "https://fluent-receiver.openshift-logging.svc:24224/logs/app",
					OutputTypeSpec: logging.OutputTypeSpec{
						Http: &logging.Http{
							Headers: headers,
							Method:  "POST",
						},
					},
					Secret: &logging.OutputSecretSpec{
						Name: "fluent-receiver",
					},
					TLS: &logging.OutputTLSSpec{
						InsecureSkipVerify: true,
					},
				},
				{
					Name: "httpout-infra",
					Type: "http",
					// Receiving fluentd instance will receive these logs under tag logs.infra
					URL: "https://fluent-receiver.openshift-logging.svc:24224/logs/infra",
					OutputTypeSpec: logging.OutputTypeSpec{
						Http: &logging.Http{
							Headers: headers,
							Method:  "POST",
						},
					},
					Secret: &logging.OutputSecretSpec{
						Name: "fluent-receiver",
					},
					TLS: &logging.OutputTLSSpec{
						InsecureSkipVerify: true,
					},
				},
				{
					Name: "httpout-audit",
					Type: "http",
					// Receiving fluentd instance will receive these logs under tag logs.audit
					URL: "https://fluent-receiver.openshift-logging.svc:24224/logs/audit",
					OutputTypeSpec: logging.OutputTypeSpec{
						Http: &logging.Http{
							Headers: headers,
							Method:  "POST",
						},
					},
					Secret: &logging.OutputSecretSpec{
						Name: "fluent-receiver",
					},
					TLS: &logging.OutputTLSSpec{
						InsecureSkipVerify: true,
					},
				},
			},
			Pipelines: []logging.PipelineSpec{
				{
					Name:       "app-logs",
					OutputRefs: []string{"httpout-app"},
					InputRefs:  []string{"application"},
				},
				{
					Name:       "infra-logs",
					OutputRefs: []string{"httpout-infra"},
					InputRefs:  []string{"infrastructure"},
				},
				{
					Name:       "audit-logs",
					OutputRefs: []string{"httpout-audit"},
					InputRefs:  []string{"audit"},
				},
			},
		}
	)
	Describe("with vector collector", func() {
		BeforeEach(func() {
			cr := helpers.NewClusterLogging(helpers.ComponentTypeCollectorVector)
			if err = e2e.CreateClusterLogging(cr); err != nil {
				Fail(fmt.Sprintf("Unable to create an instance of cluster logging: %v", err))
			}
			forwarder = testruntime.NewClusterLogForwarder()
			forwarder.Spec = fwdSpec
			if logGenNS, err = e2e.DeployLogGenerator(); err != nil {
				Fail(fmt.Sprintf("unable to deploy log generator %v.", err))
			}
		})
		It("send logs to fluentd http", func() {
			fluentDeployment, err = e2e.DeployFluentdReceiverWithConf(rootDir, true, fluentConf)
			Expect(err).To(BeNil())
			if err := e2e.CreateClusterLogForwarder(forwarder); err != nil {
				Fail(fmt.Sprintf("Unable to create an instance of logforwarder: %v", err))
			}
			components := []helpers.LogComponentType{helpers.ComponentTypeCollector}
			for _, component := range components {
				if err := e2e.WaitFor(component); err != nil {
					Fail(fmt.Sprintf("Failed waiting for component %s to be ready: %v", component, err))
				}
			}

			logStore := e2e.LogStores[fluentDeployment.GetName()]
			has, err := logStore.HasApplicationLogs(framework.DefaultWaitForLogsTimeout)
			Expect(err).To(BeNil())
			Expect(has).To(BeTrue())

			has, err = logStore.HasInfraStructureLogs(framework.DefaultWaitForLogsTimeout)
			Expect(err).To(BeNil())
			Expect(has).To(BeTrue())

			has, err = logStore.HasAuditLogs(framework.DefaultWaitForLogsTimeout)
			Expect(err).To(BeNil())
			Expect(has).To(BeTrue())

			collectedLogs, err := logStore.RetrieveLogs()
			Expect(err).To(BeNil())
			appLogsStr, ok := collectedLogs["app"]
			fmt.Printf("---- %s\n", appLogsStr)
			Expect(ok).To(BeTrue())
			appLogs := map[string]interface{}{}
			err = json.Unmarshal([]byte(appLogsStr), &appLogs)
			Expect(err).To(BeNil(), appLogsStr)
			for k, v := range headers {
				// Headers are capitalized, and sent with prefix "HTTP_"
				headerVal, ok := appLogs[fmt.Sprintf("HTTP_%s", strings.ToUpper(k))]
				Expect(ok).To(BeTrue())
				Expect(headerVal).To(Equal(v))
			}
			message, ok := appLogs["message"]
			Expect(ok).To(BeTrue())
			Expect(message).To(ContainSubstring("My life is my message"))
		})
		AfterEach(func() {
			e2e.Cleanup()
			e2e.WaitForCleanupCompletion(logGenNS, []string{"test"})
		})
	})
})
