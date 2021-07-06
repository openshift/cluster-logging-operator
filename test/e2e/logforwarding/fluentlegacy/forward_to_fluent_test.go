package fluentlegacy

import (
	"context"
	"fmt"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	"path/filepath"
	"runtime"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	apps "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/ViaQ/logerr/log"
	"github.com/openshift/cluster-logging-operator/internal/k8shandler"
	"github.com/openshift/cluster-logging-operator/internal/utils"
	"github.com/openshift/cluster-logging-operator/test/helpers"
)

// This test verifies we still support latency secure-forward with no ClusterLogForwarder.
var _ = Describe("[ClusterLogging] Forwards logs", func() {
	_, filename, _, _ := runtime.Caller(0)
	log.Info("Running ", "filename", filename)
	var (
		err              error
		fluentDeployment *apps.Deployment
		e2e              = helpers.NewE2ETestFramework()
		rootDir          string
	)
	BeforeEach(func() {
		if err := e2e.DeployLogGenerator(); err != nil {
			log.Error(err, "unable to deploy log generator.")
		}
		rootDir = filepath.Join(filepath.Dir(filename), "..", "..", "..", "..", "/")
	})

	Describe("when the output in `secure-forward.conf` confimap is a third-party managed fluentd", func() {

		Context("and the receiver is secured", func() {

			BeforeEach(func() {
				if fluentDeployment, err = e2e.DeployFluentdReceiver(rootDir, true); err != nil {
					Fail(fmt.Sprintf("Unable to deploy fluent receiver: %v", err))
				}
				//sanity check
				initialWaitForLogsTimeout, _ := time.ParseDuration("30s")
				name := fluentDeployment.GetName()
				if exist, _ := e2e.LogStores[name].HasInfraStructureLogs(initialWaitForLogsTimeout); exist {
					Fail("Found logs when we didnt expect them")
				}
				if exist, _ := e2e.LogStores[name].HasApplicationLogs(initialWaitForLogsTimeout); exist {
					Fail("Found logs when we didnt expect them")
				}

				//create configmap secure-forward/"secure-forward.conf"
				opts := metav1.CreateOptions{}
				fluentdConfigMap := k8shandler.NewConfigMap(
					"secure-forward",
					fluentDeployment.Namespace,
					map[string]string{
						"secure-forward.conf": string(utils.GetFileContents(filepath.Join(rootDir, "test/files/secure-forward.conf"))),
					},
				)
				if _, err = e2e.KubeClient.CoreV1().ConfigMaps(fluentDeployment.Namespace).Create(context.TODO(), fluentdConfigMap, opts); err != nil {
					Fail(fmt.Sprintf("Unable to create legacy fluent.conf configmap: %v", err))
				}
				e2e.AddCleanup(func() error {
					opts := metav1.DeleteOptions{}
					return e2e.KubeClient.CoreV1().ConfigMaps(fluentdConfigMap.ObjectMeta.Namespace).Delete(context.TODO(), fluentdConfigMap.ObjectMeta.Name, opts)
				})

				var secret *v1.Secret
				if secret, err = e2e.KubeClient.CoreV1().Secrets(fluentDeployment.Namespace).Get(context.TODO(), fluentDeployment.Name, metav1.GetOptions{}); err != nil {
					Fail(fmt.Sprintf("There was an error fetching the fluent-receiver secrets: %v", err))
				}

				sOpts := metav1.CreateOptions{}
				secret = k8shandler.NewSecret("secure-forward", fluentDeployment.Namespace, secret.Data)
				if _, err = e2e.KubeClient.CoreV1().Secrets(fluentDeployment.Namespace).Create(context.TODO(), secret, sOpts); err != nil {
					Fail(fmt.Sprintf("Unable to create secure-forward secret: %v", err))
				}
				e2e.AddCleanup(func() error {
					opts := metav1.DeleteOptions{}
					return e2e.KubeClient.CoreV1().Secrets(fluentDeployment.Namespace).Delete(context.TODO(), secret.ObjectMeta.Name, opts)
				})

				components := []helpers.LogComponentType{helpers.ComponentTypeCollector}
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
				name := fluentDeployment.GetName()
				Expect(e2e.LogStores[name].HasInfraStructureLogs(helpers.DefaultWaitForLogsTimeout)).To(BeTrue(), "Expected to find stored infrastructure logs")
				Expect(e2e.LogStores[name].HasApplicationLogs(helpers.DefaultWaitForLogsTimeout)).To(BeTrue(), "Expected to find stored application logs")
			})
		})

		AfterEach(func() {
			e2e.Cleanup()
			e2e.WaitForCleanupCompletion(helpers.OpenshiftLoggingNS, []string{"fluent-receiver", constants.CollectorName, "elasticsearch"})
		})
	})

})
