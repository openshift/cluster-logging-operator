package oc_test

import (
	"os"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/openshift/cluster-logging-operator/pkg/logger"
	"github.com/openshift/cluster-logging-operator/test/helpers/oc"
)

var _ = Describe("oc exec pod", func() {
	Describe("with given pod name", func() {
		Context("running a command without args", func() {
			Describe("String() invocation", func() {
				It("should form equivalent string", func() {
					occmd := oc.Exec().
						WithNamespace("openshift-logging").
						Pod("mypod").
						Container("elasticsearch").
						WithCmd("indices")
					strcmd := "oc -n openshift-logging exec mypod -c elasticsearch -- indices"
					Expect(occmd.String()).To(Equal(strcmd))
				})
			})
		})
		Context("running a command with args", func() {
			Describe("String() invocation", func() {
				It("should form equivalent string", func() {
					occmd := oc.Exec().
						WithNamespace("openshift-logging").
						Pod("mypod").
						Container("elasticsearch").
						WithCmd("es_util", " --query=\"_cat/aliases?v&bytes=m\"")
					strcmd := "oc -n openshift-logging exec mypod -c elasticsearch -- es_util --query=\"_cat/aliases?v&bytes=m\""
					Expect(occmd.String()).To(Equal(strcmd))
				})
			})
			Describe("Run() invocation", func() {
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
					Expect(oc.Literal().From("oc -n test-log-gen wait --for=condition=Ready pod/log-generator").Output()).To(Succeed())
				})
				It("should not result in error", func() {
					Expect(oc.Literal().From("oc -n test-log-gen logs log-generator -f").OutputFor(time.Second * 10)).To(Succeed())
					occmd := oc.Exec().
						WithNamespace("test-log-gen").
						Pod("log-generator").
						Container("log-generator").
						WithCmd("ls", "-al")
					_, err := occmd.Run()
					if err != nil {
						Fail("failed to run the exec command")
					}
				})
				AfterEach(func() {
					Expect(oc.Literal().From("oc delete ns test-log-gen").Run()).To(Succeed())
					if tmpFile != nil {
						_ = os.Remove(tmpFile.Name())
					} else {
						logger.Error("tmpfile is nil")
					}
				})
			})
		})
	})
	Describe("with a composed pod getter", func() {
		Context("running a command with args", func() {
			Describe("String() invocation", func() {
				It("should form equivalent string", func() {
					occmd := oc.Exec().
						WithNamespace("openshift-logging").
						WithPodGetter(oc.Get().
							WithNamespace("openshift-logging").
							Pod().
							Selector("component=elasticsearch").
							OutputJsonpath("{.items[0].metadata.name}")).
						Container("elasticsearch").
						WithCmd("es_util", " --query=\"_cat/aliases?v&bytes=m\"")
					strcmd := "oc -n openshift-logging exec $(oc -n openshift-logging get pod -l component=elasticsearch -o jsonpath={.items[0].metadata.name}) -c elasticsearch -- es_util --query=\"_cat/aliases?v&bytes=m\""
					Expect(occmd.String()).To(Equal(strcmd))
				})
			})
			Describe("Run() invocation", func() {
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
					Expect(oc.Literal().From("oc -n test-log-gen wait --for=condition=Ready pod/log-generator").Output()).To(Succeed())
				})
				It("should not result in error", func() {
					Expect(oc.Literal().From("oc -n test-log-gen logs log-generator -f").OutputFor(time.Second * 10)).To(Succeed())
					occmd := oc.Exec().
						WithNamespace("test-log-gen").
						WithPodGetter(
							oc.Get().
								WithNamespace("test-log-gen").
								Pod().
								Selector("component=test").
								OutputJsonpath("{.items[0].metadata.name}")).
						Container("log-generator").
						WithCmd("ls", "-al")
					_, err := occmd.Run()
					if err != nil {
						Fail("failed to run the exec command")
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
})
