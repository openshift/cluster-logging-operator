package apivalidations

import (
	"bytes"
	_ "embed"
	"fmt"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	framework "github.com/openshift/cluster-logging-operator/test/framework/e2e"
	"github.com/openshift/cluster-logging-operator/test/helpers/cmd"
	"os/exec"
)

var _ = Describe("", func() {

	const name = "clf-validation-test"
	var (
		e2e *framework.E2ETestFramework
	)
	AfterEach(func() {
		if e2e != nil {
			e2e.Cleanup()
		}
	})

	DescribeTable("Verifying declarative API validations", func(crFile string, assert func(string, error)) {

		e2e = framework.NewE2ETestFramework()
		crYaml, err := tomlContent.ReadFile(crFile)
		if err != nil {
			Fail(fmt.Sprintf("Error reading the file %q with exp config: %v", crFile, err))
		}
		deployNS := e2e.Test.NS.Name
		if _, err := e2e.BuildAuthorizationFor(deployNS, name).
			AllowClusterRole(framework.ClusterRoleCollectApplicationLogs).
			AllowClusterRole(framework.ClusterRoleCollectInfrastructureLogs).
			AllowClusterRole(framework.ClusterRoleCollectAuditLogs).Create(); err != nil {
			Fail(err.Error())
		}

		execCMD := exec.Command("sh", "-c", fmt.Sprintf("echo '%s' | oc -n %s create -f -", crYaml, deployNS))
		reader, err := cmd.NewReader(execCMD)
		Expect(err).ToNot(HaveOccurred())
		defer reader.Close()
		buffer := bytes.NewBuffer([]byte{})
		_, err = buffer.ReadFrom(reader)
		assert(buffer.String(), err)
	},
		Entry("LOG-5788: for multilineException filter should not fail", "log5788_mulitiline_ex_filter.yaml", func(out string, err error) {
			Expect(err).ToNot(HaveOccurred())
			Expect(out).To(MatchRegexp("clusterlogforwarder.*created"))
		}),
		Entry("LOG-5793: for lokiStack bearer token from SA should not fail", "log5793_bearer_token_from_sa.yaml", func(out string, err error) {
			Expect(err).ToNot(HaveOccurred())
			Expect(out).To(MatchRegexp("clusterlogforwarder.*created"))
		}),
		Entry("should fail with invalid name", "invalid_name.yaml", func(out string, err error) {
			Expect(err.Error()).To(MatchRegexp("Name.*valid DNS1035"))
		}),
	)
})
