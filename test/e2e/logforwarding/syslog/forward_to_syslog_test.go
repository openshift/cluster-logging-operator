package fluent

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

	logging "github.com/openshift/cluster-logging-operator/pkg/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/pkg/apis/logging/v1/outputs"
	"github.com/openshift/cluster-logging-operator/pkg/k8shandler"
	"github.com/openshift/cluster-logging-operator/pkg/logger"
	"github.com/openshift/cluster-logging-operator/test/helpers"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	// grep <generator-pod-name> <syslog-log-filename>
	rsyslogFormatStr = `grep %s %%s | tail -n 1 | awk -F' ' '{print %s}'`
)

var _ = Describe("LogForwarder", func() {
	_, filename, _, _ := runtime.Caller(0)
	logger.Infof("Running %s", filename)
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
			logger.Errorf("unable to deploy log generator. E: %s", err.Error())
		}
		logger.Infof("log generator pod: %s", logGenPod)
		testDir = filepath.Dir(filename)
		// wait for current log-generator's logs to appear in syslog
		waitlogs = fmt.Sprintf(`[ $(grep %s %%s | wc -l) -gt 0 ]`, logGenPod)
		grepappname = fmt.Sprintf(rsyslogFormatStr, logGenPod, "$4")
		greptag = fmt.Sprintf(rsyslogFormatStr, logGenPod, "$4")
		grepprocid = fmt.Sprintf(rsyslogFormatStr, logGenPod, "$5")
		grepmsgid = fmt.Sprintf(rsyslogFormatStr, logGenPod, "$6")
	})
	Describe("for Syslog", func() {
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
									Syslog: &outputs.Syslog{
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
					forwarder.Spec.Outputs[0].URL = fmt.Sprintf("%s.%s.svc:24224", syslogDeployment.ObjectMeta.Name, syslogDeployment.Namespace)
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
				Expect(e2e.LogStore.HasInfraStructureLogs(helpers.DefaultWaitForLogsTimeout)).To(BeTrue(), "Expected to find stored infrastructure logs")
				_, _ = e2e.LogStore.GrepLogs(waitlogs, helpers.DefaultWaitForLogsTimeout)
				expectedAppName := forwarder.Spec.Outputs[0].Syslog.AppName
				Expect(e2e.LogStore.GrepLogs(grepappname, helpers.DefaultWaitForLogsTimeout)).To(Equal(expectedAppName), "Expected: "+expectedAppName)
				expectedMsgID := forwarder.Spec.Outputs[0].Syslog.MsgID
				Expect(e2e.LogStore.GrepLogs(grepmsgid, helpers.DefaultWaitForLogsTimeout)).To(Equal(expectedMsgID), "Expected: "+expectedMsgID)
				expectedProcID := forwarder.Spec.Outputs[0].Syslog.ProcID
				Expect(e2e.LogStore.GrepLogs(grepprocid, helpers.DefaultWaitForLogsTimeout)).To(Equal(expectedProcID), "Expected: "+expectedProcID)
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
						forwarder.Spec.Outputs[0].URL = fmt.Sprintf("%s.%s.svc:24224", syslogDeployment.ObjectMeta.Name, syslogDeployment.Namespace)
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
						Expect(e2e.LogStore.HasInfraStructureLogs(helpers.DefaultWaitForLogsTimeout)).To(BeTrue(), "Expected to find stored infrastructure logs")
						_, _ = e2e.LogStore.GrepLogs(waitlogs, helpers.DefaultWaitForLogsTimeout)
						expectedAppName := generatorPayload["appname_key"]
						Expect(e2e.LogStore.GrepLogs(grepappname, helpers.DefaultWaitForLogsTimeout)).To(Equal(expectedAppName), "Expected: "+expectedAppName)
						expectedMsgID := generatorPayload["msgid_key"]
						Expect(e2e.LogStore.GrepLogs(grepmsgid, helpers.DefaultWaitForLogsTimeout)).To(Equal(expectedMsgID), "Expected: "+expectedMsgID)
						expectedProcID := generatorPayload["procid_key"]
						Expect(e2e.LogStore.GrepLogs(grepprocid, helpers.DefaultWaitForLogsTimeout)).To(Equal(expectedProcID), "Expected: "+expectedProcID)
					})
				})
				It("should take value from complete fluentd tag", func() {
					if syslogDeployment, err = e2e.DeploySyslogReceiver(testDir, corev1.ProtocolTCP, true, helpers.RFC5424); err != nil {
						Fail(fmt.Sprintf("Unable to deploy syslog receiver: %v", err))
					}
					forwarder.Spec.Outputs[0].URL = fmt.Sprintf("%s.%s.svc:24224", syslogDeployment.ObjectMeta.Name, syslogDeployment.Namespace)
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
					Expect(e2e.LogStore.HasInfraStructureLogs(helpers.DefaultWaitForLogsTimeout)).To(BeTrue(), "Expected to find stored infrastructure logs")
					_, _ = e2e.LogStore.GrepLogs(waitlogs, helpers.DefaultWaitForLogsTimeout)
					// typical value of tag in this test case is: kubernetes.var.log.containers.log-generator-746f659fdf-qgclg_clo-test-5764_log-generator-fef2a7848f9741bc6aeb1325aac051c0734c5dc177839c6787da207dc95530ad.log
					expectedAppNamePrefix := "kubernetes.var.log.containers"
					Expect(e2e.LogStore.GrepLogs(grepappname, helpers.DefaultWaitForLogsTimeout)).To(HavePrefix(expectedAppNamePrefix))
					expectedMsgID := forwarder.Spec.Outputs[0].Syslog.MsgID
					Expect(e2e.LogStore.GrepLogs(grepmsgid, helpers.DefaultWaitForLogsTimeout)).To(Equal(expectedMsgID), "Expected: "+expectedMsgID)
					expectedProcID := forwarder.Spec.Outputs[0].Syslog.ProcID
					Expect(e2e.LogStore.GrepLogs(grepprocid, helpers.DefaultWaitForLogsTimeout)).To(Equal(expectedProcID), "Expected: "+expectedProcID)
				})
				It("should use values from parts of tag", func() {
					if syslogDeployment, err = e2e.DeploySyslogReceiver(testDir, corev1.ProtocolTCP, true, helpers.RFC5424); err != nil {
						Fail(fmt.Sprintf("Unable to deploy syslog receiver: %v", err))
					}
					forwarder.Spec.Outputs[0].URL = fmt.Sprintf("%s.%s.svc:24224", syslogDeployment.ObjectMeta.Name, syslogDeployment.Namespace)
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
					Expect(e2e.LogStore.HasInfraStructureLogs(helpers.DefaultWaitForLogsTimeout)).To(BeTrue(), "Expected to find stored infrastructure logs")
					_, _ = e2e.LogStore.GrepLogs(waitlogs, helpers.DefaultWaitForLogsTimeout)
					// typical value of tag in this test case is: kubernetes.var.log.containers.log-generator-746f659fdf-qgclg_clo-test-5764_log-generator-fef2a7848f9741bc6aeb1325aac051c0734c5dc177839c6787da207dc95530ad.log
					// prefix expected is: kubernetes#log-generator-746f659fdf-qgclg_clo-test-5764_log-generator-fef2a7848f9741bc6aeb1325aac051c0734c5dc177839c6787da207dc95530ad.log
					expectedAppNamePrefix := "kubernetes#"
					Expect(e2e.LogStore.GrepLogs(grepappname, helpers.DefaultWaitForLogsTimeout)).To(HavePrefix(expectedAppNamePrefix))
					expectedMsgID := forwarder.Spec.Outputs[0].Syslog.MsgID
					Expect(e2e.LogStore.GrepLogs(grepmsgid, helpers.DefaultWaitForLogsTimeout)).To(Equal(expectedMsgID), "Expected: "+expectedMsgID)
					expectedProcID := forwarder.Spec.Outputs[0].Syslog.ProcID
					Expect(e2e.LogStore.GrepLogs(grepprocid, helpers.DefaultWaitForLogsTimeout)).To(Equal(expectedProcID), "Expected: "+expectedProcID)
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
					forwarder.Spec.Outputs[0].URL = fmt.Sprintf("%s.%s.svc:24224", syslogDeployment.ObjectMeta.Name, syslogDeployment.Namespace)
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
					Expect(e2e.LogStore.HasInfraStructureLogs(helpers.DefaultWaitForLogsTimeout)).To(BeTrue(), "Expected to find stored infrastructure logs")
					_, _ = e2e.LogStore.GrepLogs(waitlogs, helpers.DefaultWaitForLogsTimeout)
					// 134 = Facility(local0/16)*8 + Severity(Informational/6)
					// The 1 after <134> is version, which is always set to 1
					expectedPriority := "<134>1"
					grepPri := fmt.Sprintf(rsyslogFormatStr, logGenPod, "$1")
					Expect(e2e.LogStore.GrepLogs(grepPri, helpers.DefaultWaitForLogsTimeout)).To(Equal(expectedPriority), "Expected: "+expectedPriority)
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
					forwarder.Spec.Outputs[0].URL = fmt.Sprintf("%s.%s.svc:24224", syslogDeployment.ObjectMeta.Name, syslogDeployment.Namespace)
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
					Expect(e2e.LogStore.HasInfraStructureLogs(helpers.DefaultWaitForLogsTimeout)).To(BeTrue(), "Expected to find stored infrastructure logs")
					grepMsgContent := fmt.Sprintf(`grep %s %%s | tail -n 1 | awk -F' ' '{ s = ""; for (i = 8; i <= NF; i++) s = s $i " "; print s }'`, "rec_appname")
					str, err := e2e.LogStore.GrepLogs(grepMsgContent, helpers.DefaultWaitForLogsTimeout)
					Expect(err).To(BeNil())
					str = strings.ReplaceAll(str, "=>", ":")
					msg := map[string]interface{}{}
					err = json.Unmarshal([]byte(str), &msg)
					Expect(err).To(BeNil())
					Expect(msg["msgcontent"]).To(Equal("My life is my message"))
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
									Syslog: &outputs.Syslog{
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
					forwarder.Spec.Outputs[0].URL = fmt.Sprintf("%s.%s.svc:24224", syslogDeployment.ObjectMeta.Name, syslogDeployment.Namespace)
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
				Expect(e2e.LogStore.HasInfraStructureLogs(helpers.DefaultWaitForLogsTimeout)).To(BeTrue(), "Expected to find stored infrastructure logs")
				_, _ = e2e.LogStore.GrepLogs(waitlogs, helpers.DefaultWaitForLogsTimeout)
				if !useOldPlugin {
					expectedAppName := forwarder.Spec.Outputs[0].Syslog.Tag
					Expect(e2e.LogStore.GrepLogs(grepappname, helpers.DefaultWaitForLogsTimeout)).To(Equal(expectedAppName), "Expected: "+expectedAppName)
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
					forwarder.Spec.Outputs[0].URL = fmt.Sprintf("%s.%s.svc:24224", syslogDeployment.ObjectMeta.Name, syslogDeployment.Namespace)
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
					Expect(e2e.LogStore.HasInfraStructureLogs(helpers.DefaultWaitForLogsTimeout)).To(BeTrue(), "Expected to find stored infrastructure logs")
					_, _ = e2e.LogStore.GrepLogs(waitlogs, helpers.DefaultWaitForLogsTimeout)
					expectedTag := generatorPayload["tag_key"]
					Expect(e2e.LogStore.GrepLogs(greptag, helpers.DefaultWaitForLogsTimeout)).To(Equal(expectedTag), "Expected: "+expectedTag)
				})
				It("should take value from complete fluentd tag", func() {
					if syslogDeployment, err = e2e.DeploySyslogReceiver(testDir, corev1.ProtocolTCP, true, helpers.RFC3164); err != nil {
						Fail(fmt.Sprintf("Unable to deploy syslog receiver: %v", err))
					}
					forwarder.Spec.Outputs[0].URL = fmt.Sprintf("%s.%s.svc:24224", syslogDeployment.ObjectMeta.Name, syslogDeployment.Namespace)
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
					Expect(e2e.LogStore.HasInfraStructureLogs(helpers.DefaultWaitForLogsTimeout)).To(BeTrue(), "Expected to find stored infrastructure logs")
					_, _ = e2e.LogStore.GrepLogs(waitlogs, helpers.DefaultWaitForLogsTimeout)
					// typical value of tag in this test case is: kubernetes.var.log.containers.log-generator-746f659fdf-qgclg_clo-test-5764_log-generator-fef2a7848f9741bc6aeb1325aac051c0734c5dc177839c6787da207dc95530ad.log
					expectedTagPrefix := "kubernetes.var.log.containers"
					Expect(e2e.LogStore.GrepLogs(greptag, helpers.DefaultWaitForLogsTimeout)).To(HavePrefix(expectedTagPrefix))
				})
				It("should use values from parts of tag", func() {
					if syslogDeployment, err = e2e.DeploySyslogReceiver(testDir, corev1.ProtocolTCP, true, helpers.RFC3164); err != nil {
						Fail(fmt.Sprintf("Unable to deploy syslog receiver: %v", err))
					}
					forwarder.Spec.Outputs[0].URL = fmt.Sprintf("%s.%s.svc:24224", syslogDeployment.ObjectMeta.Name, syslogDeployment.Namespace)
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
					Expect(e2e.LogStore.HasInfraStructureLogs(helpers.DefaultWaitForLogsTimeout)).To(BeTrue(), "Expected to find stored infrastructure logs")
					_, _ = e2e.LogStore.GrepLogs(waitlogs, helpers.DefaultWaitForLogsTimeout)
					// typical value of tag in this test case is: kubernetes.var.log.containers.log-generator-746f659fdf-qgclg_clo-test-5764_log-generator-fef2a7848f9741bc6aeb1325aac051c0734c5dc177839c6787da207dc95530ad.log
					// prefix expected is: kubernetes#log-generator-746f659fdf-qgclg_clo-test-5764_log-generator-fef2a7848f9741bc6aeb1325aac051c0734c5dc177839c6787da207dc95530ad.log
					expectedTagPrefix := "kubernetes#"
					Expect(e2e.LogStore.GrepLogs(greptag, helpers.DefaultWaitForLogsTimeout)).To(HavePrefix(expectedTagPrefix))
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
					forwarder.Spec.Outputs[0].URL = fmt.Sprintf("%s.%s.svc:24224", syslogDeployment.ObjectMeta.Name, syslogDeployment.Namespace)
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
					Expect(e2e.LogStore.HasInfraStructureLogs(helpers.DefaultWaitForLogsTimeout)).To(BeTrue(), "Expected to find stored infrastructure logs")
					_, _ = e2e.LogStore.GrepLogs(waitlogs, helpers.DefaultWaitForLogsTimeout)
					// 134 = Facility(local0/16)*8 + Severity(Informational/6)
					// The 1 after <134> is version, which is always set to 1
					expectedPriority := "<134>1"
					grepPri := fmt.Sprintf(rsyslogFormatStr, logGenPod, "$1")
					Expect(e2e.LogStore.GrepLogs(grepPri, helpers.DefaultWaitForLogsTimeout)).To(Equal(expectedPriority), "Expected: "+expectedPriority)
				})
			})
			Describe("syslog payload", func() {
				Context("with new plugin", func() {
					BeforeEach(func() {
						generatorPayload["tag_key"] = "rec_tag"
					})
					It("should take from payload_key", func() {
						if syslogDeployment, err = e2e.DeploySyslogReceiver(testDir, corev1.ProtocolTCP, true, helpers.RFC3164); err != nil {
							Fail(fmt.Sprintf("Unable to deploy syslog receiver: %v", err))
						}
						forwarder.Spec.Outputs[0].URL = fmt.Sprintf("%s.%s.svc:24224", syslogDeployment.ObjectMeta.Name, syslogDeployment.Namespace)
						forwarder.Spec.Outputs[0].Secret = &logging.OutputSecretSpec{
							Name: syslogDeployment.ObjectMeta.Name,
						}
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
						Expect(e2e.LogStore.HasInfraStructureLogs(helpers.DefaultWaitForLogsTimeout)).To(BeTrue(), "Expected to find stored infrastructure logs")
						grepMsgContent := fmt.Sprintf(`grep %s %%s | tail -n 1 | awk -F' ' '{ s = ""; for (i = 8; i <= NF; i++) s = s $i " "; print s }'`, "rec_tag")
						str, err := e2e.LogStore.GrepLogs(grepMsgContent, helpers.DefaultWaitForLogsTimeout)
						Expect(err).To(BeNil())
						str = strings.ReplaceAll(str, "=>", ":")
						msg := map[string]interface{}{}
						err = json.Unmarshal([]byte(str), &msg)
						Expect(err).To(BeNil())
						Expect(msg["msgcontent"]).To(Equal("My life is my message"))
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
