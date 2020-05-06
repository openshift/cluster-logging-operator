package fluentlegacy

import (
	"fmt"
	"path/filepath"
	"runtime"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	apps "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/openshift/cluster-logging-operator/pkg/k8shandler"
	"github.com/openshift/cluster-logging-operator/pkg/logger"
	"github.com/openshift/cluster-logging-operator/pkg/utils"
	"github.com/openshift/cluster-logging-operator/test/helpers"
)

// This test verifies we still support latency secure-forward with no ClusterLogForwarder.
var _ = Describe("Backwards compatibility prior to ClusterLogForwarder", func() {
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
			logger.Errorf("unable to deploy log generator. E: %s", err.Error())
		}
		rootDir = filepath.Join(filepath.Dir(filename), "..", "..", "..", "..", "/")
	})
	Describe("when ClusterLogging is configured with no ClusterLogForwarder instance and 'forwarder' to an administrator managed fluentd", func() {
		Context("and the receiver is secured", func() {

			BeforeEach(func() {
				if fluentDeployment, err = e2e.DeployFluentdReceiver(rootDir, true); err != nil {
					Fail(fmt.Sprintf("Unable to deploy fluent receiver: %v", err))
				}
				//sanity check
				initialWaitForLogsTimeout, _ := time.ParseDuration("30s")
				if exist, _ := e2e.LogStore.HasInfraStructureLogs(initialWaitForLogsTimeout); exist {
					Fail("Found logs when we didnt expect them")
				}
				if exist, _ := e2e.LogStore.HasApplicationLogs(initialWaitForLogsTimeout); exist {
					Fail("Found logs when we didnt expect them")
				}

				//create configmap secure-forward/"secure-forward.conf"
				fluentdConfigMap := k8shandler.NewConfigMap(
					"secure-forward",
					fluentDeployment.Namespace,
					map[string]string{
						"secure-forward.conf": string(utils.GetFileContents(filepath.Join(rootDir, "test/files/secure-forward.conf"))),
					},
				)
				if _, err = e2e.KubeClient.Core().ConfigMaps(fluentDeployment.Namespace).Create(fluentdConfigMap); err != nil {
					Fail(fmt.Sprintf("Unable to create legacy fluent.conf configmap: %v", err))
				}
				e2e.AddCleanup(func() error {
					return e2e.KubeClient.Core().ConfigMaps(fluentdConfigMap.ObjectMeta.Namespace).Delete(fluentdConfigMap.ObjectMeta.Name, nil)
				})

				var secret *v1.Secret
				if secret, err = e2e.KubeClient.Core().Secrets(fluentDeployment.Namespace).Get(fluentDeployment.Name, metav1.GetOptions{}); err != nil {
					Fail(fmt.Sprintf("There was an error fetching the fluent-reciever secrets: %v", err))
				}
				secret = k8shandler.NewSecret("secure-forward", fluentDeployment.Namespace, secret.Data)
				if _, err = e2e.KubeClient.Core().Secrets(fluentDeployment.Namespace).Create(secret); err != nil {
					Fail(fmt.Sprintf("Unable to create secure-forward secret: %v", err))
				}
				e2e.AddCleanup(func() error {
					return e2e.KubeClient.Core().Secrets(fluentDeployment.Namespace).Delete(secret.ObjectMeta.Name, nil)
				})

				components := []helpers.LogComponentType{helpers.ComponentTypeCollector, helpers.ComponentTypeStore}
				cr := helpers.NewClusterLogging(components...)
				if err := e2e.CreateClusterLogging(cr); err != nil {
					Fail(fmt.Sprintf("Unable to create an instance of cluster logging: %v", err))
				}
				for _, component := range components {
					if err := e2e.WaitFor(component); err != nil {
						Fail(fmt.Sprintf("Failed waiting for component %s to be ready: %v", component, err))
					}
				}

			})
			It("should send logs to the forward.Output logstore", func() {
				Expect(e2e.LogStore.HasInfraStructureLogs(helpers.DefaultWaitForLogsTimeout)).To(BeTrue(), "Expected to find stored infrastructure logs")
				Expect(e2e.LogStore.HasApplicationLogs(helpers.DefaultWaitForLogsTimeout)).To(BeTrue(), "Expected to find stored application logs")
			})
		})

		AfterEach(func() {
			e2e.Cleanup()
			e2e.WaitForCleanupCompletion([]string{"fluent-receiver", "fluentd", "elasticsearch"})
		})
	})

})
