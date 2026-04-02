package metrics

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	"github.com/openshift/cluster-logging-operator/test/helpers/oc"
)

const (
	// namespace where the project is deployed in
	namespace = constants.OpenshiftNS

	// serviceAccountName created for the project
	serviceAccountName = constants.ClusterLoggingOperator

	// metricsServiceName is the name of the metrics service of the project
	metricsServiceName = "cluster-logging-operator-metrics"

	// metricsRoleBindingName is the name of the RBAC that will be created to allow get the metrics data
	metricsRoleBindingName = "cluster-logging-operator-metrics-reader"
	metricsRoleName        = "cluster-logging-operator-metrics-reader"
	curlImage              = "registry.access.redhat.com/ubi9/ubi"
)

var _ = Describe("Manager", Ordered, func() {
	var controllerPodName string

	// Before running the tests, set up the environment by creating the namespace,
	// enforce the restricted security policy to the namespace, installing CRDs,
	// and deploying the controller.
	BeforeAll(func() {
		By("labeling the namespace to enforce the restricted security policy")
		_, err := oc.Literal().From("oc label --overwrite ns %s pod-security.kubernetes.io/enforce=restricted", namespace).Run()
		Expect(err).NotTo(HaveOccurred(), "Failed to label namespace with restricted policy")
	})

	// After all tests have been executed, clean up by undeploying the controller, uninstalling CRDs,
	// and deleting the namespace.
	AfterAll(func() {
		if os.Getenv("DO_CLEANUP") == "" {
			By("cleaning up labeling the namespace to revert enforcing the restricted security policy")
			_, err := oc.Literal().From("oc label --overwrite ns %s pod-security.kubernetes.io/enforce=privileged", namespace).Run()
			Expect(err).NotTo(HaveOccurred(), "Failed to label namespace with privileged policy")

			By("cleaning up the curl pod for metrics")
			_, _ = oc.Literal().From("oc delete pod curl-metrics -n %s", namespace).Run()
		}
	})

	// After each test, check for failures and collect logs, events,
	// and pod descriptions for debugging.
	AfterEach(func() {
		specReport := CurrentSpecReport()
		if specReport.Failed() {
			By("Fetching controller manager pod logs")
			controllerLogs, err := oc.Logs().WithNamespace(namespace).WithPod(controllerPodName).Run()
			if err == nil {
				_, _ = fmt.Fprintf(GinkgoWriter, "Controller logs:\n %s", controllerLogs)
			} else {
				_, _ = fmt.Fprintf(GinkgoWriter, "Failed to get Controller logs: %s", err)
			}

			By("Fetching Kubernetes events")
			eventsOutput, err := oc.Literal().From("oc get events -n %s --sort-by=.lastTimestamp", namespace).Run()
			if err == nil {
				_, _ = fmt.Fprintf(GinkgoWriter, "Kubernetes events:\n%s", eventsOutput)
			} else {
				_, _ = fmt.Fprintf(GinkgoWriter, "Failed to get Kubernetes events: %s", err)
			}

			By("Fetching curl-metrics logs")
			metricsOutput, err := oc.Logs().WithNamespace(namespace).WithPod("curl-metrics").Run()
			if err == nil {
				_, _ = fmt.Fprintf(GinkgoWriter, "Metrics logs:\n %s", metricsOutput)
			} else {
				_, _ = fmt.Fprintf(GinkgoWriter, "Failed to get curl-metrics logs: %s", err)
			}

			By("Fetching controller manager pod description")
			podDescription, err := oc.Literal().From("oc describe pod %s -n %s", controllerPodName, namespace).Run()
			if err == nil {
				fmt.Println("Pod description:\n", podDescription)
			} else {
				fmt.Println("Failed to describe controller pod")
			}
		}
	})

	SetDefaultEventuallyTimeout(2 * time.Minute)
	SetDefaultEventuallyPollingInterval(time.Second)

	Context("Manager", func() {
		It("should run successfully", func() {
			By("validating that the controller-manager pod is running as expected")
			verifyControllerUp := func(g Gomega) {
				// Get the name of the controller-manager pod
				podOutput, err := oc.Get().
					WithNamespace(namespace).
					Pod().
					Selector("control-plane=controller-manager").
					OutputName().
					Run()
				g.Expect(err).NotTo(HaveOccurred(), "Failed to retrieve controller-manager pod information")
				podNames := getNonEmptyLines(podOutput)
				g.Expect(podNames).To(HaveLen(1), "expected 1 controller pod running")
				controllerPodName = podNames[0]
				g.Expect(controllerPodName).To(ContainSubstring("cluster-logging-operator"))

				// Validate the pod's status
				output, err := oc.Get().
					WithNamespace(namespace).
					Pod().
					Name(controllerPodName).
					OutputJsonpath("{.status.phase}").
					Run()
				g.Expect(err).NotTo(HaveOccurred())
				g.Expect(output).To(Equal("Running"), "Incorrect controller-manager pod status")
			}
			Eventually(verifyControllerUp).Should(Succeed())
		})

		It("should ensure the metrics endpoint is serving metrics", func() {
			By("creating a ClusterRoleBinding for the service account to allow access to metrics")
			_, err := oc.Literal().From("oc create clusterrolebinding %s --clusterrole=%s --serviceaccount=%s:%s",
				metricsRoleBindingName, metricsRoleName, namespace, serviceAccountName).Run()
			if err != nil && !strings.Contains(err.Error(), "already exists") {
				Expect(err).NotTo(HaveOccurred(), "Failed to create ClusterRoleBinding")
			}
			By("validating that the metrics service is available")
			_, err = oc.Get().WithNamespace(namespace).Resource("service", metricsServiceName).Run()
			Expect(err).NotTo(HaveOccurred(), "Metrics service should exist")

			By("getting the service account token")
			token, err := serviceAccountToken()
			Expect(err).NotTo(HaveOccurred())
			Expect(token).NotTo(BeEmpty())

			By("ensuring the controller pod is ready")
			verifyControllerPodReady := func(g Gomega) {
				output, err := oc.Get().
					WithNamespace(namespace).
					Pod().
					Name(controllerPodName).
					OutputJsonpath("{.status.conditions[?(@.type=='Ready')].status}").
					Run()
				g.Expect(err).NotTo(HaveOccurred())
				g.Expect(output).To(Equal("True"), "Controller pod not ready")
			}
			Eventually(verifyControllerPodReady, 3*time.Minute, time.Second).Should(Succeed())

			By("verifying that the controller manager is serving the metrics server")
			verifyMetricsServerStarted := func(g Gomega) {
				output, err := oc.Logs().WithNamespace(namespace).WithPod(controllerPodName).Run()
				g.Expect(err).NotTo(HaveOccurred())
				g.Expect(output).To(ContainSubstring("Serving metrics server"),
					"Metrics server not yet started")
			}
			Eventually(verifyMetricsServerStarted, 3*time.Minute, time.Second).Should(Succeed())

			By("creating the curl-metrics pod to access the metrics endpoint")
			overrides := fmt.Sprintf(`{
				"spec": {
					"containers": [{
						"name": "curl",
						"image": %q,
						"command": ["/bin/sh", "-c"],
						"args": [
							"for i in $(seq 1 30); do curl -v -H 'Authorization: Bearer %s' --cacert /var/run/secrets/kubernetes.io/serviceaccount/service-ca.crt https://%s.%s.svc.cluster.local:8443/metrics && exit 0 || sleep 2; done; exit 1"
						],
						"volumeMounts": [{
							"name": "service-ca",
							"mountPath": "/var/run/secrets/kubernetes.io/serviceaccount",
							"readOnly": true
						}],
						"securityContext": {
							"readOnlyRootFilesystem": true,
							"allowPrivilegeEscalation": false,
							"capabilities": {
								"drop": ["ALL"]
							},
							"runAsNonRoot": true,
							"runAsUser": 1000,
							"seccompProfile": {
								"type": "RuntimeDefault"
							}
						}
					}],
					"volumes": [{
						"name": "service-ca",
						"configMap": {
							"name": "openshift-service-ca.crt",
							"items": [{"key": "service-ca.crt", "path": "service-ca.crt"}]
						}
					}],
					"serviceAccountName": "%s"
				}
			}`, curlImage, token, metricsServiceName, namespace, serviceAccountName)
			_, err = oc.Run().
				Name("curl-metrics").
				WithNamespace(namespace).
				Image(curlImage).
				Restart("Never").
				Overrides(overrides).
				Run()
			Expect(err).NotTo(HaveOccurred(), "Failed to create curl-metrics pod")

			By("waiting for the curl-metrics pod to complete.")
			verifyCurlUp := func(g Gomega) {
				output, err := oc.Get().
					WithNamespace(namespace).
					Pod().
					Name("curl-metrics").
					OutputJsonpath("{.status.phase}").
					Run()
				g.Expect(err).NotTo(HaveOccurred())
				g.Expect(output).To(Equal("Succeeded"), "curl pod in wrong status")
			}
			Eventually(verifyCurlUp, 5*time.Minute).Should(Succeed())

			By("getting the metrics by checking curl-metrics logs")
			verifyMetricsAvailable := func(g Gomega) {
				metricsOutput, err := getMetricsOutput()
				g.Expect(err).NotTo(HaveOccurred(), "Failed to retrieve logs from curl pod")
				g.Expect(metricsOutput).NotTo(BeEmpty())
				g.Expect(metricsOutput).To(ContainSubstring("< HTTP/1.1 200 OK"))
			}
			Eventually(verifyMetricsAvailable, 2*time.Minute).Should(Succeed())
		})

		It("should return metrics from the endpoint serving metrics", func() {
			metricsOutput, err := getMetricsOutput()
			Expect(err).NotTo(HaveOccurred(), "Failed to retrieve logs from curl pod")
			Expect(metricsOutput).To(ContainSubstring(
				fmt.Sprintf(`controller_runtime_reconcile_total{controller="%s",result="success"}`,
					strings.ToLower("clusterlogforwarder"),
				)))
		})
	})
})

// serviceAccountToken returns a token for the specified service account in the given namespace.
// It uses the Kubernetes TokenRequest API to generate a token by directly sending a request
// and parsing the resulting token from the API response.
func serviceAccountToken() (string, error) {
	const tokenRequestRawString = `{
		"apiVersion": "authentication.k8s.io/v1",
		"kind": "TokenRequest"
	}`

	// Temporary file to store the token request
	secretName := fmt.Sprintf("%s-token-request", serviceAccountName)
	tokenRequestFile := filepath.Join("/tmp", secretName)
	err := os.WriteFile(tokenRequestFile, []byte(tokenRequestRawString), os.FileMode(0o644))
	if err != nil {
		return "", err
	}

	var out string
	verifyTokenCreation := func(g Gomega) {
		// Execute oc command to create the token
		output, err := oc.Literal().From("oc create --raw /api/v1/namespaces/%s/serviceaccounts/%s/token -f %s",
			namespace, serviceAccountName, tokenRequestFile).Run()
		g.Expect(err).NotTo(HaveOccurred())

		// Parse the JSON output to extract the token
		var token tokenRequest
		err = json.Unmarshal([]byte(output), &token)
		g.Expect(err).NotTo(HaveOccurred())

		out = token.Status.Token
	}
	Eventually(verifyTokenCreation).Should(Succeed())

	return out, err
}

// getMetricsOutput retrieves and returns the logs from the curl pod used to access the metrics endpoint.
func getMetricsOutput() (string, error) {
	By("getting the curl-metrics logs")
	return oc.Logs().WithNamespace(namespace).WithPod("curl-metrics").Run()
}

// getNonEmptyLines splits a string by newline and returns only non-empty lines.
func getNonEmptyLines(s string) []string {
	var lines []string
	for _, line := range strings.Split(s, "\n") {
		line = strings.TrimSpace(line)
		if line != "" {
			lines = append(lines, line)
		}
	}
	return lines
}

// tokenRequest is a simplified representation of the Kubernetes TokenRequest API response,
// containing only the token field that we need to extract.
type tokenRequest struct {
	Status struct {
		Token string `json:"token"`
	} `json:"status"`
}
