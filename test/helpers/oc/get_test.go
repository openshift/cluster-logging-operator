package oc_test

import (
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/openshift/cluster-logging-operator/pkg/logger"
	"github.com/openshift/cluster-logging-operator/test/helpers/oc"
)

var _ = Describe("oc get pod", func() {
	Context("with selector", func() {
		Describe("String() invocation", func() {
			It("should form equivalent command strings", func() {
				occmd := oc.Get().
					Pod().
					WithNamespace("test-log-gen").
					Selector("component=test").
					OutputJsonpath("{.items[0].metadata.name}")
				strcmd := "oc -n test-log-gen get pod -l component=test -o jsonpath={.items[0].metadata.name}"
				Expect(occmd.String()).To(Equal(strcmd))
			})
		})
		Describe("invocation", func() {
			var tmpFile *os.File
			BeforeEach(func() {
				f, err := os.Create("./podspec.yaml")
				if err != nil {
					Fail("failed to create temp file")
				}
				if _, err = f.Write([]byte(podSpec)); err != nil {
					Fail("failed to write to temp file")
				}
				if _, err = oc.Literal().From("oc create ns test-log-gen").Run(); err != nil {
					Fail("failed to create namespace")
				}
				if _, err = oc.Literal().From("oc apply -f ./podspec.yaml").Run(); err != nil {
					Fail("failed to create pod")
				}
				tmpFile = f
			})
			It("should not result in error", func() {
				occmd := oc.Get().
					WithNamespace("test-log-gen").
					Pod().
					Selector("component=test").
					OutputJsonpath("{.items[0].metadata.name}")
				str, err := occmd.Run()
				if err != nil {
					Fail("failed to run the command")
				}
				if str != "log-generator" {
					Fail("received incorrect pod name")
				}
			})
			AfterEach(func() {
				Expect(oc.Literal().From("oc delete ns test-log-gen").Run()).To(Succeed())
				if tmpFile != nil {
					os.Remove(tmpFile.Name())
				} else {
					logger.Error("tmpfile is nil")
				}
			})
		})
	})
})
