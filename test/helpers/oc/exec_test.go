package oc_test

import (
	"fmt"
	"os"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/ViaQ/logerr/v2/log"
	"github.com/openshift/cluster-logging-operator/test"
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
				var logGenNSName string
				BeforeEach(func() {
					logGenNSName = test.UniqueNameForTest()
					specfile := fmt.Sprintf("%s/podspec.yaml", os.TempDir())
					f, err := os.Create(fmt.Sprintf("%s/podspec.yaml", os.TempDir()))
					if err != nil {
						Fail("failed to create temp file")
					}
					podSpec := fmt.Sprintf(podSpecTemplate, logGenNSName)
					if _, err = f.Write([]byte(podSpec)); err != nil {
						Fail("failed to write to temp file")
					}
					if _, err = oc.Literal().From("oc create ns %s", logGenNSName).Run(); err != nil {
						Fail("failed to create namespace")
					}
					if _, err = oc.Literal().From("oc apply -f %s", specfile).Run(); err != nil {
						Fail("failed to create pod")
					}
					tmpFile = f
					Expect(oc.Literal().From("oc -n %s wait --timeout=120s --for=condition=Ready pod/log-generator", logGenNSName).Output()).To(Succeed())
				})
				It("should not result in error", func() {
					Expect(oc.Literal().From("oc -n %s logs log-generator -f", logGenNSName).OutputFor(time.Second * 10)).To(BeNil())
					occmd := oc.Exec().
						WithNamespace(logGenNSName).
						Pod("log-generator").
						Container("log-generator").
						WithCmd("ls", "-al")
					_, err := occmd.Run()
					if err != nil {
						Fail("failed to run the exec command")
					}
				})
				AfterEach(func() {
					Expect(oc.Literal().From("oc delete ns %s", logGenNSName).Output()).To(Succeed())
					if tmpFile != nil {
						_ = os.Remove(tmpFile.Name())
					} else {
						log.NewLogger("").Info("tmpfile is nil")
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
				var logGenNSName string
				BeforeEach(func() {
					logGenNSName = test.UniqueNameForTest()
					specfile := fmt.Sprintf("%s/podspec.yaml", os.TempDir())
					f, err := os.Create(specfile)
					if err != nil {
						Fail("failed to create temp file")
					}
					podSpec := fmt.Sprintf(podSpecTemplate, logGenNSName)
					if _, err = f.Write([]byte(podSpec)); err != nil {
						Fail("failed to write to temp file")
					}
					if _, err = oc.Literal().From("oc create ns %s", logGenNSName).Run(); err != nil {
						Fail("failed to create namespace")
					}
					if _, err = oc.Literal().From("oc apply -f %s", specfile).Run(); err != nil {
						Fail("failed to create pod")
					}
					tmpFile = f
					Expect(oc.Literal().From("oc -n %s wait --timeout=120s --for=condition=Ready pod/log-generator", logGenNSName).Output()).To(Succeed())
				})
				It("should not result in error", func() {
					Expect(oc.Literal().From("oc -n %s logs log-generator -f", logGenNSName).OutputFor(time.Second * 10)).To(Succeed())
					occmd := oc.Exec().
						WithNamespace(logGenNSName).
						WithPodGetter(
							oc.Get().
								WithNamespace(logGenNSName).
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
					Expect(oc.Literal().From("oc delete ns %s", logGenNSName).Output()).To(Succeed())
					if tmpFile != nil {
						os.Remove(tmpFile.Name())
					} else {
						log.NewLogger("").Info("tmpfile is nil")
					}
				})
			})
		})
	})
})
