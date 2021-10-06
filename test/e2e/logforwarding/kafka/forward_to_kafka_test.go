package kafka

import (
	"fmt"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	"runtime"
	"time"

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
		e2e = helpers.NewE2ETestFramework()
	)
	BeforeEach(func() {
		if err := e2e.DeployLogGenerator(); err != nil {
			log.Error(err, "unable to deploy log generator.")
		}
	})

	DescribeTable("when sending logs to a third-party managed kafka", func(topic, inputName string, hasLogs func(time.Duration) (bool, error)) {
		topics := []string{topic}
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
						URL:  fmt.Sprintf("tls://%s", e2e.LogStores[app.Name].ClusterLocalEndpoint()),
						OutputTypeSpec: loggingv1.OutputTypeSpec{
							Kafka: &loggingv1.Kafka{
								Topic: topic,
							},
						},
						Secret: &loggingv1.OutputSecretSpec{
							Name: kafka.DeploymentName,
						},
					},
				},
				Pipelines: []loggingv1.PipelineSpec{
					{
						Name:       topic,
						InputRefs:  []string{inputName},
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
		Expect(hasLogs(helpers.DefaultWaitForLogsTimeout)).To(BeTrue())
	},
		Entry("should store application logs", kafka.AppLogsTopic, loggingv1.InputNameApplication, e2e.LogStores[app.Name].HasApplicationLogs),
		Entry("should store infrastructure logs", kafka.InfraLogsTopic, loggingv1.InputNameInfrastructure, e2e.LogStores[app.Name].HasInfraStructureLogs),
		Entry("should store audit logs", kafka.AuditLogsTopic, loggingv1.InputNameAudit, e2e.LogStores[app.Name].HasAuditLogs),
	)

	AfterEach(func() {
		e2e.Cleanup()
	})
})
