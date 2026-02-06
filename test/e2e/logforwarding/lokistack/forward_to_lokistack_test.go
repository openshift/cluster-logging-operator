package lokistack

import (
	"encoding/json"
	"fmt"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	"github.com/openshift/cluster-logging-operator/internal/runtime"
	obsruntime "github.com/openshift/cluster-logging-operator/internal/runtime/observability"
	framework "github.com/openshift/cluster-logging-operator/test/framework/e2e"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

var _ = Describe("[ClusterLogForwarder] Forward to Lokistack", func() {
	const (
		forwarderName = "my-forwarder"
		logGenName    = "log-generator"
		outputName    = "lokistack-output"
	)
	var (
		err               error
		e2e               = framework.NewE2ETestFramework()
		forwarder         *obs.ClusterLogForwarder
		deployNS          string
		logGenNS          string
		serviceAccount    *corev1.ServiceAccount
		lokiStackOut      *obs.OutputSpec
		lokistackReceiver *framework.LokistackLogStore
	)

	BeforeEach(func() {
		deployNS = e2e.CreateTestNamespace()

		if err = e2e.DeployMinio(); err != nil {
			Fail(err.Error())
		}
		if err = e2e.DeployLokiOperator(); err != nil {
			Fail(err.Error())
		}
		if lokistackReceiver, err = e2e.DeployLokistackInNamespace(deployNS); err != nil {
			Fail(err.Error())
		}

		if serviceAccount, err = e2e.BuildAuthorizationFor(deployNS, forwarderName).
			AllowClusterRole(framework.ClusterRoleCollectApplicationLogs).
			AllowClusterRole(framework.ClusterRoleCollectInfrastructureLogs).
			AllowClusterRole(framework.ClusterRoleCollectAuditLogs).
			AllowClusterRole(framework.ClusterRoleAllLogsWriter).
			AllowClusterRole(framework.ClusterRoleAllLogsReader).Create(); err != nil {
			Fail(err.Error())
		}

		forwarder = obsruntime.NewClusterLogForwarder(deployNS, forwarderName, runtime.Initialize, func(clf *obs.ClusterLogForwarder) {
			clf.Spec.ServiceAccount.Name = serviceAccount.Name
			// Testing removal of otlp annotation validation LOG-8578
			clf.Annotations = map[string]string{}
			clf.Spec.Pipelines = append(clf.Spec.Pipelines, obs.PipelineSpec{
				Name:       "all-logs-pipeline",
				OutputRefs: []string{outputName},
				InputRefs:  []string{string(obs.InputTypeApplication), string(obs.InputTypeAudit), string(obs.InputTypeInfrastructure)},
			})
			clf.Spec.Collector = &obs.CollectorSpec{
				Resources: &corev1.ResourceRequirements{
					Requests: corev1.ResourceList{
						corev1.ResourceMemory: resource.MustParse("500m"),
					},
				},
			}
		})

		lokiStackOut = &obs.OutputSpec{
			Name: outputName,
			Type: obs.OutputTypeLokiStack,
			LokiStack: &obs.LokiStack{
				Target: obs.LokiStackTarget{
					Namespace: deployNS,
					Name:      "lokistack-dev",
				},
				Authentication: &obs.LokiStackAuthentication{
					Token: &obs.BearerToken{
						From: obs.BearerTokenFromServiceAccount,
					},
				},
			},
			TLS: &obs.OutputTLSSpec{
				TLSSpec: obs.TLSSpec{
					CA: &obs.ValueReference{
						Key:           "service-ca.crt",
						ConfigMapName: "openshift-service-ca.crt",
					},
				},
			},
		}

		// Deploy log generator
		logGenNS = e2e.CreateTestNamespaceWithPrefix("clo-test-loader")
		generatorOpt := framework.NewDefaultLogGeneratorOptions()
		generatorOpt.Count = -1
		if err = e2e.DeployLogGeneratorWithNamespaceName(logGenNS, logGenName, generatorOpt); err != nil {
			Fail(fmt.Sprintf("unable to deploy log generator %v.", err))
		}
	})

	It("should send logs to lokistack's OTLP endpoint when dataModel == Otel", func() {
		lokiStackOut.LokiStack.DataModel = obs.LokiStackDataModelOpenTelemetry
		forwarder.Spec.Outputs = append(forwarder.Spec.Outputs, *lokiStackOut)

		if err := e2e.CreateObservabilityClusterLogForwarder(forwarder); err != nil {
			Fail(fmt.Sprintf("Unable to create an instance of logforwarder: %v", err))
		}
		if err := e2e.WaitForDaemonSet(forwarder.Namespace, forwarder.Name); err != nil {
			Fail(err.Error())
		}

		found, err := lokistackReceiver.HasApplicationLogs(serviceAccount.Name, framework.DefaultWaitForLogsTimeout)
		Expect(err).To(BeNil())
		Expect(found).To(BeTrue())

		otlpKey := "k8s_container_name"
		res, err := lokistackReceiver.GetApplicationLogsByKeyValue(serviceAccount.Name, otlpKey, logGenName, framework.DefaultWaitForLogsTimeout)
		Expect(err).To(BeNil())
		Expect(res).ToNot(BeEmpty())
		Expect(len(res)).To(Equal(1))

		found, err = lokistackReceiver.HasAuditLogs(serviceAccount.Name, framework.DefaultWaitForLogsTimeout)
		Expect(err).To(BeNil())
		Expect(found).To(BeTrue())

		found, err = lokistackReceiver.HasInfrastructureLogs(serviceAccount.Name, framework.DefaultWaitForLogsTimeout)
		Expect(err).To(BeNil())
		Expect(found).To(BeTrue())
	})

	It("should send logs to lokistack when dataModel is not spec'd", func() {
		forwarder.Spec.Outputs = append(forwarder.Spec.Outputs, *lokiStackOut)

		if err := e2e.CreateObservabilityClusterLogForwarder(forwarder); err != nil {
			Fail(fmt.Sprintf("Unable to create an instance of logforwarder: %v", err))
		}
		if err := e2e.WaitForDaemonSet(forwarder.Namespace, forwarder.Name); err != nil {
			Fail(err.Error())
		}

		found, err := lokistackReceiver.HasApplicationLogs(serviceAccount.Name, framework.DefaultWaitForLogsTimeout)
		Expect(err).To(BeNil())
		Expect(found).To(BeTrue())

		expression := `| json kubernetes_container_iostream="stdout"`
		res, err := lokistackReceiver.GetApplicationLogsWithPipeline(serviceAccount.Name, expression, framework.DefaultWaitForLogsTimeout)
		Expect(err).To(BeNil())
		Expect(res).ToNot(BeEmpty())
		Expect(len(res)).To(Equal(1))

		found, err = lokistackReceiver.HasAuditLogs(serviceAccount.Name, framework.DefaultWaitForLogsTimeout)
		Expect(err).To(BeNil())
		Expect(found).To(BeTrue())

		found, err = lokistackReceiver.HasInfrastructureLogs(serviceAccount.Name, framework.DefaultWaitForLogsTimeout)
		Expect(err).To(BeNil())
		Expect(found).To(BeTrue())
	})

	It("should send logs to lokistack with HTTP receiver as audit logs", func() {
		const (
			httpReceiverPort = 8080
			httpReceiver     = "http-audit"
		)

		forwarder.Spec.Inputs = []obs.InputSpec{
			{
				Name: httpReceiver,
				Type: obs.InputTypeReceiver,
				Receiver: &obs.ReceiverSpec{
					Type: obs.ReceiverTypeHTTP,
					Port: httpReceiverPort,
					HTTP: &obs.HTTPReceiver{
						Format: obs.HTTPReceiverFormatKubeAPIAudit,
					},
				},
			},
		}

		forwarder.Spec.Pipelines = []obs.PipelineSpec{
			{
				Name:       "input-receiver-logs",
				OutputRefs: []string{outputName},
				InputRefs:  []string{httpReceiver},
			},
		}

		forwarder.Spec.Outputs = append(forwarder.Spec.Outputs, *lokiStackOut)

		if err := e2e.CreateObservabilityClusterLogForwarder(forwarder); err != nil {
			Fail(fmt.Sprintf("Unable to create an instance of logforwarder: %v", err))
		}
		if err := e2e.WaitForDaemonSet(forwarder.Namespace, forwarder.Name); err != nil {
			Fail(err.Error())
		}

		httpReceiverServiceName := fmt.Sprintf("%s-%s", forwarderName, httpReceiver)
		httpReceiverEndpoint := fmt.Sprintf("https://%s.%s.svc.cluster.local:%d", httpReceiverServiceName, deployNS, httpReceiverPort)

		if err = e2e.DeployCURLLogGeneratorWithNamespaceAndEndpoint(deployNS, httpReceiverEndpoint); err != nil {
			Fail(fmt.Sprintf("unable to deploy log generator %v.", err))
		}

		found, err := lokistackReceiver.HasAuditLogs(serviceAccount.Name, framework.DefaultWaitForLogsTimeout)
		Expect(err).To(BeNil())
		Expect(found).To(BeTrue())
	})

	It("should send logs to lokistack with Syslog receiver as infrastructure logs", func() {
		const (
			syslogReceiver     = "syslog-infra"
			syslogReceiverPort = 8443
			syslogLogGenerator = "syslog-log-generator"
		)

		forwarder.Spec.Inputs = []obs.InputSpec{
			{
				Name: syslogReceiver,
				Type: obs.InputTypeReceiver,
				Receiver: &obs.ReceiverSpec{
					Port: syslogReceiverPort,
					Type: obs.ReceiverTypeSyslog,
				},
			},
		}

		forwarder.Spec.Pipelines = []obs.PipelineSpec{
			{
				Name:       "input-receiver-logs",
				OutputRefs: []string{outputName},
				InputRefs:  []string{syslogReceiver},
			},
		}

		forwarder.Spec.Outputs = append(forwarder.Spec.Outputs, *lokiStackOut)

		if err := e2e.CreateObservabilityClusterLogForwarder(forwarder); err != nil {
			Fail(fmt.Sprintf("Unable to create an instance of logforwarder: %v", err))
		}
		if err := e2e.WaitForDaemonSet(forwarder.Namespace, forwarder.Name); err != nil {
			Fail(err.Error())
		}

		if err = e2e.DeploySocat(forwarder.Namespace, syslogLogGenerator, forwarderName, syslogReceiver, framework.NewDefaultLogGeneratorOptions()); err != nil {
			Fail(fmt.Sprintf("unable to deploy log generator %v.", err))
		}

		requiredApps := write2syslog(e2e, forwarder, syslogLogGenerator, syslogReceiverPort)
		requiredAppsChecklist := map[string]bool{}
		for _, app := range requiredApps {
			requiredAppsChecklist[app] = false
		}

		type LogLineData struct {
			AppName string `json:"appname"`
		}

		Eventually(func(g Gomega) {
			res, err := lokistackReceiver.InfrastructureLogs(serviceAccount.Name, 0, len(requiredApps))
			g.Expect(err).To(BeNil())
			for _, stream := range res {
				for _, valPair := range stream.Values {
					logLine := valPair[1]
					var data LogLineData
					err := json.Unmarshal([]byte(logLine), &data)
					if err != nil {
						GinkgoWriter.Printf("Failed to parse log line: %v\n", err)
						continue
					}
					appName := data.AppName
					if _, isRequired := requiredAppsChecklist[appName]; isRequired {
						requiredAppsChecklist[appName] = true
					}
				}
			}

			for appName, found := range requiredAppsChecklist {
				g.Expect(found).To(BeTrue(), "Failed to find required app '%s' in log streams", appName)
			}

		}).WithTimeout(framework.DefaultWaitForLogsTimeout).WithPolling(5 * time.Second).Should(Succeed())
	})

	It("should send logs to lokistack with otel equivalent default labels when data model is viaq", func() {
		forwarder.Spec.Outputs = append(forwarder.Spec.Outputs, *lokiStackOut)

		if err := e2e.CreateObservabilityClusterLogForwarder(forwarder); err != nil {
			Fail(fmt.Sprintf("Unable to create an instance of logforwarder: %v", err))
		}
		if err := e2e.WaitForDaemonSet(forwarder.Namespace, forwarder.Name); err != nil {
			Fail(err.Error())
		}

		found, err := lokistackReceiver.HasApplicationLogs(serviceAccount.Name, framework.DefaultWaitForLogsTimeout)
		Expect(err).To(BeNil())
		Expect(found).To(BeTrue())

		key := `k8s_pod_name`
		res, err := lokistackReceiver.GetApplicationLogsByKeyValue(serviceAccount.Name, key, logGenName, framework.DefaultWaitForLogsTimeout)
		Expect(err).To(BeNil())
		Expect(res).ToNot(BeEmpty())
		Expect(len(res)).To(Equal(1))

		// Check stream values here - len and contents
		stream := res[0].Stream
		Expect(len(stream)).To(Equal(10))
		wantStreamLabels := []string{
			"k8s_container_name",
			"k8s_namespace_name",
			"k8s_pod_name",
			"k8s_node_name",
			"kubernetes_container_name",
			"kubernetes_namespace_name",
			"kubernetes_pod_name",
			"kubernetes_host",
			"log_type",
			"openshift_log_type"}

		for _, key := range wantStreamLabels {
			_, ok := stream[key]
			Expect(ok).To(BeTrue())
		}
	})

	It("should send logs to lokistack when network policy is restricted", func() {
		forwarder.Spec.Outputs = append(forwarder.Spec.Outputs, *lokiStackOut)
		forwarder.Spec.Collector.NetworkPolicy = &obs.NetworkPolicy{
			RuleSet: obs.NetworkPolicyRuleSetTypeRestrictIngressEgress,
		}

		if err := e2e.CreateObservabilityClusterLogForwarder(forwarder); err != nil {
			Fail(fmt.Sprintf("Unable to create an instance of logforwarder: %v", err))
		}
		if err := e2e.WaitForDaemonSet(forwarder.Namespace, forwarder.Name); err != nil {
			Fail(err.Error())
		}

		found, err := lokistackReceiver.HasApplicationLogs(serviceAccount.Name, framework.DefaultWaitForLogsTimeout)
		Expect(err).To(BeNil())
		Expect(found).To(BeTrue())

		expression := `| json kubernetes_container_iostream="stdout"`
		res, err := lokistackReceiver.GetApplicationLogsWithPipeline(serviceAccount.Name, expression, framework.DefaultWaitForLogsTimeout)
		Expect(err).To(BeNil())
		Expect(res).ToNot(BeEmpty())
		Expect(len(res)).To(Equal(1))

		found, err = lokistackReceiver.HasAuditLogs(serviceAccount.Name, framework.DefaultWaitForLogsTimeout)
		Expect(err).To(BeNil())
		Expect(found).To(BeTrue())

		found, err = lokistackReceiver.HasInfrastructureLogs(serviceAccount.Name, framework.DefaultWaitForLogsTimeout)
		Expect(err).To(BeNil())
		Expect(found).To(BeTrue())
	})

	AfterEach(func() {
		e2e.Cleanup()
		e2e.WaitForCleanupCompletion(logGenNS, []string{"test"})
	})

})

func write2syslog(e2e *framework.E2ETestFramework, fwd *obs.ClusterLogForwarder, logGenPodName string, port int32) []string {
	const (
		host    = "acme.com"
		pid     = 6868
		msg     = "Choose Your Destiny"
		msgId   = "ID7"
		caFile  = "/etc/collector/syslog/tls.crt"
		keyFile = "/etc/collector/syslog/tls.key"
	)
	destinationHost := fmt.Sprintf("%s-syslog-infra.%s.svc.cluster.local", fwd.Name, fwd.Namespace)
	socatCmd := fmt.Sprintf("socat openssl-connect:%s:%d,verify=0,cafile=%s,cert=%s,key=%s -",
		destinationHost, port, caFile, caFile, keyFile)

	now := time.Now()
	utcTime := now.UTC()
	rfc5425Date := utcTime.Format(time.RFC3339)
	rfc3164Date := utcTime.Format(time.Stamp)

	rfc3164AppName := "app_rfc3164"
	rfc5425AppName := "app_rfc5425"

	// RFC5424 format: <pri>ver timestamp hostname app-name procid msgid SD msg
	rfc5425 := fmt.Sprintf("<39>1 %s %s %s %d %s - %s", rfc5425Date, host, rfc5425AppName, pid, msgId, msg)
	// RFC3164 format: <pri>timestamp hostname app-name[procid]: msg
	rfc3164 := fmt.Sprintf("<30>%s %s %s[%d]: %s", rfc3164Date, host, rfc3164AppName, pid, msg)

	cmd := fmt.Sprintf("echo %q | %s; echo %q | %s", rfc3164, socatCmd, rfc5425, socatCmd)

	_, err := e2e.PodExec(fwd.Namespace, logGenPodName, logGenPodName, []string{"/bin/sh", "-c", cmd})
	if err != nil {
		Fail(fmt.Sprintf("Error execution write command: %v", err))
	}
	return []string{rfc5425AppName, rfc3164AppName}
}
