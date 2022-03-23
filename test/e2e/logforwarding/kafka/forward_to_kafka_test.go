package kafka

import (
	"fmt"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	framework "github.com/openshift/cluster-logging-operator/test/framework/e2e"
	"runtime"

	"github.com/ViaQ/logerr/log"
	loggingv1 "github.com/openshift/cluster-logging-operator/pkg/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/test/helpers"
	"github.com/openshift/cluster-logging-operator/test/helpers/kafka"
	apps "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("[ClusterLogForwarder] Forwards logs", func() {
	_, filename, _, _ := runtime.Caller(0)
	log.Info("Running ", "filename", filename)
	var (
		app *apps.StatefulSet
		err error
		e2e = framework.NewE2ETestFramework()
	)
	BeforeEach(func() {
		if err := e2e.DeployLogGenerator(); err != nil {
			log.Error(err, "unable to deploy log generator.")
		}
	})
	Describe("when the output is a third-party managed kafka", func() {

		BeforeEach(func() {
			topics := []string{kafka.AppLogsTopic, kafka.AuditLogsTopic, kafka.InfraLogsTopic}
			app, err = e2e.DeployKafkaReceiver(topics)
			if err != nil {
				Fail(fmt.Sprintf("Unable to deploy kafka receiver: %v", err))
			}

			cr := helpers.NewClusterLogging(helpers.ComponentTypeCollector)
			if err := e2e.CreateClusterLogging(cr); err != nil {
				Fail(fmt.Sprintf("Unable to create an instance of cluster logging: %v", err))
			}
			forwarder := &loggingv1.ClusterLogForwarder{
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
							Name: fmt.Sprintf("%s-app-out", kafka.DeploymentName),
							Type: loggingv1.OutputTypeKafka,
							URL:  fmt.Sprintf("tls://%s", e2e.LogStores[app.Name].ClusterLocalEndpoint()),
							OutputTypeSpec: loggingv1.OutputTypeSpec{
								Kafka: &loggingv1.Kafka{
									Topic: kafka.AppLogsTopic,
								},
							},
							Secret: &loggingv1.OutputSecretSpec{
								Name: kafka.DeploymentName,
							},
						},
					},
					Pipelines: []loggingv1.PipelineSpec{
						{
							Name:       "kafka-app",
							InputRefs:  []string{loggingv1.InputNameApplication},
							OutputRefs: []string{fmt.Sprintf("%s-app-out", kafka.DeploymentName)},
						},
					},
				},
			}
			if err := e2e.CreateClusterLogForwarder(forwarder); err != nil {
				Fail(fmt.Sprintf("Unable to create an instance of ClusterLogForwarder: %v", err))
			}
			components := []helpers.LogComponentType{helpers.ComponentTypeCollector}
			for _, component := range components {
				if err := e2e.WaitFor(component); err != nil {
					Fail(fmt.Sprintf("Failed waiting for component %s to be ready: %v", component, err))
				}
			}

		})

		It("should send logs to the forward.Output logstore", func() {
			Expect(e2e.LogStores[app.Name].HasApplicationLogs(framework.DefaultWaitForLogsTimeout)).To(BeTrue(), "Expected to find stored application logs")
		})

		AfterEach(func() {
			e2e.Cleanup()
		})
	})
})
