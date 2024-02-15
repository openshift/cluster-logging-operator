package collection

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	loggingv1 "github.com/openshift/cluster-logging-operator/api/logging/v1"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	framework "github.com/openshift/cluster-logging-operator/test/framework/e2e"
	"github.com/openshift/cluster-logging-operator/test/helpers"
	testruntime "github.com/openshift/cluster-logging-operator/test/runtime"
	apps "k8s.io/api/apps/v1"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Test collector deployment type", func() {
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
		receiverPort = 8080
		receiverName = "http-audit"
	)
	var (
		err                  error
		wd, _                = os.Getwd()
		rootDir              = fmt.Sprintf("%s/../../../../", wd)
		forwarder            *loggingv1.ClusterLogForwarder
		deploymentAnnotation = map[string]string{constants.AnnotationEnableCollectorAsDeployment: "true"}
		fluentDeployment     *apps.Deployment
		logGenNS             string
		headers              = map[string]string{"h1": "v1", "h2": "v2"}
		e2e                  = framework.NewE2ETestFramework()
		fwdSpec              = loggingv1.ClusterLogForwarderSpec{
			Inputs: []loggingv1.InputSpec{
				{
					Name: receiverName,
					Receiver: &loggingv1.ReceiverSpec{
						Type: loggingv1.ReceiverTypeHttp,
						ReceiverTypeSpec: &loggingv1.ReceiverTypeSpec{
							HTTP: &loggingv1.HTTPReceiver{
								Format: loggingv1.FormatKubeAPIAudit,
								Port:   receiverPort,
							},
						},
					},
				},
			},
			Outputs: []loggingv1.OutputSpec{
				{
					Name: "httpout-audit",
					Type: loggingv1.OutputTypeHttp,
					URL:  "https://fluent-receiver.openshift-logging.svc:24224/logs/audit",
					OutputTypeSpec: loggingv1.OutputTypeSpec{
						Http: &loggingv1.Http{
							Headers: headers,
							Method:  "POST",
						},
					},
					Secret: &loggingv1.OutputSecretSpec{
						Name: "fluent-receiver",
					},
					TLS: &loggingv1.OutputTLSSpec{
						InsecureSkipVerify: true,
					},
				},
			},
			Pipelines: []loggingv1.PipelineSpec{
				{
					Name:       "input-receiver-logs",
					OutputRefs: []string{"httpout-audit"},
					InputRefs:  []string{"http-audit"},
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
			httpReceiverServiceName := fmt.Sprintf("%s-%s", constants.CollectorName, receiverName)
			httpReceiverEndpoint := fmt.Sprintf("https://%s.%s.svc.cluster.local:%d", httpReceiverServiceName, constants.OpenshiftNS, receiverPort)
			if logGenNS, err = e2e.DeployCURLLogGenerator(httpReceiverEndpoint); err != nil {
				Fail(fmt.Sprintf("unable to deploy log generator %v.", err))
			}
		})

		It("collector should be a deployment when annotated and http receiver is the only input", func() {
			fluentDeployment, err = e2e.DeployFluentdReceiverWithConf(rootDir, true, fluentConf)
			Expect(err).To(BeNil())
			forwarder.Annotations = deploymentAnnotation
			if err := e2e.CreateClusterLogForwarder(forwarder); err != nil {
				Fail(fmt.Sprintf("Unable to create an instance of logforwarder: %v", err))
			}
			components := []helpers.LogComponentType{helpers.ComponentTypeCollectorDeployment}
			for _, component := range components {
				if err := e2e.WaitFor(component); err != nil {
					Fail(fmt.Sprintf("Failed waiting for component %s to be ready: %v", component, err))
				}
			}

			logStore := e2e.LogStores[fluentDeployment.GetName()]
			has, err := logStore.HasAuditLogs(framework.DefaultWaitForLogsTimeout)
			Expect(err).To(BeNil())
			Expect(has).To(BeTrue())

			collectedLogs, err := logStore.RetrieveLogs()
			Expect(err).To(BeNil())
			auditLogsStr, ok := collectedLogs["audit"]
			fmt.Printf("---- %s\n", auditLogsStr)
			Expect(ok).To(BeTrue())
			auditLogs := map[string]interface{}{}
			err = json.Unmarshal([]byte(auditLogsStr), &auditLogs)
			Expect(err).To(BeNil(), auditLogsStr)
			for k, v := range headers {
				// Headers are capitalized, and sent with prefix "HTTP_"
				headerVal, ok := auditLogs[fmt.Sprintf("HTTP_%s", strings.ToUpper(k))]
				Expect(ok).To(BeTrue())
				Expect(headerVal).To(Equal(v))
			}
			message, ok := auditLogs["log_message"]
			Expect(ok).To(BeTrue())
			Expect(message).To(Not(BeEmpty()))
		})

		AfterEach(func() {
			e2e.Cleanup()
			e2e.WaitForCleanupCompletion(logGenNS, []string{"test"})
		})
	})

})
