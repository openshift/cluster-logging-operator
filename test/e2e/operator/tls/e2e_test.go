package tls

import (
	"context"
	"fmt"

	clolog "github.com/ViaQ/logerr/v2/log/static"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	configv1 "github.com/openshift/api/config/v1"
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	internalruntime "github.com/openshift/cluster-logging-operator/internal/runtime"
	obsruntime "github.com/openshift/cluster-logging-operator/internal/runtime/observability"
	internaltls "github.com/openshift/cluster-logging-operator/internal/tls"
	"github.com/openshift/cluster-logging-operator/internal/utils/sets"
	"github.com/openshift/cluster-logging-operator/test/client"
	framework "github.com/openshift/cluster-logging-operator/test/framework/e2e"
	tlsscanner "github.com/openshift/cluster-logging-operator/test/framework/e2e/tls"
	"github.com/openshift/cluster-logging-operator/test/helpers/loki"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	crclient "sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("[E2E][Operator][TLS] TLS Scanner Validation", func() {
	const (
		forwarderName = "tls-test-collector"
	)

	var (
		e2e         *framework.E2ETestFramework
		err         error
		k8sClient   crclient.Client
		profileSpec configv1.TLSProfileSpec
	)

	BeforeEach(func() {
		e2e = framework.NewE2ETestFramework()

		// Create a controller-runtime client with configv1 scheme for fetching APIServer TLS profile
		// This matches the production scheme setup in cmd/main.go
		scheme := runtime.NewScheme()
		Expect(configv1.AddToScheme(scheme)).To(Succeed())
		k8sClient, err = crclient.New(e2e.RestConfig, crclient.Options{Scheme: scheme})
		Expect(err).To(BeNil())

		tlsProfile, err := internaltls.FetchAPIServerTlsProfile(k8sClient)
		Expect(err).To(BeNil(), "Failed to fetch APIServer TLS profile")

		By("Fetching the cluster TLS profile")
		profileSpec = internaltls.GetClusterTLSProfileSpec(tlsProfile)
		clolog.Info("Cluster TLS Profile", "spec", profileSpec)
	})

	AfterEach(func() {
		e2e.Cleanup()
	}, framework.DefaultCleanUpTimeout)

	var (
		runTlsScanner = func(e2e *framework.E2ETestFramework, scanNS string) (results []tlsscanner.ScanResult, err error) {
			By("Deploying TLS Scanner")
			scanner := tlsscanner.NewScanner(e2e.KubeClient, &e2e.CleanupFns)
			e2e.AddCleanup(func() error {
				return e2e.KubeClient.BatchV1().Jobs(scanNS).Delete(context.TODO(), tlsscanner.Name, metav1.DeleteOptions{})
			})
			job, err := scanner.Deploy(scanNS, scanNS)
			Expect(err).To(BeNil(), "Failed to deploy TLS Scanner")
			Expect(job).NotTo(BeNil())

			By("Waiting for TLS Scanner to complete")
			err = scanner.WaitForCompletion(job, tlsscanner.JobTimeout)
			Expect(err).To(BeNil(), "TLS Scanner job did not complete successfully. It may not have matched any components to scan")

			By("Retrieving TLS scan results")
			results, err = scanner.GetResults(job)
			Expect(err).To(BeNil(), "Failed to retrieve TLS scan results")
			Expect(results).NotTo(BeEmpty(), "TLS Scanner returned no results")
			return results, err
		}

		verifyResultsHaveComponents = func(results []tlsscanner.ScanResult, epxComponents ...string) {
			components := sets.NewString()
			for _, result := range results {
				components.Insert(result.Component)
			}
			Expect(components.List()).To(ConsistOf(epxComponents))
		}
	)

	Context("when inspecting deployed ClusterLogForwarder", func() {

		var (
			testNS string
			clf    *obs.ClusterLogForwarder
			l      *loki.Receiver
			sa     *corev1.ServiceAccount
		)

		BeforeEach(func() {

			testNS = e2e.CreateTestNamespace(func(namespace *corev1.Namespace) {
				namespace.Labels = map[string]string{
					"pod-security.kubernetes.io/audit":   "privileged",
					"pod-security.kubernetes.io/enforce": "privileged",
					"pod-security.kubernetes.io/warn":    "privileged",
				}
			})

			// Create service account for the collector with permissions for application and infrastructure logs
			sa, err = e2e.BuildAuthorizationFor(testNS, forwarderName).
				AllowClusterRole(framework.ClusterRoleCollectApplicationLogs).
				AllowClusterRole(framework.ClusterRoleCollectInfrastructureLogs).
				Create()
			Expect(err).To(BeNil())

			// Deploy Loki receiver
			l = loki.NewReceiver(testNS, "loki-server")
			Expect(l.Create(client.Get())).To(Succeed())

			// Deploy ClusterLogForwarder with both default inputs and receiver inputs
			// to ensure all input receiver types are running for TLS scanning
			clf = obsruntime.NewClusterLogForwarder(testNS, forwarderName, internalruntime.Initialize, func(clf *obs.ClusterLogForwarder) {
				clf.Spec.ServiceAccount.Name = sa.Name
				clf.Spec.Inputs = []obs.InputSpec{
					{
						Name: "http-receiver",
						Type: obs.InputTypeReceiver,
						Receiver: &obs.ReceiverSpec{
							Type: obs.ReceiverTypeHTTP,
							Port: 8080,
							HTTP: &obs.HTTPReceiver{
								Format: obs.HTTPReceiverFormatKubeAPIAudit,
							},
						},
					},
					{
						Name: "syslog-receiver",
						Type: obs.InputTypeReceiver,
						Receiver: &obs.ReceiverSpec{
							Type: obs.ReceiverTypeSyslog,
							Port: 10514,
						},
					},
				}
				clf.Spec.Outputs = []obs.OutputSpec{
					{
						Name: "loki-output",
						Type: obs.OutputTypeLoki,
						Loki: &obs.Loki{
							URLSpec: obs.URLSpec{
								URL: l.InternalURL("").String(),
							},
						},
					},
				}
				clf.Spec.Pipelines = []obs.PipelineSpec{
					{
						Name:       "test-app",
						InputRefs:  []string{string(obs.InputTypeApplication)},
						OutputRefs: []string{"loki-output"},
					},
					{
						Name:       "test-receivers",
						InputRefs:  []string{"http-receiver", "syslog-receiver"},
						OutputRefs: []string{"loki-output"},
					},
				}
			})

			if err := e2e.CreateObservabilityClusterLogForwarder(clf); err != nil {
				Fail(fmt.Sprintf("Unable to create ClusterLogForwarder: %v", err))
			}

			if err := e2e.WaitForDaemonSet(clf.Namespace, clf.Name); err != nil {
				Fail(fmt.Sprintf("Failed waiting for collector DaemonSet to be ready: %v", err))
			}
		})

		It("should validate the TLS server configurations match the cluster TLS profile", func() {
			results, _ := runTlsScanner(e2e, testNS)
			clolog.Info("TLS Scanner found endpoints", "count", len(results))
			clolog.V(2).Info("TLS endpoint scanned", "result", results)
			verifyResultsHaveComponents(results, constants.VectorName)

			By("Validating TLS compliance")
			err = tlsscanner.ValidateCompliance(results, profileSpec)
			Expect(err).To(BeNil(), "TLS compliance validation failed")
		})
	})

	Context("when inspecting the operator and LogFileMetricExporter", func() {
		It("should validate the TLS configurations matches the cluster TLS profile", func() {

			// Deploy LFME
			lfme := internalruntime.NewLogFileMetricExporter(constants.OpenshiftNS, constants.SingletonName)
			e2e.AddCleanup(func() error {
				return e2e.KubeClient.AppsV1().DaemonSets(constants.OpenshiftNS).Delete(context.TODO(), lfme.Name, metav1.DeleteOptions{})
			})
			if err := e2e.Create(lfme); err != nil {
				Fail(fmt.Sprintf("Unable to create LogFileMetricExporter: %v", err))
			}
			if err := e2e.WaitForDaemonSet(lfme.Namespace, constants.LogfilesmetricexporterName); err != nil {
				Fail(fmt.Sprintf("Failed waiting for lfme DaemonSet to be ready: %v", err))
			}

			results, _ := runTlsScanner(e2e, constants.OpenshiftNS)
			clolog.Info("TLS Scanner found endpoints", "count", len(results))
			verifyResultsHaveComponents(results, constants.ClusterLoggingOperator, constants.LogfilesmetricexporterName)

			By("Validating TLS compliance")
			err = tlsscanner.ValidateCompliance(results, profileSpec)
			Expect(err).To(BeNil(), "TLS compliance validation failed")
		})
	})
})
