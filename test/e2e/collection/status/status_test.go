package status

import (
	"fmt"
	log "github.com/ViaQ/logerr/v2/log/static"
	loggingv1 "github.com/openshift/cluster-logging-operator/api/logging/v1"
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	"github.com/openshift/cluster-logging-operator/internal/runtime"
	obsruntime "github.com/openshift/cluster-logging-operator/internal/runtime/observability"
	corev1 "k8s.io/api/core/v1"
	"time"

	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/util/retry"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/openshift/cluster-logging-operator/test/framework/e2e"
)

// The goal of the test is to make sure that the CLF resource remains stable for a sufficiently long duration - which
// we establish here as 15 seconds which in testing was enough to detect the issue.
// Ref: https://github.com/openshift/cluster-logging-operator/issues/2315
var _ = Describe("ClusterLogForwarderReconciliation", func() {

	const miscellaneousReceiverName = "miscellaneous-receiver"

	var (
		framework      *e2e.E2ETestFramework
		forwarder      *obs.ClusterLogForwarder
		err            error
		serviceAccount *corev1.ServiceAccount
	)

	BeforeEach(func() {
		framework = e2e.NewE2ETestFramework()

		if serviceAccount, err = framework.BuildAuthorizationFor(constants.OpenshiftNS, constants.SingletonName).
			AllowClusterRole(e2e.ClusterRoleCollectApplicationLogs).
			AllowClusterRole(e2e.ClusterRoleCollectInfrastructureLogs).
			AllowClusterRole(e2e.ClusterRoleCollectAuditLogs).Create(); err != nil {
			Fail(err.Error())
		}

		forwarder = obsruntime.NewClusterLogForwarder(constants.OpenshiftNS, constants.SingletonName, runtime.Initialize, func(clf *obs.ClusterLogForwarder) {
			clf.Spec = obs.ClusterLogForwarderSpec{
				ServiceAccount: obs.ServiceAccount{
					Name: serviceAccount.Name,
				},
				Outputs: []obs.OutputSpec{
					{
						Name: miscellaneousReceiverName,
						Type: loggingv1.OutputTypeLoki,
						Loki: &obs.Loki{
							URLSpec: obs.URLSpec{
								URL: "http://127.0.0.1:3100",
							},
						},
					},
				},
				Pipelines: []obs.PipelineSpec{
					{
						Name:       "test-app",
						InputRefs:  []string{string(obs.InputTypeApplication)},
						OutputRefs: []string{miscellaneousReceiverName},
					},
				},
			}
		})
	})

	AfterEach(func() {
		framework.Cleanup()
	})

	It("should not constantly replace the status during reconciliation", func() {

		// We now expect to see no validation error.
		if err := framework.CreateObservabilityClusterLogForwarder(forwarder); err != nil {
			Fail(fmt.Sprintf("Unable to create an instance of logforwarder: %v", err))
		}
		if err := framework.WaitForDaemonSet(forwarder.Namespace, forwarder.Name); err != nil {
			Fail(err.Error())
		}

		// Now, make sure that the CLF resource version remains unchanged. We run the test 5 times with 1 second between
		// the tests. The test itself lasts 15 seconds.
		retryErr := retry.OnError(
			wait.Backoff{Steps: 5, Duration: 1 * time.Second, Factor: 1.0},
			func(error) bool { return true },
			func() error {
				var resourceVersion string
				log.V(3).Info("Retrieving CLF status for the first time")
				if err := framework.Test.Client.Get(forwarder); err != nil {
					return err
				}
				resourceVersion = forwarder.ResourceVersion
				log.V(3).Info("Sleeping for some time")
				time.Sleep(15 * time.Second)
				log.V(3).Info("Retrieving CLF status for the second time")
				if err := framework.Test.Client.Get(forwarder); err != nil {
					return err
				}
				if resourceVersion != forwarder.ResourceVersion {
					log.V(3).Info("ResourceVersions do not match, CLF was updated. Retrying ...")
					return fmt.Errorf("ResourceVersion not stable, it changed from %q to %q",
						resourceVersion, forwarder.ResourceVersion)
				}
				return nil
			},
		)
		Expect(retryErr).To(BeNil())
	})

})
