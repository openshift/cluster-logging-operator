package syslog

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"runtime"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	apps "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"

	"github.com/ViaQ/logerr/log"
	logging "github.com/openshift/cluster-logging-operator/pkg/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/pkg/k8shandler"
	"github.com/openshift/cluster-logging-operator/test/helpers"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	// grep "pod_name"=>"<generator-pod-name>" <syslog-log-filename>
	rsyslogFormatStr = `grep %s %%s| grep pod_name | tail -n 1 | awk -F' ' '{print %s}'`
)

var _ = Describe("[ClusterLogForwarder] Forwards logs", func() {
	_, filename, _, _ := runtime.Caller(0)
	log.Info("Running ", "filename", filename)
	var (
		err              error
		syslogDeployment *apps.Deployment
		e2e              = helpers.NewE2ETestFramework()
		testDir          string
		forwarder        *logging.ClusterLogForwarder
		generatorPayload map[string]string
		waitlogs         string
		grepappname      string
		grepprocid       string
		grepmsgid        string
		greptag          string
		logGenPod        string
		logGenNS         string
	)
	BeforeEach(func() {
		generatorPayload = map[string]string{
			"msgcontent": "My life is my message",
		}
	})
	JustBeforeEach(func() {
		if logGenNS, logGenPod, err = e2e.DeployJsonLogGenerator(generatorPayload); err != nil {
			log.Error(err, "unable to deploy log generator.")
		}
		log.Info("log generator pod: ", "podname", logGenPod)
		testDir = filepath.Dir(filename)
		// wait for current log-generator's logs to appear in syslog
		waitlogs = fmt.Sprintf(`[ $(grep %s %%s |grep pod_name| wc -l) -gt 0 ]`, logGenPod)
		grepappname = fmt.Sprintf(rsyslogFormatStr, logGenPod, "$4")
		greptag = fmt.Sprintf(rsyslogFormatStr, logGenPod, "$4")
		grepprocid = fmt.Sprintf(rsyslogFormatStr, logGenPod, "$5")
		grepmsgid = fmt.Sprintf(rsyslogFormatStr, logGenPod, "$6")
	})
	Describe("when the output is a third-party managed syslog", func() {
		BeforeEach(func() {
			cr := helpers.NewClusterLogging(helpers.ComponentTypeCollector)
			if err := e2e.CreateClusterLogging(cr); err != nil {
				Fail(fmt.Sprintf("Unable to create an instance of cluster logging: %v", err))
			}
		})
		Describe("with rfc5424", func() {
			BeforeEach(func() {
				forwarder = &logging.ClusterLogForwarder{
					TypeMeta: metav1.TypeMeta{
						Kind:       logging.ClusterLogForwarderKind,
						APIVersion: logging.SchemeGroupVersion.String(),
					},
					ObjectMeta: metav1.ObjectMeta{
						Name: "instance",
					},
					Spec: logging.ClusterLogForwarderSpec{
						Outputs: []logging.OutputSpec{
							{
								Name: "syslogout",
								Type: "syslog",
								OutputTypeSpec: logging.OutputTypeSpec{
									Syslog: &logging.Syslog{
										Facility: "user",
										Severity: "debug",
										AppName:  "myapp",
										ProcID:   "myproc",
										MsgID:    "mymsg",
										RFC:      "RFC5424",
									},
								},
							},
						},
						Pipelines: []logging.PipelineSpec{
							{
								Name:       "test-app",
								OutputRefs: []string{"syslogout"},
								InputRefs:  []string{"application"},
							},
						},
					},
				}
			})
			DescribeTable("should be able to send logs to syslog receiver", func(tls bool, protocol corev1.Protocol) {
				if syslogDeployment, err = e2e.DeploySyslogReceiver(testDir, protocol, tls, helpers.RFC5424); err != nil {
					Fail(fmt.Sprintf("Unable to deploy syslog receiver: %v", err))
				}
				if protocol == corev1.ProtocolTCP {
					forwarder.Spec.Outputs[0].URL = fmt.Sprintf("tcp://%s.%s.svc:24224", syslogDeployment.ObjectMeta.Name, syslogDeployment.Namespace)
				} else {
					forwarder.Spec.Outputs[0].URL = fmt.Sprintf("udp://%s.%s.svc:24224", syslogDeployment.ObjectMeta.Name, syslogDeployment.Namespace)
				}

				if tls {
					forwarder.Spec.Outputs[0].Secret = &logging.OutputSecretSpec{
						Name: syslogDeployment.ObjectMeta.Name,
					}
				}
				if err := e2e.CreateClusterLogForwarder(forwarder); err != nil {
					Fail(fmt.Sprintf("Unable to create an instance of logforwarder: %v", err))
				}
				components := []helpers.LogComponentType{helpers.ComponentTypeCollector}
				for _, component := range components {
					if err := e2e.WaitFor(component); err != nil {
						Fail(fmt.Sprintf("Failed waiting for component %s to be ready: %v", component, err))
					}
				}
				logStore := e2e.LogStores[syslogDeployment.GetName()]
				Expect(logStore.HasInfraStructureLogs(helpers.DefaultWaitForLogsTimeout)).To(BeTrue(), "Expected to find stored infrastructure logs")
				_, _ = logStore.GrepLogs(waitlogs, helpers.DefaultWaitForLogsTimeout)
				expectedAppName := forwarder.Spec.Outputs[0].Syslog.AppName
				Expect(logStore.GrepLogs(grepappname, helpers.DefaultWaitForLogsTimeout)).To(Equal(expectedAppName), "Expected: "+expectedAppName)
				expectedMsgID := forwarder.Spec.Outputs[0].Syslog.MsgID
				Expect(logStore.GrepLogs(grepmsgid, helpers.DefaultWaitForLogsTimeout)).To(Equal(expectedMsgID), "Expected: "+expectedMsgID)
				expectedProcID := forwarder.Spec.Outputs[0].Syslog.ProcID
				Expect(logStore.GrepLogs(grepprocid, helpers.DefaultWaitForLogsTimeout)).To(Equal(expectedProcID), "Expected: "+expectedProcID)
			},
				Entry("with TLS disabled, with TCP", false, corev1.ProtocolTCP),
				Entry("with TLS disabled, with UDP", false, corev1.ProtocolUDP),
				Entry("with TLS enabled, with TCP", true, corev1.ProtocolTCP),
				Entry("with TLS enabled, with UDP", true, corev1.ProtocolUDP),
			)
			Describe("values for appname", func() {
				Context("and msgid,procid", func() {
					BeforeEach(func() {
						generatorPayload["appname_key"] = "rec_appname"
						generatorPayload["msgid_key"] = "rec_msgid"
						generatorPayload["procid_key"] = "rec_procid"
					})
					It("should use values from record", func() {
						if syslogDeployment, err = e2e.DeploySyslogReceiver(testDir, corev1.ProtocolTCP, true, helpers.RFC5424); err != nil {
							Fail(fmt.Sprintf("Unable to deploy syslog receiver: %v", err))
						}
						forwarder.Spec.Outputs[0].URL = fmt.Sprintf("tcp://%s.%s.svc:24224", syslogDeployment.ObjectMeta.Name, syslogDeployment.Namespace)
						forwarder.Spec.Outputs[0].Secret = &logging.OutputSecretSpec{
							Name: syslogDeployment.ObjectMeta.Name,
						}
						forwarder.Spec.Outputs[0].Syslog.RFC = helpers.RFC5424.String()
						forwarder.Spec.Outputs[0].Syslog.AppName = "$.message.appname_key"
						forwarder.Spec.Outputs[0].Syslog.MsgID = "$.message.msgid_key"
						forwarder.Spec.Outputs[0].Syslog.ProcID = "$.message.procid_key"
						if err := e2e.CreateClusterLogForwarder(forwarder); err != nil {
							Fail(fmt.Sprintf("Unable to create an instance of logforwarder: %v", err))
						}
						components := []helpers.LogComponentType{helpers.ComponentTypeCollector}
						for _, component := range components {
							if err := e2e.WaitFor(component); err != nil {
								Fail(fmt.Sprintf("Failed waiting for component %s to be ready: %v", component, err))
							}
						}
						logStore := e2e.LogStores[syslogDeployment.GetName()]
						Expect(logStore.HasInfraStructureLogs(helpers.DefaultWaitForLogsTimeout)).To(BeTrue(), "Expected to find stored infrastructure logs")
						_, _ = logStore.GrepLogs(waitlogs, helpers.DefaultWaitForLogsTimeout)
						expectedAppName := generatorPayload["appname_key"]
						Expect(logStore.GrepLogs(grepappname, helpers.DefaultWaitForLogsTimeout)).To(Equal(expectedAppName), "Expected: "+expectedAppName)
						expectedMsgID := generatorPayload["msgid_key"]
						Expect(logStore.GrepLogs(grepmsgid, helpers.DefaultWaitForLogsTimeout)).To(Equal(expectedMsgID), "Expected: "+expectedMsgID)
						expectedProcID := generatorPayload["procid_key"]
						Expect(logStore.GrepLogs(grepprocid, helpers.DefaultWaitForLogsTimeout)).To(Equal(expectedProcID), "Expected: "+expectedProcID)
					})
				})
				It("should take value from complete fluentd tag", func() {
					if syslogDeployment, err = e2e.DeploySyslogReceiver(testDir, corev1.ProtocolTCP, true, helpers.RFC5424); err != nil {
						Fail(fmt.Sprintf("Unable to deploy syslog receiver: %v", err))
					}
					forwarder.Spec.Outputs[0].URL = fmt.Sprintf("tcp://%s.%s.svc:24224", syslogDeployment.ObjectMeta.Name, syslogDeployment.Namespace)
					forwarder.Spec.Outputs[0].Secret = &logging.OutputSecretSpec{
						Name: syslogDeployment.ObjectMeta.Name,
					}
					forwarder.Spec.Outputs[0].Syslog.RFC = helpers.RFC5424.String()
					forwarder.Spec.Outputs[0].Syslog.AppName = "tag"
					if err := e2e.CreateClusterLogForwarder(forwarder); err != nil {
						Fail(fmt.Sprintf("Unable to create an instance of logforwarder: %v", err))
					}
					components := []helpers.LogComponentType{helpers.ComponentTypeCollector}
					for _, component := range components {
						if err := e2e.WaitFor(component); err != nil {
							Fail(fmt.Sprintf("Failed waiting for component %s to be ready: %v", component, err))
						}
					}
					logStore := e2e.LogStores[syslogDeployment.GetName()]
					Expect(logStore.HasInfraStructureLogs(helpers.DefaultWaitForLogsTimeout)).To(BeTrue(), "Expected to find stored infrastructure logs")
					_, _ = logStore.GrepLogs(waitlogs, helpers.DefaultWaitForLogsTimeout)
					// typical value of tag in this test case is: kubernetes.var.log.containers.log-generator-746f659fdf-qgclg_clo-test-5764_log-generator-fef2a7848f9741bc6aeb1325aac051c0734c5dc177839c6787da207dc95530ad.log
					expectedAppNamePrefix := "kubernetes.var.log.containers"
					Expect(logStore.GrepLogs(grepappname, helpers.DefaultWaitForLogsTimeout)).To(HavePrefix(expectedAppNamePrefix))
					expectedMsgID := forwarder.Spec.Outputs[0].Syslog.MsgID
					Expect(logStore.GrepLogs(grepmsgid, helpers.DefaultWaitForLogsTimeout)).To(Equal(expectedMsgID), "Expected: "+expectedMsgID)
					expectedProcID := forwarder.Spec.Outputs[0].Syslog.ProcID
					Expect(logStore.GrepLogs(grepprocid, helpers.DefaultWaitForLogsTimeout)).To(Equal(expectedProcID), "Expected: "+expectedProcID)
				})
				It("should use values from parts of tag", func() {
					if syslogDeployment, err = e2e.DeploySyslogReceiver(testDir, corev1.ProtocolTCP, true, helpers.RFC5424); err != nil {
						Fail(fmt.Sprintf("Unable to deploy syslog receiver: %v", err))
					}
					forwarder.Spec.Outputs[0].URL = fmt.Sprintf("tcp://%s.%s.svc:24224", syslogDeployment.ObjectMeta.Name, syslogDeployment.Namespace)
					forwarder.Spec.Outputs[0].Secret = &logging.OutputSecretSpec{
						Name: syslogDeployment.ObjectMeta.Name,
					}
					forwarder.Spec.Outputs[0].Syslog.RFC = helpers.RFC5424.String()
					forwarder.Spec.Outputs[0].Syslog.AppName = "${tag[0]}#${tag[-2]}"
					if err := e2e.CreateClusterLogForwarder(forwarder); err != nil {
						Fail(fmt.Sprintf("Unable to create an instance of logforwarder: %v", err))
					}
					components := []helpers.LogComponentType{helpers.ComponentTypeCollector}
					for _, component := range components {
						if err := e2e.WaitFor(component); err != nil {
							Fail(fmt.Sprintf("Failed waiting for component %s to be ready: %v", component, err))
						}
					}
					logStore := e2e.LogStores[syslogDeployment.GetName()]
					Expect(logStore.HasInfraStructureLogs(helpers.DefaultWaitForLogsTimeout)).To(BeTrue(), "Expected to find stored infrastructure logs")
					_, _ = logStore.GrepLogs(waitlogs, helpers.DefaultWaitForLogsTimeout)
					// typical value of tag in this test case is: kubernetes.var.log.containers.log-generator-746f659fdf-qgclg_clo-test-5764_log-generator-fef2a7848f9741bc6aeb1325aac051c0734c5dc177839c6787da207dc95530ad.log
					// prefix expected is: kubernetes#
					expectedAppNamePrefix := "kubernetes#"
					Expect(logStore.GrepLogs(grepappname, helpers.DefaultWaitForLogsTimeout)).To(HavePrefix(expectedAppNamePrefix))
					expectedMsgID := forwarder.Spec.Outputs[0].Syslog.MsgID
					Expect(logStore.GrepLogs(grepmsgid, helpers.DefaultWaitForLogsTimeout)).To(Equal(expectedMsgID), "Expected: "+expectedMsgID)
					expectedProcID := forwarder.Spec.Outputs[0].Syslog.ProcID
					Expect(logStore.GrepLogs(grepprocid, helpers.DefaultWaitForLogsTimeout)).To(Equal(expectedProcID), "Expected: "+expectedProcID)
				})
			})
			Describe("values for facility,severity", func() {
				BeforeEach(func() {
					generatorPayload["facility_key"] = "local0"
					generatorPayload["severity_key"] = "Informational"
				})
				It("should take from record", func() {
					if syslogDeployment, err = e2e.DeploySyslogReceiver(testDir, corev1.ProtocolTCP, true, helpers.RFC5424); err != nil {
						Fail(fmt.Sprintf("Unable to deploy syslog receiver: %v", err))
					}
					forwarder.Spec.Outputs[0].URL = fmt.Sprintf("tcp://%s.%s.svc:24224", syslogDeployment.ObjectMeta.Name, syslogDeployment.Namespace)
					forwarder.Spec.Outputs[0].Secret = &logging.OutputSecretSpec{
						Name: syslogDeployment.ObjectMeta.Name,
					}
					forwarder.Spec.Outputs[0].Syslog.RFC = helpers.RFC5424.String()
					forwarder.Spec.Outputs[0].Syslog.Facility = "$.message.facility_key"
					forwarder.Spec.Outputs[0].Syslog.Severity = "$.message.severity_key"
					if err := e2e.CreateClusterLogForwarder(forwarder); err != nil {
						Fail(fmt.Sprintf("Unable to create an instance of logforwarder: %v", err))
					}
					components := []helpers.LogComponentType{helpers.ComponentTypeCollector}
					for _, component := range components {
						if err := e2e.WaitFor(component); err != nil {
							Fail(fmt.Sprintf("Failed waiting for component %s to be ready: %v", component, err))
						}
					}
					logStore := e2e.LogStores[syslogDeployment.GetName()]
					Expect(logStore.HasInfraStructureLogs(helpers.DefaultWaitForLogsTimeout)).To(BeTrue(), "Expected to find stored infrastructure logs")
					_, _ = logStore.GrepLogs(waitlogs, helpers.DefaultWaitForLogsTimeout)
					// 134 = Facility(local0/16)*8 + Severity(Informational/6)
					// The 1 after <134> is version, which is always set to 1
					expectedPriority := "<134>1"
					grepPri := fmt.Sprintf(rsyslogFormatStr, logGenPod, "$1")
					Expect(logStore.GrepLogs(grepPri, helpers.DefaultWaitForLogsTimeout)).To(Equal(expectedPriority), "Expected: "+expectedPriority)
				})
			})
			Describe("syslog payload", func() {
				BeforeEach(func() {
					generatorPayload["appname_key"] = "rec_appname"
				})
				It("should take from payload_key", func() {
					if syslogDeployment, err = e2e.DeploySyslogReceiver(testDir, corev1.ProtocolTCP, true, helpers.RFC5424); err != nil {
						Fail(fmt.Sprintf("Unable to deploy syslog receiver: %v", err))
					}
					forwarder.Spec.Outputs[0].URL = fmt.Sprintf("tcp://%s.%s.svc:24224", syslogDeployment.ObjectMeta.Name, syslogDeployment.Namespace)
					forwarder.Spec.Outputs[0].Secret = &logging.OutputSecretSpec{
						Name: syslogDeployment.ObjectMeta.Name,
					}
					forwarder.Spec.Outputs[0].Syslog.RFC = helpers.RFC5424.String()
					forwarder.Spec.Outputs[0].Syslog.AppName = "$.message.appname_key"
					forwarder.Spec.Outputs[0].Syslog.PayloadKey = "message"
					if err := e2e.CreateClusterLogForwarder(forwarder); err != nil {
						Fail(fmt.Sprintf("Unable to create an instance of logforwarder: %v", err))
					}
					components := []helpers.LogComponentType{helpers.ComponentTypeCollector}
					for _, component := range components {
						if err := e2e.WaitFor(component); err != nil {
							Fail(fmt.Sprintf("Failed waiting for component %s to be ready: %v", component, err))
						}
					}
					logStore := e2e.LogStores[syslogDeployment.GetName()]
					Expect(logStore.HasInfraStructureLogs(helpers.DefaultWaitForLogsTimeout)).To(BeTrue(), "Expected to find stored infrastructure logs")
					_, _ = logStore.GrepLogs(waitlogs, helpers.DefaultWaitForLogsTimeout)
					grepMsgContent := fmt.Sprintf(`grep %s %%s | tail -n 1 | awk -F' ' '{ s = ""; for (i = 8; i <= NF; i++) s = s $i " "; print s }'`, "rec_appname")
					str, err := logStore.GrepLogs(grepMsgContent, helpers.DefaultWaitForLogsTimeout)
					Expect(err).To(BeNil())
					str = strings.ReplaceAll(str, "=>", ":")
					msg := map[string]interface{}{}
					err = json.Unmarshal([]byte(str), &msg)
					Expect(err).To(BeNil())
					Expect(msg["msgcontent"]).To(Equal("My life is my message"))
				})
			})
			Describe("large syslog payload", func() {
				BeforeEach(func() {
					str := strings.ReplaceAll(largeStackTrace, "\n", "    ")
					generatorPayload["msgcontent"] = str
				})
				It("should be able to send", func() {
					if syslogDeployment, err = e2e.DeploySyslogReceiver(testDir, corev1.ProtocolTCP, true, helpers.RFC5424); err != nil {
						Fail(fmt.Sprintf("Unable to deploy syslog receiver: %v", err))
					}
					forwarder.Spec.Outputs[0].URL = fmt.Sprintf("tcp://%s.%s.svc:24224", syslogDeployment.ObjectMeta.Name, syslogDeployment.Namespace)
					forwarder.Spec.Outputs[0].Secret = &logging.OutputSecretSpec{
						Name: syslogDeployment.ObjectMeta.Name,
					}
					forwarder.Spec.Outputs[0].Syslog.RFC = helpers.RFC5424.String()
					if err := e2e.CreateClusterLogForwarder(forwarder); err != nil {
						Fail(fmt.Sprintf("Unable to create an instance of logforwarder: %v", err))
					}
					components := []helpers.LogComponentType{helpers.ComponentTypeCollector}
					for _, component := range components {
						if err := e2e.WaitFor(component); err != nil {
							Fail(fmt.Sprintf("Failed waiting for component %s to be ready: %v", component, err))
						}
					}
					logStore := e2e.LogStores[syslogDeployment.GetName()]
					Expect(logStore.HasInfraStructureLogs(helpers.DefaultWaitForLogsTimeout)).To(BeTrue(), "Expected to find stored infrastructure logs")
					_, _ = logStore.GrepLogs(waitlogs, helpers.DefaultWaitForLogsTimeout)
					expectedAppName := forwarder.Spec.Outputs[0].Syslog.AppName
					Expect(logStore.GrepLogs(grepappname, helpers.DefaultWaitForLogsTimeout)).To(Equal(expectedAppName), "Expected: "+expectedAppName)
					expectedMsgID := forwarder.Spec.Outputs[0].Syslog.MsgID
					Expect(logStore.GrepLogs(grepmsgid, helpers.DefaultWaitForLogsTimeout)).To(Equal(expectedMsgID), "Expected: "+expectedMsgID)
					expectedProcID := forwarder.Spec.Outputs[0].Syslog.ProcID
					Expect(logStore.GrepLogs(grepprocid, helpers.DefaultWaitForLogsTimeout)).To(Equal(expectedProcID), "Expected: "+expectedProcID)
				})
			})
		})
		Describe("with rfc3164", func() {
			BeforeEach(func() {
				forwarder = &logging.ClusterLogForwarder{
					TypeMeta: metav1.TypeMeta{
						Kind:       logging.ClusterLogForwarderKind,
						APIVersion: logging.SchemeGroupVersion.String(),
					},
					ObjectMeta: metav1.ObjectMeta{
						Name: "instance",
					},
					Spec: logging.ClusterLogForwarderSpec{
						Outputs: []logging.OutputSpec{
							{
								Name: "syslogout",
								Type: "syslog",
								OutputTypeSpec: logging.OutputTypeSpec{
									Syslog: &logging.Syslog{
										Facility: "user",
										Severity: "debug",
										RFC:      "RFC3164",
										Tag:      "mytag",
									},
								},
							},
						},
						Pipelines: []logging.PipelineSpec{
							{
								Name:       "test-app",
								OutputRefs: []string{"syslogout"},
								InputRefs:  []string{"application"},
							},
						},
					},
				}
			})
			DescribeTable("should be able to send logs to syslog receiver", func(useOldPlugin bool, tls bool, protocol corev1.Protocol) {
				if syslogDeployment, err = e2e.DeploySyslogReceiver(testDir, protocol, tls, helpers.RFC3164); err != nil {
					Fail(fmt.Sprintf("Unable to deploy syslog receiver: %v", err))
				}
				if protocol == corev1.ProtocolTCP {
					forwarder.Spec.Outputs[0].URL = fmt.Sprintf("tcp://%s.%s.svc:24224", syslogDeployment.ObjectMeta.Name, syslogDeployment.Namespace)
				} else {
					forwarder.Spec.Outputs[0].URL = fmt.Sprintf("udp://%s.%s.svc:24224", syslogDeployment.ObjectMeta.Name, syslogDeployment.Namespace)
				}

				if tls {
					forwarder.Spec.Outputs[0].Secret = &logging.OutputSecretSpec{
						Name: syslogDeployment.ObjectMeta.Name,
					}
				}
				if useOldPlugin {
					forwarder.ObjectMeta.Annotations = map[string]string{
						k8shandler.UseOldRemoteSyslogPlugin: "enabled",
					}
				}
				if err := e2e.CreateClusterLogForwarder(forwarder); err != nil {
					Fail(fmt.Sprintf("Unable to create an instance of logforwarder: %v", err))
				}
				components := []helpers.LogComponentType{helpers.ComponentTypeCollector}
				for _, component := range components {
					if err := e2e.WaitFor(component); err != nil {
						Fail(fmt.Sprintf("Failed waiting for component %s to be ready: %v", component, err))
					}
				}
				logStore := e2e.LogStores[syslogDeployment.GetName()]
				Expect(logStore.HasInfraStructureLogs(helpers.DefaultWaitForLogsTimeout)).To(BeTrue(), "Expected to find stored infrastructure logs")
				if !useOldPlugin {
					_, _ = logStore.GrepLogs(waitlogs, helpers.DefaultWaitForLogsTimeout)
					expectedAppName := forwarder.Spec.Outputs[0].Syslog.Tag
					Expect(logStore.GrepLogs(grepappname, helpers.DefaultWaitForLogsTimeout)).To(Equal(expectedAppName), "Expected: "+expectedAppName)
				}
			},
				// old syslog plugin does not support TLS, so set false for tls
				Entry("with old syslog plugin, with TCP", true, false, corev1.ProtocolTCP),
				Entry("with old syslog plugin, with UDP", true, false, corev1.ProtocolUDP),
				// using new plugin
				Entry("with new syslog plugin, without TLS, with TCP", false, false, corev1.ProtocolTCP),
				Entry("with new syslog plugin, without TLS, with UDP", false, false, corev1.ProtocolTCP),
				Entry("with new syslog plugin, with TLS, with TCP", false, true, corev1.ProtocolTCP),
				Entry("with new syslog plugin, with TLS, with UDP", false, true, corev1.ProtocolTCP),
			)
			Describe("values for tag", func() {
				BeforeEach(func() {
					generatorPayload["tag_key"] = "rec_tag"
				})
				It("should use values from record", func() {
					if syslogDeployment, err = e2e.DeploySyslogReceiver(testDir, corev1.ProtocolTCP, true, helpers.RFC3164); err != nil {
						Fail(fmt.Sprintf("Unable to deploy syslog receiver: %v", err))
					}
					forwarder.Spec.Outputs[0].URL = fmt.Sprintf("tcp://%s.%s.svc:24224", syslogDeployment.ObjectMeta.Name, syslogDeployment.Namespace)
					forwarder.Spec.Outputs[0].Secret = &logging.OutputSecretSpec{
						Name: syslogDeployment.ObjectMeta.Name,
					}
					forwarder.Spec.Outputs[0].Syslog.RFC = helpers.RFC3164.String()
					forwarder.Spec.Outputs[0].Syslog.Tag = "$.message.tag_key"
					if err := e2e.CreateClusterLogForwarder(forwarder); err != nil {
						Fail(fmt.Sprintf("Unable to create an instance of logforwarder: %v", err))
					}
					components := []helpers.LogComponentType{helpers.ComponentTypeCollector}
					for _, component := range components {
						if err := e2e.WaitFor(component); err != nil {
							Fail(fmt.Sprintf("Failed waiting for component %s to be ready: %v", component, err))
						}
					}
					logStore := e2e.LogStores[syslogDeployment.GetName()]
					Expect(logStore.HasInfraStructureLogs(helpers.DefaultWaitForLogsTimeout)).To(BeTrue(), "Expected to find stored infrastructure logs")
					_, _ = logStore.GrepLogs(waitlogs, helpers.DefaultWaitForLogsTimeout)
					expectedTag := generatorPayload["tag_key"]
					Expect(logStore.GrepLogs(greptag, helpers.DefaultWaitForLogsTimeout)).To(Equal(expectedTag), "Expected: "+expectedTag)
				})
				It("should take value from complete fluentd tag", func() {
					if syslogDeployment, err = e2e.DeploySyslogReceiver(testDir, corev1.ProtocolTCP, true, helpers.RFC3164); err != nil {
						Fail(fmt.Sprintf("Unable to deploy syslog receiver: %v", err))
					}
					forwarder.Spec.Outputs[0].URL = fmt.Sprintf("tcp://%s.%s.svc:24224", syslogDeployment.ObjectMeta.Name, syslogDeployment.Namespace)
					forwarder.Spec.Outputs[0].Secret = &logging.OutputSecretSpec{
						Name: syslogDeployment.ObjectMeta.Name,
					}
					forwarder.Spec.Outputs[0].Syslog.RFC = helpers.RFC3164.String()
					forwarder.Spec.Outputs[0].Syslog.Tag = "tag"
					if err := e2e.CreateClusterLogForwarder(forwarder); err != nil {
						Fail(fmt.Sprintf("Unable to create an instance of logforwarder: %v", err))
					}
					components := []helpers.LogComponentType{helpers.ComponentTypeCollector}
					for _, component := range components {
						if err := e2e.WaitFor(component); err != nil {
							Fail(fmt.Sprintf("Failed waiting for component %s to be ready: %v", component, err))
						}
					}
					logStore := e2e.LogStores[syslogDeployment.GetName()]
					Expect(logStore.HasInfraStructureLogs(helpers.DefaultWaitForLogsTimeout)).To(BeTrue(), "Expected to find stored infrastructure logs")
					_, _ = logStore.GrepLogs(waitlogs, helpers.DefaultWaitForLogsTimeout)
					// typical value of tag in this test case is: kubernetes.var.log.containers.log-generator-746f659fdf-qgclg_clo-test-5764_log-generator-fef2a7848f9741bc6aeb1325aac051c0734c5dc177839c6787da207dc95530ad.log
					expectedTagPrefix := "kubernetes.var.log.containers"
					Expect(logStore.GrepLogs(greptag, helpers.DefaultWaitForLogsTimeout)).To(HavePrefix(expectedTagPrefix))
				})
				It("should use values from parts of tag", func() {
					if syslogDeployment, err = e2e.DeploySyslogReceiver(testDir, corev1.ProtocolTCP, true, helpers.RFC3164); err != nil {
						Fail(fmt.Sprintf("Unable to deploy syslog receiver: %v", err))
					}
					forwarder.Spec.Outputs[0].URL = fmt.Sprintf("tcp://%s.%s.svc:24224", syslogDeployment.ObjectMeta.Name, syslogDeployment.Namespace)
					forwarder.Spec.Outputs[0].Secret = &logging.OutputSecretSpec{
						Name: syslogDeployment.ObjectMeta.Name,
					}
					forwarder.Spec.Outputs[0].Syslog.RFC = helpers.RFC3164.String()
					forwarder.Spec.Outputs[0].Syslog.Tag = "${tag[0]}#${tag[-2]}"
					if err := e2e.CreateClusterLogForwarder(forwarder); err != nil {
						Fail(fmt.Sprintf("Unable to create an instance of logforwarder: %v", err))
					}
					components := []helpers.LogComponentType{helpers.ComponentTypeCollector}
					for _, component := range components {
						if err := e2e.WaitFor(component); err != nil {
							Fail(fmt.Sprintf("Failed waiting for component %s to be ready: %v", component, err))
						}
					}
					logStore := e2e.LogStores[syslogDeployment.GetName()]
					Expect(logStore.HasInfraStructureLogs(helpers.DefaultWaitForLogsTimeout)).To(BeTrue(), "Expected to find stored infrastructure logs")
					_, _ = logStore.GrepLogs(waitlogs, helpers.DefaultWaitForLogsTimeout)
					// typical value of tag in this test case is: kubernetes.var.log.containers.log-generator-746f659fdf-qgclg_clo-test-5764_log-generator-fef2a7848f9741bc6aeb1325aac051c0734c5dc177839c6787da207dc95530ad.log
					// prefix expected is: kubernetes#
					expectedTagPrefix := "kubernetes#"
					Expect(logStore.GrepLogs(greptag, helpers.DefaultWaitForLogsTimeout)).To(HavePrefix(expectedTagPrefix))
				})
			})
			Describe("values for facility,severity", func() {
				BeforeEach(func() {
					generatorPayload["facility_key"] = "local0"
					generatorPayload["severity_key"] = "Informational"
				})
				It("should take from record", func() {
					if syslogDeployment, err = e2e.DeploySyslogReceiver(testDir, corev1.ProtocolTCP, true, helpers.RFC3164); err != nil {
						Fail(fmt.Sprintf("Unable to deploy syslog receiver: %v", err))
					}
					forwarder.Spec.Outputs[0].URL = fmt.Sprintf("tcp://%s.%s.svc:24224", syslogDeployment.ObjectMeta.Name, syslogDeployment.Namespace)
					forwarder.Spec.Outputs[0].Secret = &logging.OutputSecretSpec{
						Name: syslogDeployment.ObjectMeta.Name,
					}
					forwarder.Spec.Outputs[0].Syslog.RFC = helpers.RFC3164.String()
					forwarder.Spec.Outputs[0].Syslog.Facility = "$.message.facility_key"
					forwarder.Spec.Outputs[0].Syslog.Severity = "$.message.severity_key"
					if err := e2e.CreateClusterLogForwarder(forwarder); err != nil {
						Fail(fmt.Sprintf("Unable to create an instance of logforwarder: %v", err))
					}
					components := []helpers.LogComponentType{helpers.ComponentTypeCollector}
					for _, component := range components {
						if err := e2e.WaitFor(component); err != nil {
							Fail(fmt.Sprintf("Failed waiting for component %s to be ready: %v", component, err))
						}
					}
					logStore := e2e.LogStores[syslogDeployment.GetName()]
					Expect(logStore.HasInfraStructureLogs(helpers.DefaultWaitForLogsTimeout)).To(BeTrue(), "Expected to find stored infrastructure logs")
					_, _ = logStore.GrepLogs(waitlogs, helpers.DefaultWaitForLogsTimeout)
					// 134 = Facility(local0/16)*8 + Severity(Informational/6)
					// The 1 after <134> is version, which is always set to 1
					expectedPriority := "<134>1"
					grepPri := fmt.Sprintf(rsyslogFormatStr, logGenPod, "$1")
					Expect(logStore.GrepLogs(grepPri, helpers.DefaultWaitForLogsTimeout)).To(Equal(expectedPriority), "Expected: "+expectedPriority)
				})
			})
			Describe("syslog payload", func() {
				Context("with new plugin", func() {
					BeforeEach(func() {
						generatorPayload["tag_key"] = "rec_tag"
					})
					It("should take from payload_key", func() {
						if syslogDeployment, err = e2e.DeploySyslogReceiver(testDir, corev1.ProtocolTCP, false, helpers.RFC3164); err != nil {
							Fail(fmt.Sprintf("Unable to deploy syslog receiver: %v", err))
						}
						forwarder.Spec.Outputs[0].URL = fmt.Sprintf("tls://%s.%s.svc:24224", syslogDeployment.ObjectMeta.Name, syslogDeployment.Namespace)
						forwarder.Spec.Outputs[0].Syslog.RFC = helpers.RFC3164.String()
						forwarder.Spec.Outputs[0].Syslog.Tag = "$.message.tag_key"
						forwarder.Spec.Outputs[0].Syslog.PayloadKey = "message"
						if err := e2e.CreateClusterLogForwarder(forwarder); err != nil {
							Fail(fmt.Sprintf("Unable to create an instance of logforwarder: %v", err))
						}
						components := []helpers.LogComponentType{helpers.ComponentTypeCollector}
						for _, component := range components {
							if err := e2e.WaitFor(component); err != nil {
								Fail(fmt.Sprintf("Failed waiting for component %s to be ready: %v", component, err))
							}
						}
						logStore := e2e.LogStores[syslogDeployment.GetName()]
						Expect(logStore.HasInfraStructureLogs(helpers.DefaultWaitForLogsTimeout)).To(BeTrue(), "Expected to find stored infrastructure logs")
						_, _ = logStore.GrepLogs(waitlogs, helpers.DefaultWaitForLogsTimeout)
						grepMsgContent := fmt.Sprintf(`grep %s %%s | tail -n 1 | awk -F' ' '{ s = ""; for (i = 8; i <= NF; i++) s = s $i " "; print s }'`, "rec_tag")
						str, err := logStore.GrepLogs(grepMsgContent, helpers.DefaultWaitForLogsTimeout)
						Expect(err).To(BeNil())
						str = strings.ReplaceAll(str, "=>", ":")
						msg := map[string]interface{}{}
						err = json.Unmarshal([]byte(str), &msg)
						Expect(err).To(BeNil())
						Expect(msg["msgcontent"]).To(Equal("My life is my message"))
					})
				})
			})
			Describe("syslog payload", func() {
				Context("with addLogSourceToMessage flag", func() {
					BeforeEach(func() {
						generatorPayload["tag_key"] = "rec_tag"
					})
					It("should add namespace, pod, container name to log message", func() {
						if syslogDeployment, err = e2e.DeploySyslogReceiver(testDir, corev1.ProtocolTCP, false, helpers.RFC3164); err != nil {
							Fail(fmt.Sprintf("Unable to deploy syslog receiver: %v", err))
						}
						forwarder.Spec.Outputs[0].URL = fmt.Sprintf("tls://%s.%s.svc:24224", syslogDeployment.ObjectMeta.Name, syslogDeployment.Namespace)
						forwarder.Spec.Outputs[0].Syslog.RFC = helpers.RFC3164.String()
						forwarder.Spec.Outputs[0].Syslog.PayloadKey = "message"
						forwarder.Spec.Outputs[0].Syslog.AddLogSource = true
						if err := e2e.CreateClusterLogForwarder(forwarder); err != nil {
							Fail(fmt.Sprintf("Unable to create an instance of logforwarder: %v", err))
						}
						components := []helpers.LogComponentType{helpers.ComponentTypeCollector}
						for _, component := range components {
							if err := e2e.WaitFor(component); err != nil {
								Fail(fmt.Sprintf("Failed waiting for component %s to be ready: %v", component, err))
							}
						}
						logStore := e2e.LogStores[syslogDeployment.GetName()]
						Expect(logStore.HasInfraStructureLogs(helpers.DefaultWaitForLogsTimeout)).To(BeTrue(), "Expected to find stored infrastructure logs")
						_, _ = logStore.GrepLogs(waitlogs, helpers.DefaultWaitForLogsTimeout)
						grepMsgContent := fmt.Sprintf(`grep %s %%s | tail -n 1 | awk -F' ' '{ s = ""; for (i = 8; i <= NF; i++) s = s $i " "; print s }'`, "namespace_name")
						_, err := logStore.GrepLogs(grepMsgContent, helpers.DefaultWaitForLogsTimeout)
						Expect(err).To(BeNil())
					})
				})
				Context("with addLogSourceToMessage flag", func() {
					Context("for audit logs", func() {
						BeforeEach(func() {
							generatorPayload["tag_key"] = "rec_tag"
							forwarder.Spec.Pipelines[0].InputRefs = []string{"audit"}
						})
						It("should send log message successfully", func() {
							if syslogDeployment, err = e2e.DeploySyslogReceiver(testDir, corev1.ProtocolTCP, false, helpers.RFC3164); err != nil {
								Fail(fmt.Sprintf("Unable to deploy syslog receiver: %v", err))
							}
							forwarder.Spec.Outputs[0].URL = fmt.Sprintf("tls://%s.%s.svc:24224", syslogDeployment.ObjectMeta.Name, syslogDeployment.Namespace)
							forwarder.Spec.Outputs[0].Syslog.RFC = helpers.RFC3164.String()
							forwarder.Spec.Outputs[0].Syslog.PayloadKey = "message"
							forwarder.Spec.Outputs[0].Syslog.AddLogSource = true
							if err := e2e.CreateClusterLogForwarder(forwarder); err != nil {
								Fail(fmt.Sprintf("Unable to create an instance of logforwarder: %v", err))
							}
							components := []helpers.LogComponentType{helpers.ComponentTypeCollector}
							for _, component := range components {
								if err := e2e.WaitFor(component); err != nil {
									Fail(fmt.Sprintf("Failed waiting for component %s to be ready: %v", component, err))
								}
							}
							logStore := e2e.LogStores[syslogDeployment.GetName()]
							Expect(logStore.HasInfraStructureLogs(helpers.DefaultWaitForLogsTimeout)).To(BeTrue(), "Expected to find stored infrastructure logs")
							_, _ = logStore.GrepLogs(waitlogs, helpers.DefaultWaitForLogsTimeout)
							grepMsgContent := fmt.Sprintf(`grep %s %%s | tail -n 1 | awk -F' ' '{ s = ""; for (i = 8; i <= NF; i++) s = s $i " "; print s }'`, "rec_tag")
							_, err := logStore.GrepLogs(grepMsgContent, helpers.DefaultWaitForLogsTimeout)
							Expect(err).To(BeNil())
						})
					})
				})
			})
		})
		AfterEach(func() {
			e2e.Cleanup()
			e2e.WaitForCleanupCompletion(logGenNS, []string{"test"})
			e2e.WaitForCleanupCompletion(helpers.OpenshiftLoggingNS, []string{"fluentd", "syslog-receiver"})
			generatorPayload = map[string]string{}
		})
	})
})

const (
	largeStackTrace = `java.lang.RuntimeException: Error grabbing Grapes -- [unresolved dependency: org.jenkins-ci#trilead-ssh2;build-217-jenkins-14: not found, unresolved dependency: org.kohsuke.stapler#stapler-groovy;1.256: not found, unresolved dependency: org.kohsuke.stapler#stapler-jrebel;1.256: not found, unresolved dependency: io.jenkins.stapler#jenkins-stapler-support;1.0: not found, unresolved dependency: commons-httpclient#commons-httpclient;3.1-jenkins-1: not found, unresolved dependency: org.jenkins-ci#bytecode-compatibility-transformer;2.0-beta-2: not found, unresolved dependency: org.jenkins-ci#task-reactor;1.5: not found, download failed: org.jenkins-ci.main#jenkins-core;2.164!jenkins-core.jar, download failed: org.jenkins-ci.plugins.icon-shim#icon-set;1.0.5!icon-set.jar, download failed: org.jenkins-ci.main#remoting;3.29!remoting.jar, download failed: org.jenkins-ci#constant-pool-scanner;1.2!constant-pool-scanner.jar, download failed: org.jenkins-ci.main#cli;2.164!cli.jar, download failed: org.kohsuke#access-modifier-annotation;1.14!access-modifier-annotation.jar, download failed: org.jenkins-ci#annotation-indexer;1.12!annotation-indexer.jar, download failed: org.jvnet.localizer#localizer;1.24!localizer.jar, download failed: net.i2p.crypto#eddsa;0.3.0!eddsa.jar(bundle), download failed: org.jenkins-ci#version-number;1.6!version-number.jar, download failed: com.google.code.findbugs#annotations;3.0.0!annotations.jar, download failed: org.jenkins-ci#crypto-util;1.1!crypto-util.jar, download failed: org.jvnet.hudson#jtidy;4aug2000r7-dev-hudson-1!jtidy.jar, download failed: com.google.inject#guice;4.0!guice.jar, download failed: org.jruby.ext.posix#jna-posix;1.0.3-jenkins-1!jna-posix.jar, download failed: com.github.jnr#jnr-posix;3.0.45!jnr-posix.jar, download failed: com.github.jnr#jnr-ffi;2.1.8!jnr-ffi.jar, download failed: org.slf4j#jcl-over-slf4j;1.7.25!jcl-over-slf4j.jar, download failed: org.slf4j#log4j-over-slf4j;1.7.25!log4j-over-slf4j.jar, download failed: javax.xml.stream#stax-api;1.0-2!stax-api.jar]
	at sun.reflect.NativeConstructorAccessorImpl.newInstance0(Native Method)
	at sun.reflect.NativeConstructorAccessorImpl.newInstance(NativeConstructorAccessorImpl.java:62)
	at sun.reflect.DelegatingConstructorAccessorImpl.newInstance(DelegatingConstructorAccessorImpl.java:45)
	at java.lang.reflect.Constructor.newInstance(Constructor.java:423)
	at org.codehaus.groovy.reflection.CachedConstructor.invoke(CachedConstructor.java:83)
	at org.codehaus.groovy.reflection.CachedConstructor.doConstructorInvoke(CachedConstructor.java:77)
	at org.codehaus.groovy.runtime.callsite.ConstructorSite$ConstructorSiteNoUnwrap.callConstructor(ConstructorSite.java:84)
	at org.codehaus.groovy.runtime.callsite.CallSiteArray.defaultCallConstructor(CallSiteArray.java:60)
	at org.codehaus.groovy.runtime.callsite.AbstractCallSite.callConstructor(AbstractCallSite.java:235)
	at org.codehaus.groovy.runtime.callsite.AbstractCallSite.callConstructor(AbstractCallSite.java:247)
	at groovy.grape.GrapeIvy.getDependencies(GrapeIvy.groovy:424)
	at sun.reflect.NativeMethodAccessorImpl.invoke0(Native Method)
	at sun.reflect.NativeMethodAccessorImpl.invoke(NativeMethodAccessorImpl.java:62)
	at sun.reflect.DelegatingMethodAccessorImpl.invoke(DelegatingMethodAccessorImpl.java:43)
	at java.lang.reflect.Method.invoke(Method.java:498)
	at org.codehaus.groovy.runtime.callsite.PogoMetaMethodSite$PogoCachedMethodSite.invoke(PogoMetaMethodSite.java:169)
	at org.codehaus.groovy.runtime.callsite.PogoMetaMethodSite.callCurrent(PogoMetaMethodSite.java:59)
	at org.codehaus.groovy.runtime.callsite.CallSiteArray.defaultCallCurrent(CallSiteArray.java:52)
	at org.codehaus.groovy.runtime.callsite.AbstractCallSite.callCurrent(AbstractCallSite.java:154)
	at groovy.grape.GrapeIvy.resolve(GrapeIvy.groovy:571)
	at groovy.grape.GrapeIvy$resolve$1.callCurrent(Unknown Source)
	at org.codehaus.groovy.runtime.callsite.CallSiteArray.defaultCallCurrent(CallSiteArray.java:52)
	at org.codehaus.groovy.runtime.callsite.AbstractCallSite.callCurrent(AbstractCallSite.java:154)
	at org.codehaus.groovy.runtime.callsite.AbstractCallSite.callCurrent(AbstractCallSite.java:190)
	at groovy.grape.GrapeIvy.resolve(GrapeIvy.groovy:538)
	at groovy.grape.GrapeIvy$resolve$0.callCurrent(Unknown Source)
	at org.codehaus.groovy.runtime.callsite.CallSiteArray.defaultCallCurrent(CallSiteArray.java:52)
	at org.codehaus.groovy.runtime.callsite.AbstractCallSite.callCurrent(AbstractCallSite.java:154)
	at org.codehaus.groovy.runtime.callsite.AbstractCallSite.callCurrent(AbstractCallSite.java:182)
	at groovy.grape.GrapeIvy.grab(GrapeIvy.groovy:256)
	at groovy.grape.Grape.grab(Grape.java:167)
	at groovy.grape.GrabAnnotationTransformation.visit(GrabAnnotationTransformation.java:378)
	at org.codehaus.groovy.transform.ASTTransformationVisitor$3.call(ASTTransformationVisitor.java:321)
	at org.codehaus.groovy.control.CompilationUnit.applyToSourceUnits(CompilationUnit.java:943)
	at org.codehaus.groovy.control.CompilationUnit.doPhaseOperation(CompilationUnit.java:605)
	at org.codehaus.groovy.control.CompilationUnit.processPhaseOperations(CompilationUnit.java:581)
	at org.codehaus.groovy.control.CompilationUnit.compile(CompilationUnit.java:558)
	at groovy.lang.GroovyClassLoader.doParseClass(GroovyClassLoader.java:298)
	at groovy.lang.GroovyClassLoader.parseClass(GroovyClassLoader.java:268)
	at groovy.lang.GroovyShell.parseClass(GroovyShell.java:688)
	at groovy.lang.GroovyShell.run(GroovyShell.java:517)
	at groovy.lang.GroovyShell.run(GroovyShell.java:507)
	at groovy.ui.GroovyMain.processOnce(GroovyMain.java:653)
	at groovy.ui.GroovyMain.run(GroovyMain.java:384)
	at groovy.ui.GroovyMain.process(GroovyMain.java:370)
	at groovy.ui.GroovyMain.processArgs(GroovyMain.java:129)
	at groovy.ui.GroovyMain.main(GroovyMain.java:109)
	at sun.reflect.NativeMethodAccessorImpl.invoke0(Native Method)
	at sun.reflect.NativeMethodAccessorImpl.invoke(NativeMethodAccessorImpl.java:62)
	at sun.reflect.DelegatingMethodAccessorImpl.invoke(DelegatingMethodAccessorImpl.java:43)
	at java.lang.reflect.Method.invoke(Method.java:498)
	at org.codehaus.groovy.tools.GroovyStarter.rootLoader(GroovyStarter.java:109)
	at org.codehaus.groovy.tools.GroovyStarter.main(GroovyStarter.java:131)
	`
)
