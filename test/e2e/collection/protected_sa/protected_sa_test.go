package protected_sa

import (
	"fmt"
	"os/exec"
	"strings"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/openshift/cluster-logging-operator/internal/admission"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	framework "github.com/openshift/cluster-logging-operator/test/framework/e2e"
)

// This test validates the protected collector ServiceAccount admission policies.
//
// A user who can write a ClusterLogForwarder can name any existing SA; the
// operator marks that SA "protected" and two ValidatingAdmissionPolicies then
// ensure the SA can only be used by CLO-managed collector workloads. The trust
// anchor is request.userInfo (the authenticated creator), NOT Pod metadata, so
// copying the collector's labels/annotations/name does not help an attacker.
//
// It mirrors hack/test-protected-sa.sh: admission-only checks at CREATE. The CLF
// may be Not Ready (no collect roles / no LokiStack) — that is expected.
const clfName = "protected-sa-test"

var _ = Describe("[collection] Protected collector ServiceAccount admission", Ordered, func() {

	const (
		collectorSA    = "log-collector"
		unprotectedSA  = "plain-sa"
		restrictedUser = "system:serviceaccount:%s:restricted-user"

		clfFmt = `
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
`
	)

	var (
		e2e      *framework.E2ETestFramework
		deployNS string
		user     string
	)

	BeforeAll(func() {
		e2e = framework.NewE2ETestFramework()
		deployNS = e2e.Test.NS.Name
		user = fmt.Sprintf(restrictedUser, deployNS)

		checkVAPInstalled()

		// The three ServiceAccounts: the collector SA that becomes protected, an
		// unprotected SA for the pass-through case, and the restricted user.
		for _, sa := range []string{collectorSA, unprotectedSA, "restricted-user"} {
			_, err := e2e.BuildAuthorizationFor(deployNS, sa).Create()
			Expect(err).ToNot(HaveOccurred(), "create ServiceAccount %s", sa)
		}

		// Allow the restricted user to create Pods and Deployments. Without the
		// VAPs this user could then run any workload as the collector SA.
		grantWorkloadEditor(deployNS)

		// Create the CLF as cluster-admin so the operator marks collectorSA
		// protected.
		out, err := ocCreate(deployNS, fmt.Sprintf(clfFmt, clfName, deployNS, collectorSA))
		Expect(err).ToNot(HaveOccurred(), "create CLF: %s", out)

		waitForSAProtected(deployNS, collectorSA)
	})

	AfterAll(func() {
		if e2e != nil {
			e2e.Cleanup()
		}
	})

	It("denies a bare Pod referencing the protected SA even with copied collector metadata", func() {
		out, err := ocCreateAs(deployNS, user, spoofedPodYAML(deployNS, collectorSA, "evil-pod"))
		Expect(err).To(HaveOccurred(), "expected Pod creation to be denied, got: %s", out)
		Expect(out).To(ContainSubstring("protected collector ServiceAccount"))
	})

	It("denies a Deployment referencing the protected SA", func() {
		out, err := ocCreateAs(deployNS, user, deploymentYAML(deployNS, collectorSA, "evil-deploy"))
		Expect(err).To(HaveOccurred(), "expected Deployment creation to be denied, got: %s", out)
		Expect(out).To(ContainSubstring("protected collector ServiceAccount"))
	})

	It("allows a Pod referencing an unprotected SA (no collateral damage)", func() {
		out, err := ocCreateAs(deployNS, user, spoofedPodYAML(deployNS, unprotectedSA, "plain-pod"))
		Expect(err).ToNot(HaveOccurred(), "expected Pod with unprotected SA to be admitted, got: %s", out)
	})
})

func checkVAPInstalled() {
	for _, resource := range []string{
		"validatingadmissionpolicy/" + admission.ProtectedSAPodsPolicyName,
		"validatingadmissionpolicybinding/" + admission.ProtectedSAPodsBindingName,
		"validatingadmissionpolicy/" + admission.ProtectedSAWorkloadsPolicyName,
		"validatingadmissionpolicybinding/" + admission.ProtectedSAWorkloadsBindingName,
	} {
		out, err := exec.Command("oc", "get", resource).CombinedOutput()
		Expect(err).ToNot(HaveOccurred(), "expected %s to exist (operator must reconcile VAP): %s", resource, out)
	}
}

// waitForSAProtected polls the param ConfigMap until the operator has added the
// key for the given ServiceAccount. The key format is sa_<ns>_<name> (see
// admission.protectedSAKeyPrefix); it can't be read via jsonpath because SA
// names may contain dots, so we match it in the raw JSON.
func waitForSAProtected(namespace, sa string) {
	key := fmt.Sprintf("sa_%s_%s", namespace, sa)
	Eventually(func(g Gomega) {
		out, err := exec.Command("oc", "get", "configmap", admission.ProtectedSAConfigMapName,
			"-n", constants.OpenshiftNS, "-o", "json").CombinedOutput()
		g.Expect(err).ToNot(HaveOccurred(), "get param ConfigMap: %s", out)
		g.Expect(string(out)).To(ContainSubstring(key),
			"operator did not mark %s/%s protected (is CLO running the protected-SA reconciler?)", namespace, sa)
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

// spoofedPodYAML copies the collector's visible metadata to prove that spoofing
// labels/annotations/name does not bypass the policy.
func spoofedPodYAML(namespace, sa, name string) string {
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
`, name, namespace, clfName, sa)
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
