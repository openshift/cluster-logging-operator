package admission

import (
	"fmt"
	"os/exec"
	"strings"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	internaladmission "github.com/openshift/cluster-logging-operator/internal/admission"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	framework "github.com/openshift/cluster-logging-operator/test/framework/e2e"
)

var _ = Describe("[collection] Protected collector ServiceAccount admission", Ordered, func() {

	const (
		clfName       = "protected-sa-test"
		collectorSA   = "log-collector"
		unprotectedSA = "plain-sa"
	)

	var (
		e2e      *framework.E2ETestFramework
		deployNS string
		user     string
	)

	BeforeAll(func() {
		e2e = framework.NewE2ETestFramework()
		deployNS = e2e.Test.NS.Name
		user = fmt.Sprintf("system:serviceaccount:%s:restricted-user", deployNS)

		checkVAPInstalled()

		for _, sa := range []string{collectorSA, unprotectedSA, "restricted-user"} {
			_, err := e2e.BuildAuthorizationFor(deployNS, sa).Create()
			Expect(err).ToNot(HaveOccurred(), "create ServiceAccount %s", sa)
		}

		grantWorkloadEditor(deployNS)

		clf := fmt.Sprintf(`
apiVersion: observability.openshift.io/v1
kind: ClusterLogForwarder
metadata:
  name: %s
  namespace: %s
spec:
  serviceAccount:
    name: %s
  outputs:
  - name: test-output
    type: lokiStack
    lokiStack:
      target:
        name: lokistack
        namespace: openshift-logging
      authentication:
        token:
          from: serviceAccount
  pipelines:
  - name: test-pipe
    inputRefs:
    - application
    outputRefs:
    - test-output
`, clfName, deployNS, collectorSA)

		out, err := ocCreate(deployNS, clf)
		Expect(err).ToNot(HaveOccurred(), "create CLF: %s", out)

		waitForSAProtected(deployNS, collectorSA)
	})

	AfterAll(func() {
		if e2e != nil {
			e2e.Cleanup()
		}
	})

	It("denies a bare Pod referencing the protected SA even with copied collector metadata", func() {
		out, err := ocCreateAs(deployNS, user, spoofedPodYAML(deployNS, collectorSA, "evil-pod", clfName))
		Expect(err).To(HaveOccurred(), "expected Pod creation to be denied, got: %s", out)
		Expect(out).To(ContainSubstring("protected ServiceAccount"))
	})

	It("denies a Deployment referencing the protected SA", func() {
		out, err := ocCreateAs(deployNS, user, deploymentYAML(deployNS, collectorSA, "evil-deploy"))
		Expect(err).To(HaveOccurred(), "expected Deployment creation to be denied, got: %s", out)
		Expect(out).To(ContainSubstring("protected ServiceAccount"))
	})

	It("allows a Pod referencing an unprotected SA (no collateral damage)", func() {
		out, err := ocCreateAs(deployNS, user, spoofedPodYAML(deployNS, unprotectedSA, "plain-pod", clfName))
		Expect(err).ToNot(HaveOccurred(), "expected Pod with unprotected SA to be admitted, got: %s", out)
	})
})

func checkVAPInstalled() {
	for _, resource := range []string{
		"validatingadmissionpolicy/" + internaladmission.ProtectedSAPodsPolicyName,
		"validatingadmissionpolicybinding/" + internaladmission.ProtectedSAPodsBindingName,
		"validatingadmissionpolicy/" + internaladmission.ProtectedSAWorkloadsPolicyName,
		"validatingadmissionpolicybinding/" + internaladmission.ProtectedSAWorkloadsBindingName,
	} {
		out, err := exec.Command("oc", "get", resource).CombinedOutput()
		Expect(err).ToNot(HaveOccurred(), "expected %s to exist (operator must reconcile VAP): %s", resource, out)
	}
}

func waitForSAProtected(namespace, sa string) {
	key := fmt.Sprintf("sa_%s_%s", namespace, sa)
	Eventually(func(g Gomega) {
		out, err := exec.Command("oc", "get", "configmap", internaladmission.ProtectedSAConfigMapName,
			"-n", constants.OpenshiftNS, "-o", "json").CombinedOutput()
		g.Expect(err).ToNot(HaveOccurred(), "get param ConfigMap: %s", out)
		g.Expect(string(out)).To(ContainSubstring(key),
			"operator did not mark %s/%s protected", namespace, sa)
	}, 2*time.Minute, 5*time.Second).Should(Succeed())
}

func ocCreate(namespace, yaml string) (string, error) {
	cmd := exec.Command("oc", "create", "-n", namespace, "-f", "-")
	cmd.Stdin = strings.NewReader(yaml)
	out, err := cmd.CombinedOutput()
	return string(out), err
}

func ocCreateAs(namespace, user, yaml string) (string, error) {
	cmd := exec.Command("oc", "create", "-n", namespace, "--as", user, "-f", "-")
	cmd.Stdin = strings.NewReader(yaml)
	out, err := cmd.CombinedOutput()
	return string(out), err
}

func grantWorkloadEditor(namespace string) {
	out, err := exec.Command("oc", "create", "role", "workload-editor",
		"--verb=create,get,list,delete",
		"--resource=pods,deployments.apps",
		"-n", namespace).CombinedOutput()
	if err != nil {
		Fail(fmt.Sprintf("create workload-editor role: %s %v", out, err))
	}

	out, err = exec.Command("oc", "create", "rolebinding", "restricted-user-workload-editor",
		"--role=workload-editor",
		"--serviceaccount="+namespace+":restricted-user",
		"-n", namespace).CombinedOutput()
	if err != nil {
		Fail(fmt.Sprintf("create workload-editor rolebinding: %s %v", out, err))
	}
}

func spoofedPodYAML(namespace, sa, name, instanceName string) string {
	return fmt.Sprintf(`
apiVersion: v1
kind: Pod
metadata:
  name: %s
  namespace: %s
  labels:
    app.kubernetes.io/name: vector
    app.kubernetes.io/instance: %s
    app.kubernetes.io/component: collector
    app.kubernetes.io/part-of: cluster-logging
    app.kubernetes.io/managed-by: cluster-logging-operator
    vector.dev/exclude: "true"
  annotations:
    target.workload.openshift.io/management: '{"effect": "PreferredDuringScheduling"}'
spec:
  serviceAccountName: %s
  containers:
  - name: c
    image: registry.redhat.io/ubi9/ubi-minimal:latest
    command: ["sleep", "3600"]
`, name, namespace, instanceName, sa)
}

func deploymentYAML(namespace, sa, name string) string {
	return fmt.Sprintf(`
apiVersion: apps/v1
kind: Deployment
metadata:
  name: %s
  namespace: %s
spec:
  replicas: 1
  selector:
    matchLabels:
      app: %s
  template:
    metadata:
      labels:
        app: %s
        app.kubernetes.io/name: vector
        app.kubernetes.io/component: collector
    spec:
      serviceAccountName: %s
      containers:
      - name: c
        image: registry.redhat.io/ubi9/ubi-minimal:latest
        command: ["sleep", "3600"]
`, name, namespace, name, name, sa)
}
