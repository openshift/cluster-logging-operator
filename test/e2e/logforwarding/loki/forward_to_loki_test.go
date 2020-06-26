package loki

import (
	"fmt"
	"runtime"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	loggingv1 "github.com/openshift/cluster-logging-operator/pkg/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/pkg/logger"
	"github.com/openshift/cluster-logging-operator/test/helpers"
	"github.com/openshift/cluster-logging-operator/test/helpers/loki"

	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("LogForwarding", func() {
	_, filename, _, _ := runtime.Caller(0)
	logger.Infof("Running %s", filename)
	var (
		e2e = helpers.NewE2ETestFramework()
		rcv *appsv1.StatefulSet
		err error
	)
	BeforeEach(func() {
		if err := e2e.DeployLogGenerator(); err != nil {
			logger.Errorf("unable to deploy log generator. E: %s", err.Error())
		}
	})
	Describe("when ClusterLogging is configured with 'forwarding' to an administrator managed Loki", func() {

		BeforeEach(func() {
			rcv, err = e2e.DeployLokiReceiver()
			if err != nil {
				Fail(fmt.Sprintf("Unable to deploy loki receiver: %v", err))
			}

			endpoint := e2e.LogStores[rcv.GetName()].ClusterLocalEndpoint()
			cr := helpers.NewClusterLogging(helpers.ComponentTypeCollector)
			if err := e2e.CreateClusterLogging(cr); err != nil {
				Fail(fmt.Sprintf("Unable to create an instance of cluster logging: %v", err))
			}
			forwarding := &loggingv1.ClusterLogForwarder{
				TypeMeta: metav1.TypeMeta{
					Kind:       loggingv1.ClusterLogForwarderKind,
					APIVersion: loggingv1.SchemeGroupVersion.String(),
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: "instance",
				},
				Spec: loggingv1.ClusterLogForwarderSpec{
					Outputs: []loggingv1.OutputSpec{
						{
							Name: loki.DeploymentName,
							Type: loggingv1.OutputTypeLoki,
							URL:  fmt.Sprintf("http://%s", endpoint),
						},
					},
					Pipelines: []loggingv1.PipelineSpec{
						{
							Name:       "test-app",
							InputRefs:  []string{loggingv1.InputNameApplication},
							OutputRefs: []string{loki.DeploymentName},
						},
						{
							Name:       "test-audit",
							InputRefs:  []string{loggingv1.InputNameAudit},
							OutputRefs: []string{loki.DeploymentName},
						},
						{
							Name:       "test-infra",
							InputRefs:  []string{loggingv1.InputNameInfrastructure},
							OutputRefs: []string{loki.DeploymentName},
						},
					},
				},
			}
			if err := e2e.CreateClusterLogForwarder(forwarding); err != nil {
				Fail(fmt.Sprintf("Unable to create an instance of logforwarding: %v", err))
			}
			components := []helpers.LogComponentType{helpers.ComponentTypeCollector}
			for _, component := range components {
				if err := e2e.WaitFor(component); err != nil {
					Fail(fmt.Sprintf("Failed waiting for component %s to be ready: %v", component, err))
				}
			}

		})

		It("should send logs to the forward.Output logstore", func() {
			name := rcv.GetName()
			Expect(e2e.LogStores[name].HasInfraStructureLogs(helpers.DefaultWaitForLogsTimeout)).To(BeTrue(), "Expected to find stored infrastructure logs")
			Expect(e2e.LogStores[name].HasApplicationLogs(helpers.DefaultWaitForLogsTimeout)).To(BeTrue(), "Expected to find stored application logs")
			Expect(e2e.LogStores[name].HasAuditLogs(helpers.DefaultWaitForLogsTimeout)).To(BeTrue(), "Expected to find stored audit logs")
		})

		AfterEach(func() {
			e2e.Cleanup()
		})
	})
})
