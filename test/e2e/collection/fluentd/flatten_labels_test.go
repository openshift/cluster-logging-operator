package fluentd

import (
	"fmt"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	apps "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	logforward "github.com/openshift/cluster-logging-operator/pkg/apis/logging/v1alpha1"
	"github.com/openshift/cluster-logging-operator/pkg/logger"
	"github.com/openshift/cluster-logging-operator/test/helpers"
)

var _ = Describe("Fluentd message filtering", func() {
	_, filename, _, _ := runtime.Caller(0)
	logger.Infof("Running %s", filename)
	var (
		err              error
		fluentDeployment *apps.Deployment
		e2e              = helpers.NewE2ETestFramework()
		rootDir          string
	)

	BeforeEach(func() {
		if err := e2e.DeployLogGenerator(); err != nil {
			Fail(fmt.Sprintf("Timed out waiting for the log generator to deploy: %v", err))
		}
		rootDir = filepath.Join(filepath.Dir(filename), "..", "..", "..", "..", "/")
		if fluentDeployment, err = e2e.DeployFluentdReceiver(rootDir, false); err != nil {
			Fail(fmt.Sprintf("Unable to deploy fluent receiver: %v", err))
		}

		forwarding := &logforward.LogForwarding{
			TypeMeta: metav1.TypeMeta{
				Kind:       logforward.LogForwardingKind,
				APIVersion: logforward.SchemeGroupVersion.String(),
			},
			ObjectMeta: metav1.ObjectMeta{
				Name: "instance",
			},
			Spec: logforward.ForwardingSpec{
				Outputs: []logforward.OutputSpec{
					logforward.OutputSpec{
						Name:     fluentDeployment.ObjectMeta.Name,
						Type:     logforward.OutputTypeForward,
						Endpoint: fmt.Sprintf("%s.%s.svc:24224", fluentDeployment.ObjectMeta.Name, fluentDeployment.Namespace),
					},
				},
				Pipelines: []logforward.PipelineSpec{
					logforward.PipelineSpec{
						Name:       "test-app",
						OutputRefs: []string{fluentDeployment.ObjectMeta.Name},
						SourceType: logforward.LogSourceTypeApp,
					},
				},
			},
		}
		if err := e2e.CreateLogForwarding(forwarding); err != nil {
			Fail(fmt.Sprintf("Unable to create an instance of logforwarding: %v", err))
		}
		cr := helpers.NewClusterLogging(helpers.ComponentTypeCollector)
		if err := e2e.CreateClusterLogging(cr); err != nil {
			Fail(fmt.Sprintf("Unable to create an instance of cluster logging: %v", err))
		}
		if err := e2e.WaitFor(helpers.ComponentTypeCollector); err != nil {
			Fail(fmt.Sprintf("Failed waiting for component %s to be ready: %v", helpers.ComponentTypeCollector, err))
		}

	})

	AfterEach(func() {
		e2e.Cleanup()
	}, helpers.DefaultCleanUpTimeout)

	It("should remove 'kubernetes.labels' and create 'kubernetes.flat_labels' with an array of 'kubernetes.labels'", func() {
		Expect(e2e.LogStore.HasApplicationLogs(helpers.DefaultWaitForLogsTimeout)).To(BeTrue(), "Expected to find stored application logs")

		//verify infra namespaces are not stored to their own index
		response, err := e2e.LogStore.ApplicationLogs(helpers.DefaultWaitForLogsTimeout)
		Expect(err).To(BeNil(), fmt.Sprintf("Error fetching logs: %v", err))
		logs := strings.Split(response, "\n")
		Expect(len(logs)).To(Not(Equal(0)), "There were no documents returned in the logs")

		//verify the new key exists
		reTimeUnit := regexp.MustCompile(".*\\\"flat_labels\\\":\\[(.*=.*)*,?\\].*")
		Expect(reTimeUnit.MatchString(logs[0])).To(BeTrue(), fmt.Sprintf("Expected to find the kubernetes.flat_labels key in '%s'", logs[0]))

		//verify we removed the old key
		reTimeUnit = regexp.MustCompile(".*\\\"labels\\\":{")
		Expect(reTimeUnit.MatchString(logs[0])).To(BeFalse(), fmt.Sprintf("Did not expect to find the kubernetes.labels key in '%s'", logs[0]))

	})

})
