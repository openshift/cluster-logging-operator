package sa_authorization

import (
	"fmt"
	"os/exec"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/openshift/cluster-logging-operator/internal/admission"
	framework "github.com/openshift/cluster-logging-operator/test/framework/e2e"
)

// This test validates CVE-2026-10609 fix:
// A ValidatingAdmissionPolicy enforces that users must have 'use' permission
// on a ServiceAccount to reference it in a ClusterLogForwarder.
//
// It mirrors hack/test-sa-authorization.sh: admission-only checks without
// collector RBAC or a ready CLF status.

var _ = Describe("[CVE-2026-10609] ServiceAccount usage authorization", Ordered, func() {

	const (
		clfName           = "sa-auth-test"
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
		user     string
		clfYaml  string
	)

	BeforeAll(func() {
		e2e = framework.NewE2ETestFramework()
		deployNS = e2e.Test.NS.Name
		user = fmt.Sprintf(restrictedUser, deployNS)
		clfYaml = fmt.Sprintf(clfWithSATokenFmt, collectorSAName)

		checkVAPInstalled()

		_, err := e2e.BuildAuthorizationFor(deployNS, collectorSAName).Create()
		Expect(err).ToNot(HaveOccurred())

		_, err = e2e.BuildAuthorizationFor(deployNS, "restricted-user").Create()
		Expect(err).ToNot(HaveOccurred())

		grantCLFAccess(deployNS)
	})

	AfterAll(func() {
		if e2e != nil {
			e2e.Cleanup()
		}
	})

	It("should reject CLF creation when user cannot 'use' the referenced ServiceAccount", func() {
		revokeSAUsage(deployNS)

		out, err := ocCreateAs(deployNS, user, clfYaml)
		Expect(err).To(HaveOccurred(), "Expected CLF creation to be rejected, but got: %s", out)
		Expect(out).To(ContainSubstring("not authorized to use"))
	})

	It("should allow CLF creation when user has 'use' permission on the ServiceAccount", func() {
		grantSAUsage(deployNS, "restricted-user", collectorSAName)

		out, err := ocCreateAs(deployNS, user, clfYaml)
		Expect(err).ToNot(HaveOccurred(), "Expected CLF creation to succeed, but got error: %s %v", out, err)
	})

	It("should reject CLF update when user cannot 'use' the referenced ServiceAccount", func() {
		revokeSAUsage(deployNS)

		out, err := ocPatchCLFAs(deployNS, user, clfName, `{"metadata":{"annotations":{"sa-auth-test":"update-attempt"}}}`)
		Expect(err).To(HaveOccurred(), "Expected CLF update to be rejected, but got: %s", out)
		Expect(out).To(ContainSubstring("not authorized to use"))
	})

	It("should allow CLF deletion when user cannot 'use' the referenced ServiceAccount", func() {
		revokeSAUsage(deployNS)

		err := ocDeleteCLFAs(deployNS, user, clfName)
		Expect(err).ToNot(HaveOccurred(), "Expected CLF deletion to succeed without 'use' permission")
	})
})

func checkVAPInstalled() {
	for _, resource := range []string{
		"validatingadmissionpolicy/" + admission.SAUsagePolicyName,
		"validatingadmissionpolicybinding/" + admission.SAUsagePolicyBindingName,
	} {
		cmd := exec.Command("oc", "get", resource)
		out, err := cmd.CombinedOutput()
		Expect(err).ToNot(HaveOccurred(), "Expected %s to exist (operator must reconcile VAP): %s", resource, out)
	}
}

func ocCreateAs(namespace, user, yaml string) (string, error) {
	cmd := exec.Command("oc", "create", "-n", namespace, "--as", user, "-f", "-")
	cmd.Stdin = strings.NewReader(yaml)
	out, err := cmd.CombinedOutput()
	return string(out), err
}

func ocPatchCLFAs(namespace, user, name, patch string) (string, error) {
	cmd := exec.Command("oc", "patch", "clusterlogforwarder", name,
		"-n", namespace, "--as", user, "--type=merge", "-p", patch)
	out, err := cmd.CombinedOutput()
	return string(out), err
}

func ocDeleteCLFAs(namespace, user, name string) error {
	cmd := exec.Command("oc", "delete", "clusterlogforwarder", name,
		"-n", namespace, "--as", user, "--ignore-not-found")
	_, err := cmd.CombinedOutput()
	return err
}

func grantCLFAccess(namespace string) {
	cmd := exec.Command("oc", "create", "role", "clf-editor",
		"--verb=create,update,patch,get,list,watch,delete",
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

func revokeSAUsage(namespace string) {
	_ = exec.Command("oc", "delete", "rolebinding", "restricted-user-sa-user",
		"-n", namespace, "--ignore-not-found").Run()
	_ = exec.Command("oc", "delete", "role", "sa-user",
		"-n", namespace, "--ignore-not-found").Run()
}

func grantSAUsage(namespace, userName, saName string) {
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
