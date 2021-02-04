package oc_test

import (
	"fmt"
	"os"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/openshift/cluster-logging-operator/test"
	"github.com/openshift/cluster-logging-operator/test/helpers/oc"
)

var _ = Describe("literal command", func() {
	Context("", func() {
		It("should tolerate extra spaces", func() {
			cmd := oc.Literal().From(" oc  apply    -f   ./tmp/podspec.yaml ")
			Expect(cmd.String()).To(BeIdenticalTo("oc apply -f ./tmp/podspec.yaml"))
			cmd = oc.Literal().From("oc   ")
			Expect(cmd.String()).To(BeIdenticalTo("command too small"))
			cmd = oc.Literal().From("abc xyz")
			Expect(cmd.String()).To(BeIdenticalTo("error: command string must start with 'oc'"))
		})
		It("From() should support variadic args", func() {
			cmd := oc.Literal().From("oc -n %s get pod/%s -o yaml", "openshift-logging", "fluentd")
			Expect(cmd.String()).To(BeIdenticalTo("oc -n openshift-logging get pod/fluentd -o yaml"))
		})
	})
	Context("", func() {
		var tmpFile *os.File
		var logGenNSName string
		var specfile string
		BeforeEach(func() {
			logGenNSName = test.UniqueNameForTest()
			specfile = fmt.Sprintf("%s/podspec.yaml", os.TempDir())
			f, err := os.Create(specfile)
			if err != nil {
				Fail("failed to create temp file")
			}
			podSpec := fmt.Sprintf(podSpecTemplate, logGenNSName)
			if _, err = f.Write([]byte(podSpec)); err != nil {
				Fail("failed to write to temp file")
			}
			tmpFile = f
		})
		It("should run complete pod cycle", func() {
			Expect(oc.Literal().From("oc create ns %s", logGenNSName).Output()).To(Succeed())
			Expect(oc.Literal().From("oc apply -f %s", specfile).Output()).To(Succeed())
			Expect(oc.Literal().From("oc -n %s get pod -l component=test -o jsonpath={.items[0].metadata.name}", logGenNSName).Output()).To(Succeed())
			Expect(oc.Literal().From("oc -n %s wait --for=condition=Ready pod/log-generator", logGenNSName).Output()).To(Succeed())
			Expect(oc.Literal().From("oc -n %s logs log-generator -f", logGenNSName).OutputFor(time.Second * 10)).To(Succeed())
			// currently oc.Literal for oc exec does not support bash -c commands
			Expect(oc.Literal().From("oc -n %s exec log-generator -c log-generator -- ls -al", logGenNSName).Output()).To(Succeed())
			Expect(oc.Literal().From("oc delete ns %s", logGenNSName).Output()).To(Succeed())
		})
		AfterEach(func() {
			if tmpFile != nil {
				os.Remove(tmpFile.Name())
			}
		})
	})
})
