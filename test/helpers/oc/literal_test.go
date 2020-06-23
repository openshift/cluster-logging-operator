package oc_test

import (
	"os"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

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
	})
	Context("", func() {
		var tmpFile *os.File
		BeforeEach(func() {
			f, err := os.Create("./podspec.yaml")
			if err != nil {
				Fail("failed to create temp file")
			}
			if _, err = f.Write([]byte(podSpec)); err != nil {
				Fail("failed to write to temp file")
			}
			tmpFile = f
		})
		It("should run complete pod cycle", func() {
			Expect(oc.Literal().From("oc create ns test-log-gen").Output()).To(Succeed())
			Expect(oc.Literal().From("oc apply -f ./podspec.yaml").Output()).To(Succeed())
			Expect(oc.Literal().From("oc -n test-log-gen get pod -l component=test -o jsonpath={.items[0].metadata.name}").Output()).To(Succeed())
			Expect(oc.Literal().From("oc -n test-log-gen wait --for=condition=Ready pod/log-generator").Output()).To(Succeed())
			Expect(oc.Literal().From("oc -n test-log-gen logs log-generator -f").OutputFor(time.Second * 10)).To(Succeed())
			// currently oc.Literal for oc exec does not support bash -c commands
			Expect(oc.Literal().From("oc -n test-log-gen exec log-generator -c log-generator -- ls -al").Output()).To(Succeed())
			Expect(oc.Literal().From("oc delete ns test-log-gen").Output()).To(Succeed())
		})
		AfterEach(func() {
			if tmpFile != nil {
				os.Remove(tmpFile.Name())
			}
		})
	})
})
