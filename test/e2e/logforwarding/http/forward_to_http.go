package http

import (
	"encoding/json"
	"fmt"
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	"github.com/openshift/cluster-logging-operator/internal/runtime"
	obsruntime "github.com/openshift/cluster-logging-operator/internal/runtime/observability"
	corev1 "k8s.io/api/core/v1"
	"strings"

	framework "github.com/openshift/cluster-logging-operator/test/framework/e2e"
	apps "k8s.io/api/apps/v1"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("[ClusterLogForwarder] Forwards logs", func() {
	var (
		err              error
		e2e              = framework.NewE2ETestFramework()
		forwarder        *obs.ClusterLogForwarder
		forwarderName    = "my-forwarder"
		deployNS         string
		logGenNS         string
		fluentDeployment *apps.Deployment
		headers          = map[string]string{"h1": "v1", "h2": "v2"}
		serviceAccount   *corev1.ServiceAccount
	)
	Describe("with vector collector", func() {
		BeforeEach(func() {
			deployNS = e2e.CreateTestNamespace()

			fluentDeployment, err = e2e.DeployFluentdReceiverWithConf(deployNS, true, framework.FluentConfHTTPWithTLS)
			Expect(err).To(BeNil())
			logStore := e2e.LogStores[fluentDeployment.GetName()]

			if serviceAccount, err = e2e.BuildAuthorizationFor(deployNS, forwarderName).
				AllowClusterRole(framework.ClusterRoleCollectApplicationLogs).
				AllowClusterRole(framework.ClusterRoleCollectInfrastructureLogs).
				AllowClusterRole(framework.ClusterRoleCollectAuditLogs).Create(); err != nil {
				Fail(err.Error())
			}

			forwarder = obsruntime.NewClusterLogForwarder(deployNS, "my-forwarder", runtime.Initialize, func(clf *obs.ClusterLogForwarder) {
				for _, input := range []obs.InputType{obs.InputTypeApplication, obs.InputTypeInfrastructure, obs.InputTypeAudit} {
					clf.Spec.ServiceAccount.Name = serviceAccount.Name
					outputName := fmt.Sprintf("http-%s", input)
					clf.Spec.Outputs = append(clf.Spec.Outputs, obs.OutputSpec{
						Name: outputName,
						Type: obs.OutputTypeHTTP,
						HTTP: &obs.HTTP{
							URLSpec: obs.URLSpec{
								URL: fmt.Sprintf("%s/logs/%s", logStore.ClusterLocalEndpoint(), input),
							},
							Headers: headers,
							Method:  "POST",
						},
						TLS: &obs.OutputTLSSpec{
							InsecureSkipVerify: true,
							TLSSpec: obs.TLSSpec{
								CA: &obs.ConfigMapOrSecretKey{
									Key: constants.TrustedCABundleKey,
									Secret: &corev1.LocalObjectReference{
										Name: framework.FluentdSecretName,
									},
								},
							},
						},
					})
					clf.Spec.Pipelines = append(clf.Spec.Pipelines, obs.PipelineSpec{
						Name:       fmt.Sprintf("%s-logs", input),
						OutputRefs: []string{outputName},
						InputRefs:  []string{string(input)},
					})
				}
			})
			logGenNS := e2e.CreateTestNamespaceWithPrefix("clo-test-loader")
			if err = e2e.DeployLogGeneratorWithNamespaceName(logGenNS, "log-generator", framework.NewDefaultLogGeneratorOptions()); err != nil {
				Fail(fmt.Sprintf("unable to deploy log generator %v.", err))
			}
		})
		It("should send logs to fluentd http", func() {
			if err := e2e.CreateObservabilityClusterLogForwarder(forwarder); err != nil {
				Fail(fmt.Sprintf("Unable to create an instance of logforwarder: %v", err))
			}
			if err := e2e.WaitForDaemonSet(forwarder.Namespace, forwarder.Name); err != nil {
				Fail(err.Error())
			}

			logStore := e2e.LogStores[fluentDeployment.GetName()]
			Expect(logStore.HasApplicationLogs(framework.DefaultWaitForLogsTimeout)).To(BeTrue(), "expected to collect application logs")
			Expect(logStore.HasInfraStructureLogs(framework.DefaultWaitForLogsTimeout)).To(BeTrue(), "expected to collect infrastructure logs")
			Expect(logStore.HasAuditLogs(framework.DefaultWaitForLogsTimeout)).To(BeTrue(), "expected to collect audit logs")

			collectedLogs, err := logStore.RetrieveLogs()
			Expect(err).To(BeNil())
			appLogsStr, ok := collectedLogs["app"]
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
			Expect(message).To(Not(BeEmpty()))
		})
		AfterEach(func() {
			e2e.Cleanup()
			e2e.WaitForCleanupCompletion(logGenNS, []string{"test"})
		})
	})
})
