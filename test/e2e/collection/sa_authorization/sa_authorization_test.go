package sa_authorization

import (
	"fmt"
	"os/exec"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	framework "github.com/openshift/cluster-logging-operator/test/framework/e2e"
)

// This test validates CVE-2026-10609 fix:
// A ValidatingAdmissionPolicy enforces that users must have 'use' permission
// on a ServiceAccount to reference it in a ClusterLogForwarder.

var _ = Describe("[CVE-2026-10609] ServiceAccount usage authorization", func() {

	const (
		collectorSAName   = "log-collector"
		restrictedUser    = "system:serviceaccount:%s:restricted-user"
		clfWithSATokenFmt = `
apiVersion: observability.openshift.io/v1
kind: ClusterLogForwarder
metadata:
  name: sa-auth-test
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
  - name: test-pipeline
    inputRefs:
    - application
    outputRefs:
    - test-output
`
	)

	var (
		e2e      *framework.E2ETestFramework
		deployNS string
	)

	BeforeEach(func() {
		e2e = framework.NewE2ETestFramework()
		deployNS = e2e.Test.NS.Name

		// Create the collector SA with collect permissions
		_, err := e2e.BuildAuthorizationFor(deployNS, collectorSAName).
			AllowClusterRole(framework.ClusterRoleCollectApplicationLogs).
			AllowClusterRole(framework.ClusterRoleCollectInfrastructureLogs).
			Create()
		Expect(err).ToNot(HaveOccurred())

		// Create a restricted user SA (used for impersonation)
		_, err = e2e.BuildAuthorizationFor(deployNS, "restricted-user").Create()
		Expect(err).ToNot(HaveOccurred())

		// Grant restricted-user permission to create CLFs in the namespace
		grantCLFAccess(deployNS)
	})

	AfterEach(func() {
		if e2e != nil {
			e2e.Cleanup()
		}
	})

	It("should reject CLF creation when user cannot 'use' the referenced ServiceAccount", func() {
		user := fmt.Sprintf(restrictedUser, deployNS)
		clfYaml := fmt.Sprintf(clfWithSATokenFmt, collectorSAName)

		// Create CLF as restricted-user who does NOT have 'use' permission on the collector SA
		out, err := ocCreateAs(deployNS, user, clfYaml)
		Expect(err).To(HaveOccurred(), "Expected CLF creation to be rejected, but got: %s", out)
		Expect(out).To(ContainSubstring("not authorized to use"))
	})

	It("should allow CLF creation when user has 'use' permission on the ServiceAccount", func() {
		user := fmt.Sprintf(restrictedUser, deployNS)
		clfYaml := fmt.Sprintf(clfWithSATokenFmt, collectorSAName)

		// Grant 'use' permission on the collector SA to restricted-user
		grantSAUsage(deployNS, "restricted-user", collectorSAName)

		// Create CLF as restricted-user who now HAS 'use' permission
		out, err := ocCreateAs(deployNS, user, clfYaml)
		Expect(err).ToNot(HaveOccurred(), "Expected CLF creation to succeed, but got error: %s %v", out, err)
	})
})

func ocCreateAs(namespace, user, yaml string) (string, error) {
	cmd := exec.Command("oc", "create", "-n", namespace, "--as", user, "-f", "-")
	cmd.Stdin = strings.NewReader(yaml)
	out, err := cmd.CombinedOutput()
	return string(out), err
}

func grantCLFAccess(namespace string) {
	cmd := exec.Command("oc", "create", "role", "clf-editor",
		"--verb=create,update,get,list,watch,delete",
		"--resource=clusterlogforwarders.observability.openshift.io",
		"-n", namespace)
	out, err := cmd.CombinedOutput()
	if err != nil {
		Fail(fmt.Sprintf("Failed to create clf-editor role: %s %v", out, err))
	}

	cmd = exec.Command("oc", "create", "rolebinding", "restricted-user-clf-editor",
		"--role=clf-editor",
		"--serviceaccount="+namespace+":restricted-user",
		"-n", namespace)
	out, err = cmd.CombinedOutput()
	if err != nil {
		Fail(fmt.Sprintf("Failed to create rolebinding: %s %v", out, err))
	}
}

func grantSAUsage(namespace, userName, saName string) {
	// 'use' is not a standard Kubernetes verb, so 'oc create role' rejects it.
	// Create the Role via YAML instead.
	roleYaml := fmt.Sprintf(`apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: sa-user
  namespace: %s
rules:
- apiGroups: [""]
  resources: ["serviceaccounts"]
  resourceNames: ["%s"]
  verbs: ["use"]`, namespace, saName)

	cmd := exec.Command("oc", "create", "-f", "-")
	cmd.Stdin = strings.NewReader(roleYaml)
	out, err := cmd.CombinedOutput()
	if err != nil {
		Fail(fmt.Sprintf("Failed to create sa-user role: %s %v", out, err))
	}

	cmd = exec.Command("oc", "create", "rolebinding", "restricted-user-sa-user",
		"--role=sa-user",
		"--serviceaccount="+namespace+":"+userName,
		"-n", namespace)
	out, err = cmd.CombinedOutput()
	if err != nil {
		Fail(fmt.Sprintf("Failed to create rolebinding: %s %v", out, err))
	}
}
