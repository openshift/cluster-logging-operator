package kafka

import (
	"fmt"
	"runtime"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/ViaQ/logerr/log"
	loggingv1 "github.com/openshift/cluster-logging-operator/pkg/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/test/helpers"
	"github.com/openshift/cluster-logging-operator/test/helpers/kafka"
	apps "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
)

var _ = Describe("[ClusterLogForwarder] Forwards logs", func() {
	_, filename, _, _ := runtime.Caller(0)
	log.Info("Running ", "filename", filename)
	var (
		app *apps.StatefulSet
		err error
		e2e = helpers.NewE2ETestFramework()
	)
	BeforeEach(func() {
		if err := e2e.DeployLogGenerator(); err != nil {
			log.Error(err, "unable to deploy log generator.")
		}
	})
	Describe("when the output is a third-party managed kafka", func() {

		Context("write app, audit and infra logs on a single topic", func() {
			BeforeEach(func() {
				topics := []string{kafka.DefaultTopic}
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
								Name: kafka.DeploymentName,
								Type: loggingv1.OutputTypeKafka,
								URL: fmt.Sprintf(
									"tls://%s/%s",
									e2e.LogStores[app.Name].ClusterLocalEndpoint(),
									kafka.DefaultTopic,
								),
								Secret: &loggingv1.OutputSecretSpec{
									Name: kafka.DeploymentName,
								},
							},
						},
						Pipelines: []loggingv1.PipelineSpec{
							{
								Name:       "kafka-app",
								InputRefs:  []string{loggingv1.InputNameApplication},
								OutputRefs: []string{kafka.DeploymentName},
							},
							{
								Name:       "kafka-audit",
								InputRefs:  []string{loggingv1.InputNameAudit},
								OutputRefs: []string{kafka.DeploymentName},
							},
							{
								Name:       "kafka-infra",
								InputRefs:  []string{loggingv1.InputNameInfrastructure},
								OutputRefs: []string{kafka.DeploymentName},
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
				hasAppLogs := false
				hasInfraLogs := false
				hasAuditLogs := false
				_ = wait.PollImmediate(time.Second*2, time.Minute*5, func() (bool, error) {
					hasAppLogs, err = e2e.LogStores[app.Name].HasApplicationLogs(helpers.DefaultWaitForLogsTimeout)
					if err != nil {
						return false, nil
					}
					hasInfraLogs, err = e2e.LogStores[app.Name].HasInfraStructureLogs(helpers.DefaultWaitForLogsTimeout)
					if err != nil {
						return false, nil
					}
					hasAuditLogs, err = e2e.LogStores[app.Name].HasAuditLogs(helpers.DefaultWaitForLogsTimeout)
					if err != nil {
						return false, nil
					}
					return hasAppLogs && hasInfraLogs && hasAuditLogs, nil
				})
				Expect(hasAppLogs).To(BeTrue(), "Expected to find stored application logs")
				Expect(hasInfraLogs).To(BeTrue(), "Expected to find stored infrastructure logs")
				Expect(hasAuditLogs).To(BeTrue(), "Expected to find stored audit logs")
			})

			AfterEach(func() {
				e2e.Cleanup()
			})
		})

		Context("split app, audit and infra on different topics", func() {
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
							{
								Name: fmt.Sprintf("%s-audit-out", kafka.DeploymentName),
								Type: loggingv1.OutputTypeKafka,
								URL:  fmt.Sprintf("tls://%s", e2e.LogStores[app.Name].ClusterLocalEndpoint()),
								OutputTypeSpec: loggingv1.OutputTypeSpec{
									Kafka: &loggingv1.Kafka{
										Topic: kafka.AuditLogsTopic,
									},
								},
								Secret: &loggingv1.OutputSecretSpec{
									Name: kafka.DeploymentName,
								},
							},
							{
								Name: fmt.Sprintf("%s-infra-out", kafka.DeploymentName),
								Type: loggingv1.OutputTypeKafka,
								URL:  fmt.Sprintf("tls://%s", e2e.LogStores[app.Name].ClusterLocalEndpoint()),
								OutputTypeSpec: loggingv1.OutputTypeSpec{
									Kafka: &loggingv1.Kafka{
										Topic: kafka.InfraLogsTopic,
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
							{
								Name:       "kafka-audit",
								InputRefs:  []string{loggingv1.InputNameAudit},
								OutputRefs: []string{fmt.Sprintf("%s-audit-out", kafka.DeploymentName)},
							},
							{
								Name:       "kafka-infra",
								InputRefs:  []string{loggingv1.InputNameInfrastructure},
								OutputRefs: []string{fmt.Sprintf("%s-infra-out", kafka.DeploymentName)},
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
				Expect(e2e.LogStores[app.Name].HasInfraStructureLogs(helpers.DefaultWaitForLogsTimeout)).To(BeTrue(), "Expected to find stored infrastructure logs")
				Expect(e2e.LogStores[app.Name].HasApplicationLogs(helpers.DefaultWaitForLogsTimeout)).To(BeTrue(), "Expected to find stored application logs")
				Expect(e2e.LogStores[app.Name].HasAuditLogs(helpers.DefaultWaitForLogsTimeout)).To(BeTrue(), "Expected to find stored audit logs")
			})

			AfterEach(func() {
				e2e.Cleanup()
			})
		})
	})
})
