package outputs

import (
	"fmt"
	"strings"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	logging "github.com/openshift/cluster-logging-operator/pkg/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/test"
	"github.com/openshift/cluster-logging-operator/test/functional"
	//. "github.com/openshift/cluster-logging-operator/test/matchers"
)

var _ = Describe("[LogForwarding][Syslog] Functional tests", func() {

	var (
		framework *functional.FluentdFunctionalFramework
	)

	BeforeEach(func() {
		framework = functional.NewFluentdFunctionalFramework()
	})
	AfterEach(func() {
		framework.Cleanup()
	})

	setSyslogSpecValues := func(outspec *logging.OutputSpec) {
		outspec.Syslog = &logging.Syslog{
			Facility: "user",
			Severity: "debug",
			AppName:  "myapp",
			ProcID:   "myproc",
			MsgID:    "mymsg",
			RFC:      "RFC5424",
		}
	}

	join := func(
		f1 func(spec *logging.OutputSpec),
		f2 func(spec *logging.OutputSpec)) func(*logging.OutputSpec) {
		return func(s *logging.OutputSpec) {
			f1(s)
			f2(s)
		}
	}

	getAppName := func(fields []string) string {
		return fields[3]
	}
	getProcID := func(fields []string) string {
		return fields[4]
	}
	getMsgID := func(fields []string) string {
		return fields[5]
	}

	timestamp := "2013-03-28T14:36:03.243000+00:00"

	Context("Application Logs", func() {
		It("should send large message over UDP", func() {
			functional.NewClusterLogForwarderBuilder(framework.Forwarder).
				FromInput(logging.InputNameApplication).
				ToOutputWithVisitor(join(setSyslogSpecValues, func(spec *logging.OutputSpec) {
					spec.URL = "udp://0.0.0.0:24224"
				}), logging.OutputTypeSyslog)
			Expect(framework.Deploy()).To(BeNil())

			var MaxLen uint64 = 40000
			Expect(framework.WritesNApplicationLogsOfSize(1, MaxLen)).To(BeNil())
			// Read line from Syslog output
			outputlogs, err := framework.ReadApplicationLogsFrom(logging.OutputTypeSyslog)
			Expect(err).To(BeNil(), "Expected no errors reading the logs")
			Expect(outputlogs).ToNot(BeEmpty())
			fields := strings.Split(outputlogs[0], "#011")
			msg := fields[2]
			// adjust for "message:" prefix in the received message
			ReceivedLen := uint64(len(msg[8:]))
			Expect(ReceivedLen).To(Equal(MaxLen))
		})
		It("should send NonJson App logs to syslog", func() {
			functional.NewClusterLogForwarderBuilder(framework.Forwarder).
				FromInput(logging.InputNameApplication).
				ToOutputWithVisitor(setSyslogSpecValues, logging.OutputTypeSyslog)
			Expect(framework.Deploy()).To(BeNil())

			// Log message data
			//Expect(framework.WriteMessagesToApplicationLog("hello world", 10)).To(BeNil())
			for _, log := range NonJsonAppLogs {
				//log = test.Escapelines(log)
				log = functional.NewFullCRIOLogMessage(timestamp, log)
				Expect(framework.WriteMessagesToApplicationLog(log, 1)).To(BeNil())
			}
			// Read line from Syslog output
			outputlogs, err := framework.ReadApplicationLogsFrom(logging.OutputTypeSyslog)
			Expect(err).To(BeNil(), "Expected no errors reading the logs")
			Expect(outputlogs).ToNot(BeEmpty())
			fields := strings.Split(outputlogs[0], " ")
			Expect(getAppName(fields)).To(Equal("myapp"))
			Expect(getProcID(fields)).To(Equal("myproc"))
			Expect(getMsgID(fields)).To(Equal("mymsg"))
		})
		It("should take values of appname, procid, messageid from record", func() {
			functional.NewClusterLogForwarderBuilder(framework.Forwarder).
				FromInput(logging.InputNameApplication).
				ToOutputWithVisitor(join(setSyslogSpecValues, func(spec *logging.OutputSpec) {
					spec.Syslog.AppName = "$.message.appname_key"
					spec.Syslog.ProcID = "$.message.procid_key"
					spec.Syslog.MsgID = "$.message.msgid_key"
				}), logging.OutputTypeSyslog)
			Expect(framework.Deploy()).To(BeNil())

			// Log message data
			for _, log := range JSONApplicationLogs {
				log = test.Escapelines(log)
				log = functional.NewFullCRIOLogMessage(timestamp, log)
				Expect(framework.WriteMessagesToApplicationLog(log, 1)).To(BeNil())
			}
			// Read line from Syslog output
			outputlogs, err := framework.ReadApplicationLogsFrom(logging.OutputTypeSyslog)
			Expect(err).To(BeNil(), "Expected no errors reading the logs")
			Expect(outputlogs).ToNot(BeEmpty())
			fields := strings.Split(outputlogs[0], " ")
			Expect(getAppName(fields)).To(Equal("rec_appname"))
			Expect(getProcID(fields)).To(Equal("rec_procid"))
			Expect(getMsgID(fields)).To(Equal("rec_msgid"))
		})
		It("should take values from fluent tag", func() {
			functional.NewClusterLogForwarderBuilder(framework.Forwarder).
				FromInput(logging.InputNameApplication).
				ToOutputWithVisitor(join(setSyslogSpecValues, func(spec *logging.OutputSpec) {
					spec.Syslog.AppName = "tag"
				}), logging.OutputTypeSyslog)
			Expect(framework.Deploy()).To(BeNil())

			// Log message data
			for _, log := range JSONApplicationLogs {
				log = test.Escapelines(log)
				log = functional.NewFullCRIOLogMessage(timestamp, log)
				Expect(framework.WriteMessagesToApplicationLog(log, 1)).To(BeNil())
			}
			// Read line from Syslog output
			outputlogs, err := framework.ReadApplicationLogsFrom(logging.OutputTypeSyslog)
			Expect(err).To(BeNil(), "Expected no errors reading the logs")
			Expect(outputlogs).ToNot(BeEmpty())
			fields := strings.Split(outputlogs[0], " ")
			Expect(getAppName(fields)).To(HavePrefix("kubernetes."))
		})
	})
	Context("Audit logs", func() {
		It("should send kubernetes audit logs", func() {
			functional.NewClusterLogForwarderBuilder(framework.Forwarder).
				FromInput(logging.InputNameAudit).
				ToOutputWithVisitor(setSyslogSpecValues, logging.OutputTypeSyslog)
			Expect(framework.Deploy()).To(BeNil())

			// Log message data
			for _, log := range K8sAuditLogs {
				log = test.Escapelines(log)
				//log = functional.NewFullCRIOLogMessage(timestamp, log)
				Expect(framework.WriteMessagesTok8sAuditLog(log, 1)).To(BeNil())
			}

			time.Sleep(time.Minute * 5)
			// Read line from Syslog output
			outputlogs, err := framework.ReadAuditLogsFrom(logging.OutputTypeSyslog)
			Expect(err).To(BeNil(), "Expected no errors reading the logs")
			Expect(outputlogs).ToNot(BeEmpty())
			for _, o := range outputlogs {
				fmt.Printf("log received %s\n", o)
			}
			fields := strings.Split(outputlogs[0], " ")
			Expect(getAppName(fields)).To(Equal("myapp"))
			Expect(getProcID(fields)).To(Equal("myproc"))
			Expect(getMsgID(fields)).To(Equal("mymsg"))
		})
		It("should send openshift audit logs", func() {
			functional.NewClusterLogForwarderBuilder(framework.Forwarder).
				FromInput(logging.InputNameAudit).
				ToOutputWithVisitor(setSyslogSpecValues, logging.OutputTypeSyslog)
			Expect(framework.Deploy()).To(BeNil())

			// Log message data
			for _, log := range OpenshiftAuditLogs {
				log = test.Escapelines(log)
				//log = functional.NewFullCRIOLogMessage(timestamp, log)
				Expect(framework.WriteMessagesToOpenshiftAuditLog(log, 1)).To(BeNil())
			}

			// Read line from Syslog output
			outputlogs, err := framework.ReadAuditLogsFrom(logging.OutputTypeSyslog)
			Expect(err).To(BeNil(), "Expected no errors reading the logs")
			Expect(outputlogs).ToNot(BeEmpty())
			fields := strings.Split(outputlogs[0], " ")
			Expect(getAppName(fields)).To(Equal("myapp"))
			Expect(getProcID(fields)).To(Equal("myproc"))
			Expect(getMsgID(fields)).To(Equal("mymsg"))
		})
	})
})

var (
	JSONApplicationLogs = []string{
		`{"appname_key":"rec_appname","msgcontent":"My life is my message","msgid_key":"rec_msgid","procid_key":"rec_procid","timestamp":"2021-02-16 18:54:52"}`,
		/**/
		`{"appname_key":"rec_appname","msgcontent":"My life is my message","msgid_key":"rec_msgid","procid_key":"rec_procid","timestamp":"2021-02-16 18:54:53"}`,
		`{"appname_key":"rec_appname","msgcontent":"My life is my message","msgid_key":"rec_msgid","procid_key":"rec_procid","timestamp":"2021-02-16 18:54:54"}`,
		`{"appname_key":"rec_appname","msgcontent":"My life is my message","msgid_key":"rec_msgid","procid_key":"rec_procid","timestamp":"2021-02-16 18:54:55"}`,
		`{"appname_key":"rec_appname","msgcontent":"My life is my message","msgid_key":"rec_msgid","procid_key":"rec_procid","timestamp":"2021-02-16 18:54:56"}`,
		`{"appname_key":"rec_appname","msgcontent":"My life is my message","msgid_key":"rec_msgid","procid_key":"rec_procid","timestamp":"2021-02-16 18:54:57"}`,
		`{"appname_key":"rec_appname","msgcontent":"My life is my message","msgid_key":"rec_msgid","procid_key":"rec_procid","timestamp":"2021-02-16 18:54:58"}`,
		`{"appname_key":"rec_appname","msgcontent":"My life is my message","msgid_key":"rec_msgid","procid_key":"rec_procid","timestamp":"2021-02-16 18:54:59"}`,
		`{"appname_key":"rec_appname","msgcontent":"My life is my message","msgid_key":"rec_msgid","procid_key":"rec_procid","timestamp":"2021-02-16 18:55:00"}`,
		`{"appname_key":"rec_appname","msgcontent":"My life is my message","msgid_key":"rec_msgid","procid_key":"rec_procid","timestamp":"2021-02-16 18:55:01"}`,
	}

	NonJsonAppLogs = []string{
		`2021-02-17 17:46:27 "hello world"`,
		`2021-02-17 17:46:28 "hello world"`,
		`2021-02-17 17:46:29 "hello world"`,
		`2021-02-17 17:46:30 "hello world"`,
		`2021-02-17 17:46:31 "hello world"`,
		`2021-02-17 17:46:32 "hello world"`,
		`2021-02-17 17:46:33 "hello world"`,
		`2021-02-17 17:46:34 "hello world"`,
		`2021-02-17 17:46:35 "hello world"`,
		`2021-02-17 17:46:36 "hello world"`,
	}

	K8sAuditLogs = []string{
		`{"kind":"Event","apiVersion":"audit.k8s.io/v1","level":"Metadata","auditID":"45ad7c9f-486d-46c7-b29f-8fdf46784ea9","stage":"ResponseComplete","requestURI":"/apis/monitoring.coreos.com/v1/namespaces/openshift-logging/prometheusrules","verb":"create","user":{"username":"system:serviceaccount:openshift-operators-redhat:elasticsearch-operator","uid":"ba1589ec-be2b-4d61-9946-c1e3b6f635d2","groups":["system:serviceaccounts","system:serviceaccounts:openshift-operators-redhat","system:authenticated"]},"sourceIPs":["63.29.116.39"],"userAgent":"elasticsearch-operator/v0.0.0 (linux/amd64) kubernetes/$Format","objectRef":{"resource":"prometheusrules","namespace":"openshift-logging","name":"elasticsearch-prometheus-rules","apiGroup":"monitoring.coreos.com","apiVersion":"v1"},"responseStatus":{"metadata":{},"status":"Failure","reason":"AlreadyExists","code":409},"requestReceivedTimestamp":"2021-03-12T15:11:04.610686Z","stageTimestamp":"2021-03-12T15:11:04.625042Z","annotations":{"authentication.k8s.io/legacy-token":"system:serviceaccount:openshift-operators-redhat:elasticsearch-operator","authorization.k8s.io/decision":"allow","authorization.k8s.io/reason":"RBAC: allowed by ClusterRoleBinding \"elasticsearch-operator.4.6.0-202102130420.p0-6c59c9c74d\" of ClusterRole \"elasticsearch-operator.4.6.0-202102130420.p0-6c59c9c74d\" to ServiceAccount \"elasticsearch-operator/openshift-operators-redhat\""}}`,
		`{"kind":"Event","apiVersion":"audit.k8s.io/v1","level":"Metadata","auditID":"847f784a-fca5-43c8-84da-7518efa4f2eb","stage":"ResponseComplete","requestURI":"/api/v1/namespaces/openshift-kube-apiserver/pods?labelSelector=apiserver%3Dtrue","verb":"list","user":{"username":"system:serviceaccount:openshift-kube-apiserver-operator:kube-apiserver-operator","uid":"88dab0d1-833a-4e19-8a57-9f8719e02359","groups":["system:serviceaccounts","system:serviceaccounts:openshift-kube-apiserver-operator","system:authenticated"]},"sourceIPs":["10.128.0.24"],"userAgent":"cluster-kube-apiserver-operator/v0.0.0 (linux/amd64) kubernetes/$Format","objectRef":{"resource":"pods","namespace":"openshift-kube-apiserver","apiVersion":"v1"},"responseStatus":{"metadata":{},"code":200},"requestReceivedTimestamp":"2021-03-12T15:11:04.621761Z","stageTimestamp":"2021-03-12T15:11:04.626590Z","annotations":{"authentication.k8s.io/legacy-token":"system:serviceaccount:openshift-kube-apiserver-operator:kube-apiserver-operator","authorization.k8s.io/decision":"allow","authorization.k8s.io/reason":"RBAC: allowed by ClusterRoleBinding \"system:openshift:operator:kube-apiserver-operator\" of ClusterRole \"cluster-admin\" to ServiceAccount \"kube-apiserver-operator/openshift-kube-apiserver-operator\""}}`,
		`{"kind":"Event","apiVersion":"audit.k8s.io/v1","level":"Metadata","auditID":"f777f446-6432-4194-80c3-d10ef62c85be","stage":"ResponseComplete","requestURI":"/apis/monitoring.coreos.com/v1/namespaces/openshift-logging/prometheusrules/elasticsearch-prometheus-rules","verb":"update","user":{"username":"system:serviceaccount:openshift-operators-redhat:elasticsearch-operator","uid":"ba1589ec-be2b-4d61-9946-c1e3b6f635d2","groups":["system:serviceaccounts","system:serviceaccounts:openshift-operators-redhat","system:authenticated"]},"sourceIPs":["63.29.116.39"],"userAgent":"elasticsearch-operator/v0.0.0 (linux/amd64) kubernetes/$Format","objectRef":{"resource":"prometheusrules","namespace":"openshift-logging","name":"elasticsearch-prometheus-rules","uid":"9dae364b-67c5-43d6-8dc6-9f9f369b2534","apiGroup":"monitoring.coreos.com","apiVersion":"v1","resourceVersion":"7842583"},"responseStatus":{"metadata":{},"code":200},"requestReceivedTimestamp":"2021-03-12T15:11:04.626526Z","stageTimestamp":"2021-03-12T15:11:04.637990Z","annotations":{"authentication.k8s.io/legacy-token":"system:serviceaccount:openshift-operators-redhat:elasticsearch-operator","authorization.k8s.io/decision":"allow","authorization.k8s.io/reason":"RBAC: allowed by ClusterRoleBinding \"elasticsearch-operator.4.6.0-202102130420.p0-6c59c9c74d\" of ClusterRole \"elasticsearch-operator.4.6.0-202102130420.p0-6c59c9c74d\" to ServiceAccount \"elasticsearch-operator/openshift-operators-redhat\""}}`,
		`{"kind":"Event","apiVersion":"audit.k8s.io/v1","level":"Metadata","auditID":"ec4744ca-4fed-4c2e-9342-97bafb0057cd","stage":"ResponseComplete","requestURI":"/apis/authentication.k8s.io/v1/tokenreviews","verb":"create","user":{"username":"system:node:tdclsoshptva010.verizon.com","groups":["system:nodes","system:authenticated"]},"sourceIPs":["63.29.115.145"],"userAgent":"kubelet/v1.19.0+9f84db3 (linux/amd64) kubernetes/9f84db3","objectRef":{"resource":"tokenreviews","apiGroup":"authentication.k8s.io","apiVersion":"v1"},"responseStatus":{"metadata":{},"code":201},"requestReceivedTimestamp":"2021-03-12T15:11:04.640370Z","stageTimestamp":"2021-03-12T15:11:04.641962Z","annotations":{"authorization.k8s.io/decision":"allow","authorization.k8s.io/reason":""}}`,
		`{"kind":"Event","apiVersion":"audit.k8s.io/v1","level":"Metadata","auditID":"36dc2d61-9c12-49fe-aa9c-8dd4dbc174fc","stage":"ResponseComplete","requestURI":"/apis/coordination.k8s.io/v1/namespaces/kube-node-lease/leases/tdclsoshptva018.verizon.com?timeout=10s","verb":"update","user":{"username":"system:node:tdclsoshptva018.verizon.com","groups":["system:nodes","system:authenticated"]},"sourceIPs":["63.29.115.145"],"userAgent":"kubelet/v1.19.0+9f84db3 (linux/amd64) kubernetes/9f84db3","objectRef":{"resource":"leases","namespace":"kube-node-lease","name":"tdclsoshptva018.verizon.com","uid":"135d02d3-47ec-46ba-84ba-7a9361cf7067","apiGroup":"coordination.k8s.io","apiVersion":"v1","resourceVersion":"25718357"},"responseStatus":{"metadata":{},"code":200},"requestReceivedTimestamp":"2021-03-12T15:11:04.653148Z","stageTimestamp":"2021-03-12T15:11:04.658278Z","annotations":{"authorization.k8s.io/decision":"allow","authorization.k8s.io/reason":""}}`,
		`{"kind":"Event","apiVersion":"audit.k8s.io/v1","level":"Metadata","auditID":"d3148006-7438-4a38-bc25-91bbb4c67267","stage":"ResponseComplete","requestURI":"/apis/coordination.k8s.io/v1/namespaces/kube-node-lease/leases/tdclsoshptva012.verizon.com?timeout=10s","verb":"update","user":{"username":"system:node:tdclsoshptva012.verizon.com","groups":["system:nodes","system:authenticated"]},"sourceIPs":["63.29.115.145"],"userAgent":"kubelet/v1.19.0+9f84db3 (linux/amd64) kubernetes/9f84db3","objectRef":{"resource":"leases","namespace":"kube-node-lease","name":"tdclsoshptva012.verizon.com","uid":"f076dde3-b46f-458f-8256-7d0e839b9d7c","apiGroup":"coordination.k8s.io","apiVersion":"v1","resourceVersion":"25718358"},"responseStatus":{"metadata":{},"code":200},"requestReceivedTimestamp":"2021-03-12T15:11:04.674175Z","stageTimestamp":"2021-03-12T15:11:04.679253Z","annotations":{"authorization.k8s.io/decision":"allow","authorization.k8s.io/reason":""}}`,
		`{"kind":"Event","apiVersion":"audit.k8s.io/v1","level":"Metadata","auditID":"1f9633aa-c087-447f-b6b7-6d7abc8f7b81","stage":"ResponseComplete","requestURI":"/apis/coordination.k8s.io/v1/namespaces/kube-node-lease/leases/tdclsoshptva007.verizon.com?timeout=10s","verb":"update","user":{"username":"system:node:tdclsoshptva007.verizon.com","groups":["system:nodes","system:authenticated"]},"sourceIPs":["63.29.115.145"],"userAgent":"kubelet/v1.19.0+9f84db3 (linux/amd64) kubernetes/9f84db3","objectRef":{"resource":"leases","namespace":"kube-node-lease","name":"tdclsoshptva007.verizon.com","uid":"fdf21a46-9c64-49ab-94fa-6c89fa1012af","apiGroup":"coordination.k8s.io","apiVersion":"v1","resourceVersion":"25718359"},"responseStatus":{"metadata":{},"code":200},"requestReceivedTimestamp":"2021-03-12T15:11:04.679575Z","stageTimestamp":"2021-03-12T15:11:04.685269Z","annotations":{"authorization.k8s.io/decision":"allow","authorization.k8s.io/reason":""}}`,
		`{"kind":"Event","apiVersion":"audit.k8s.io/v1","level":"Metadata","auditID":"1fd61cd1-3bc9-4031-a7cf-62affd986abc","stage":"ResponseComplete","requestURI":"/api/v1/namespaces/openshift-kube-storage-version-migrator","verb":"get","user":{"username":"system:serviceaccount:openshift-kube-storage-version-migrator-operator:kube-storage-version-migrator-operator","uid":"b407f494-ca8c-4026-803b-e0c750bb9a66","groups":["system:serviceaccounts","system:serviceaccounts:openshift-kube-storage-version-migrator-operator","system:authenticated"]},"sourceIPs":["10.128.0.27"],"userAgent":"cluster-kube-storage-version-migrator-operator/v0.0.0 (linux/amd64) kubernetes/$Format","objectRef":{"resource":"namespaces","namespace":"openshift-kube-storage-version-migrator","name":"openshift-kube-storage-version-migrator","apiVersion":"v1"},"responseStatus":{"metadata":{},"code":200},"requestReceivedTimestamp":"2021-03-12T15:11:04.683187Z","stageTimestamp":"2021-03-12T15:11:04.685427Z","annotations":{"authentication.k8s.io/legacy-token":"system:serviceaccount:openshift-kube-storage-version-migrator-operator:kube-storage-version-migrator-operator","authorization.k8s.io/decision":"allow","authorization.k8s.io/reason":"RBAC: allowed by ClusterRoleBinding \"system:openshift:operator:kube-storage-version-migrator-operator\" of ClusterRole \"cluster-admin\" to ServiceAccount \"kube-storage-version-migrator-operator/openshift-kube-storage-version-migrator-operator\""}}`,
		`{"kind":"Event","apiVersion":"audit.k8s.io/v1","level":"Metadata","auditID":"daffa6ef-f273-4e9f-ae82-353a375029a7","stage":"ResponseComplete","requestURI":"/api/v1/namespaces/openshift-logging/secrets","verb":"create","user":{"username":"system:serviceaccount:openshift-logging:cluster-logging-operator","uid":"a4400956-f5f9-41d0-a26d-6ce8a849436e","groups":["system:serviceaccounts","system:serviceaccounts:openshift-logging","system:authenticated"]},"sourceIPs":["63.29.116.32"],"userAgent":"cluster-logging-operator/v0.0.0 (linux/amd64) kubernetes/$Format","objectRef":{"resource":"secrets","namespace":"openshift-logging","name":"elasticsearch","apiVersion":"v1"},"responseStatus":{"metadata":{},"status":"Failure","reason":"AlreadyExists","code":409},"requestReceivedTimestamp":"2021-03-12T15:11:04.740708Z","stageTimestamp":"2021-03-12T15:11:04.748684Z","annotations":{"authentication.k8s.io/legacy-token":"system:serviceaccount:openshift-logging:cluster-logging-operator","authorization.k8s.io/decision":"allow","authorization.k8s.io/reason":"RBAC: allowed by RoleBinding \"clusterlogging.4.6.0-202102130420.p0-cluster-logging-6d9c4dfd7f/openshift-logging\" of Role \"clusterlogging.4.6.0-202102130420.p0-cluster-logging-6d9c4dfd7f\" to ServiceAccount \"cluster-logging-operator/openshift-logging\""}}`,
		`{"kind":"Event","apiVersion":"audit.k8s.io/v1","level":"Metadata","auditID":"8efacc26-cca6-46e3-9610-bd2df5a88b4e","stage":"ResponseComplete","requestURI":"/api/v1/namespaces/openshift-kube-controller-manager-operator/services?resourceVersion=25701453","verb":"list","user":{"username":"system:serviceaccount:openshift-monitoring:prometheus-k8s","uid":"4e515dd7-0617-4696-8e71-86ffd082801e","groups":["system:serviceaccounts","system:serviceaccounts:openshift-monitoring","system:authenticated"]},"sourceIPs":["63.29.116.44"],"userAgent":"Prometheus/2.21.0","objectRef":{"resource":"services","namespace":"openshift-kube-controller-manager-operator","apiVersion":"v1"},"responseStatus":{"metadata":{},"code":200},"requestReceivedTimestamp":"2021-03-12T15:11:04.753641Z","stageTimestamp":"2021-03-12T15:11:04.754328Z","annotations":{"authentication.k8s.io/legacy-token":"system:serviceaccount:openshift-monitoring:prometheus-k8s","authorization.k8s.io/decision":"allow","authorization.k8s.io/reason":"RBAC: allowed by ClusterRoleBinding \"elasticsearch-metrics\" of ClusterRole \"elasticsearch-metrics\" to ServiceAccount \"prometheus-k8s/openshift-monitoring\""}}`,
	}

	OpenshiftAuditLogs = []string{
		`{"kind":"Event","apiVersion":"audit.k8s.io/v1","level":"Metadata","auditID":"c1842bd8-535f-4f0b-b301-a516aea529a5","stage":"ResponseComplete","requestURI":"/openapi/v2","verb":"get","user":{"username":"system:aggregator","groups":["system:authenticated"]},"sourceIPs":["10.217.0.1"],"responseStatus":{"metadata":{},"code":304},"requestReceivedTimestamp":"2021-03-17T10:38:17.838892Z","stageTimestamp":"2021-03-17T10:38:17.841481Z","annotations":{"authorization.k8s.io/decision":"allow","authorization.k8s.io/reason":"RBAC: allowed by ClusterRoleBinding \"system:discovery\" of ClusterRole \"system:discovery\" to Group \"system:authenticated\""}}`,
		`{"kind":"Event","apiVersion":"audit.k8s.io/v1","level":"Metadata","auditID":"17a26963-7c8b-43db-a168-d7595169202e","stage":"ResponseComplete","requestURI":"/openapi/v2","verb":"get","user":{"username":"system:aggregator","groups":["system:authenticated"]},"sourceIPs":["10.217.0.1"],"responseStatus":{"metadata":{},"code":304},"requestReceivedTimestamp":"2021-03-17T10:38:17.842021Z","stageTimestamp":"2021-03-17T10:38:17.842274Z","annotations":{"authorization.k8s.io/decision":"allow","authorization.k8s.io/reason":"RBAC: allowed by ClusterRoleBinding \"system:discovery\" of ClusterRole \"system:discovery\" to Group \"system:authenticated\""}}`,
		`{"kind":"Event","apiVersion":"audit.k8s.io/v1","level":"Metadata","auditID":"f4a22397-c644-48e6-9db6-5008b69e5302","stage":"ResponseComplete","requestURI":"/openapi/v2","verb":"get","user":{"username":"system:aggregator","groups":["system:authenticated"]},"sourceIPs":["10.217.0.1"],"responseStatus":{"metadata":{},"code":304},"requestReceivedTimestamp":"2021-03-17T10:38:17.842677Z","stageTimestamp":"2021-03-17T10:38:17.842878Z","annotations":{"authorization.k8s.io/decision":"allow","authorization.k8s.io/reason":"RBAC: allowed by ClusterRoleBinding \"system:discovery\" of ClusterRole \"system:discovery\" to Group \"system:authenticated\""}}`,
		`{"kind":"Event","apiVersion":"audit.k8s.io/v1","level":"Metadata","auditID":"a1c58527-fa63-49c5-9e1f-d82cbe79d972","stage":"ResponseComplete","requestURI":"/openapi/v2","verb":"get","user":{"username":"system:aggregator","groups":["system:authenticated"]},"sourceIPs":["10.217.0.1"],"responseStatus":{"metadata":{},"code":304},"requestReceivedTimestamp":"2021-03-17T10:38:17.843357Z","stageTimestamp":"2021-03-17T10:38:17.843601Z","annotations":{"authorization.k8s.io/decision":"allow","authorization.k8s.io/reason":"RBAC: allowed by ClusterRoleBinding \"system:discovery\" of ClusterRole \"system:discovery\" to Group \"system:authenticated\""}}`,
		`{"kind":"Event","apiVersion":"audit.k8s.io/v1","level":"Metadata","auditID":"3aee04a2-5ba5-421a-9630-01556008fe55","stage":"ResponseComplete","requestURI":"/openapi/v2","verb":"get","user":{"username":"system:aggregator","groups":["system:authenticated"]},"sourceIPs":["10.217.0.1"],"responseStatus":{"metadata":{},"code":304},"requestReceivedTimestamp":"2021-03-17T10:38:17.844323Z","stageTimestamp":"2021-03-17T10:38:17.844558Z","annotations":{"authorization.k8s.io/decision":"allow","authorization.k8s.io/reason":"RBAC: allowed by ClusterRoleBinding \"system:discovery\" of ClusterRole \"system:discovery\" to Group \"system:authenticated\""}}`,
		`{"kind":"Event","apiVersion":"audit.k8s.io/v1","level":"Metadata","auditID":"02a7dfd5-50d0-4024-b7b5-31431e40b1c7","stage":"ResponseComplete","requestURI":"/openapi/v2","verb":"get","user":{"username":"system:aggregator","groups":["system:authenticated"]},"sourceIPs":["10.217.0.1"],"responseStatus":{"metadata":{},"code":304},"requestReceivedTimestamp":"2021-03-17T10:38:17.845004Z","stageTimestamp":"2021-03-17T10:38:17.845275Z","annotations":{"authorization.k8s.io/decision":"allow","authorization.k8s.io/reason":"RBAC: allowed by ClusterRoleBinding \"system:discovery\" of ClusterRole \"system:discovery\" to Group \"system:authenticated\""}}`,
		`{"kind":"Event","apiVersion":"audit.k8s.io/v1","level":"Metadata","auditID":"405147fb-ef88-44d6-b0c0-1bf680293e7d","stage":"ResponseComplete","requestURI":"/openapi/v2","verb":"get","user":{"username":"system:aggregator","groups":["system:authenticated"]},"sourceIPs":["10.217.0.1"],"responseStatus":{"metadata":{},"code":304},"requestReceivedTimestamp":"2021-03-17T10:38:17.847060Z","stageTimestamp":"2021-03-17T10:38:17.847294Z","annotations":{"authorization.k8s.io/decision":"allow","authorization.k8s.io/reason":"RBAC: allowed by ClusterRoleBinding \"system:discovery\" of ClusterRole \"system:discovery\" to Group \"system:authenticated\""}}`,
		`{"kind":"Event","apiVersion":"audit.k8s.io/v1","level":"Metadata","auditID":"511ffb85-7bac-4166-a6c1-0a3a5f7cf8df","stage":"ResponseComplete","requestURI":"/openapi/v2","verb":"get","user":{"username":"system:aggregator","groups":["system:authenticated"]},"sourceIPs":["10.217.0.1"],"responseStatus":{"metadata":{},"code":304},"requestReceivedTimestamp":"2021-03-17T10:38:17.847792Z","stageTimestamp":"2021-03-17T10:38:17.848070Z","annotations":{"authorization.k8s.io/decision":"allow","authorization.k8s.io/reason":"RBAC: allowed by ClusterRoleBinding \"system:discovery\" of ClusterRole \"system:discovery\" to Group \"system:authenticated\""}}`,
		`{"kind":"Event","apiVersion":"audit.k8s.io/v1","level":"Metadata","auditID":"49eae05e-5ba0-4ab9-96d4-5723440c2d2e","stage":"ResponseComplete","requestURI":"/openapi/v2","verb":"get","user":{"username":"system:aggregator","groups":["system:authenticated"]},"sourceIPs":["10.217.0.1"],"responseStatus":{"metadata":{},"code":304},"requestReceivedTimestamp":"2021-03-17T10:38:17.848724Z","stageTimestamp":"2021-03-17T10:38:17.849022Z","annotations":{"authorization.k8s.io/decision":"allow","authorization.k8s.io/reason":"RBAC: allowed by ClusterRoleBinding \"system:discovery\" of ClusterRole \"system:discovery\" to Group \"system:authenticated\""}}`,
		`{"kind":"Event","apiVersion":"audit.k8s.io/v1","level":"Metadata","auditID":"28c808cc-ff30-4995-8afa-3ca6b96428bf","stage":"ResponseComplete","requestURI":"/apis/route.openshift.io/v1/namespaces/openshift-authentication/routes/oauth-openshift","verb":"get","user":{"username":"system:serviceaccount:openshift-authentication-operator:authentication-operator","groups":["system:serviceaccounts","system:serviceaccounts:openshift-authentication-operator","system:authenticated"]},"sourceIPs":["10.217.0.24","10.217.0.1"],"userAgent":"authentication-operator/v0.0.0 (linux/amd64) kubernetes/$Format","objectRef":{"resource":"routes","namespace":"openshift-authentication","name":"oauth-openshift","apiGroup":"route.openshift.io","apiVersion":"v1"},"responseStatus":{"metadata":{},"code":200},"requestReceivedTimestamp":"2021-03-17T10:38:26.304916Z","stageTimestamp":"2021-03-17T10:38:26.310483Z","annotations":{"authorization.k8s.io/decision":"allow","authorization.k8s.io/reason":"RBAC: allowed by ClusterRoleBinding \"system:openshift:operator:authentication\" of ClusterRole \"cluster-admin\" to ServiceAccount \"authentication-operator/openshift-authentication-operator\""}}`,
		`{"kind":"Event","apiVersion":"audit.k8s.io/v1","level":"Metadata","auditID":"fea73a85-8ef6-4dc7-90cd-a15c9289c4dc","stage":"ResponseComplete","requestURI":"/apis/route.openshift.io/v1/namespaces/openshift-authentication/routes/oauth-openshift","verb":"get","user":{"username":"system:serviceaccount:openshift-authentication-operator:authentication-operator","groups":["system:serviceaccounts","system:serviceaccounts:openshift-authentication-operator","system:authenticated"]},"sourceIPs":["10.217.0.24","10.217.0.1"],"userAgent":"authentication-operator/v0.0.0 (linux/amd64) kubernetes/$Format","objectRef":{"resource":"routes","namespace":"openshift-authentication","name":"oauth-openshift","apiGroup":"route.openshift.io","apiVersion":"v1"},"responseStatus":{"metadata":{},"code":200},"requestReceivedTimestamp":"2021-03-17T10:38:26.312431Z","stageTimestamp":"2021-03-17T10:38:26.313960Z","annotations":{"authorization.k8s.io/decision":"allow","authorization.k8s.io/reason":"RBAC: allowed by ClusterRoleBinding \"system:openshift:operator:authentication\" of ClusterRole \"cluster-admin\" to ServiceAccount \"authentication-operator/openshift-authentication-operator\""}}`,
		`{"kind":"Event","apiVersion":"audit.k8s.io/v1","level":"Metadata","auditID":"dcd6f8b9-346f-4c09-8273-e45ec3e5129d","stage":"ResponseStarted","requestURI":"/apis/build.openshift.io/v1/builds?allowWatchBookmarks=true\u0026resourceVersion=1111225\u0026timeout=7m35s\u0026timeoutSeconds=455\u0026watch=true","verb":"watch","user":{"username":"system:kube-controller-manager","groups":["system:authenticated"]},"sourceIPs":["192.168.130.11","10.217.0.1"],"userAgent":"cluster-policy-controller/v0.0.0 (linux/amd64) kubernetes/$Format","objectRef":{"resource":"builds","apiGroup":"build.openshift.io","apiVersion":"v1"},"responseStatus":{"metadata":{},"status":"Success","message":"Connection closed early","code":200},"requestReceivedTimestamp":"2021-03-17T10:31:05.145955Z","stageTimestamp":"2021-03-17T10:38:40.164196Z","annotations":{"authorization.k8s.io/decision":"allow","authorization.k8s.io/reason":"RBAC: allowed by ClusterRoleBinding \"system:kube-controller-manager\" of ClusterRole \"system:kube-controller-manager\" to User \"system:kube-controller-manager\""}}`,
		`{"kind":"Event","apiVersion":"audit.k8s.io/v1","level":"Metadata","auditID":"dcd6f8b9-346f-4c09-8273-e45ec3e5129d","stage":"ResponseComplete","requestURI":"/apis/build.openshift.io/v1/builds?allowWatchBookmarks=true\u0026resourceVersion=1111225\u0026timeout=7m35s\u0026timeoutSeconds=455\u0026watch=true","verb":"watch","user":{"username":"system:kube-controller-manager","groups":["system:authenticated"]},"sourceIPs":["192.168.130.11","10.217.0.1"],"userAgent":"cluster-policy-controller/v0.0.0 (linux/amd64) kubernetes/$Format","objectRef":{"resource":"builds","apiGroup":"build.openshift.io","apiVersion":"v1"},"responseStatus":{"metadata":{},"status":"Success","message":"Connection closed early","code":200},"requestReceivedTimestamp":"2021-03-17T10:31:05.145955Z","stageTimestamp":"2021-03-17T10:38:40.164288Z","annotations":{"authorization.k8s.io/decision":"allow","authorization.k8s.io/reason":"RBAC: allowed by ClusterRoleBinding \"system:kube-controller-manager\" of ClusterRole \"system:kube-controller-manager\" to User \"system:kube-controller-manager\""}}`,
		`{"kind":"Event","apiVersion":"audit.k8s.io/v1","level":"Metadata","auditID":"72d9ea59-2969-435e-b300-2d464dc7b08d","stage":"ResponseComplete","requestURI":"/apis/route.openshift.io/v1/namespaces/openshift-ingress-canary/routes/canary","verb":"get","user":{"username":"system:serviceaccount:openshift-ingress-operator:ingress-operator","groups":["system:serviceaccounts","system:serviceaccounts:openshift-ingress-operator","system:authenticated"]},"sourceIPs":["10.217.0.5","10.217.0.1"],"userAgent":"ingress-operator/v0.0.0 (linux/amd64) kubernetes/$Format","objectRef":{"resource":"routes","namespace":"openshift-ingress-canary","name":"canary","apiGroup":"route.openshift.io","apiVersion":"v1"},"responseStatus":{"metadata":{},"code":200},"requestReceivedTimestamp":"2021-03-17T10:38:45.798670Z","stageTimestamp":"2021-03-17T10:38:45.806030Z","annotations":{"authorization.k8s.io/decision":"allow","authorization.k8s.io/reason":"RBAC: allowed by ClusterRoleBinding \"openshift-ingress-operator\" of ClusterRole \"openshift-ingress-operator\" to ServiceAccount \"ingress-operator/openshift-ingress-operator\""}}`,
		`{"kind":"Event","apiVersion":"audit.k8s.io/v1","level":"Metadata","auditID":"ae844730-d4b6-403a-9acc-b97c4e777171","stage":"ResponseComplete","requestURI":"/apis/route.openshift.io/v1/namespaces/openshift-authentication/routes/oauth-openshift","verb":"get","user":{"username":"system:serviceaccount:openshift-authentication-operator:authentication-operator","groups":["system:serviceaccounts","system:serviceaccounts:openshift-authentication-operator","system:authenticated"]},"sourceIPs":["10.217.0.24","10.217.0.1"],"userAgent":"authentication-operator/v0.0.0 (linux/amd64) kubernetes/$Format","objectRef":{"resource":"routes","namespace":"openshift-authentication","name":"oauth-openshift","apiGroup":"route.openshift.io","apiVersion":"v1"},"responseStatus":{"metadata":{},"code":200},"requestReceivedTimestamp":"2021-03-17T10:38:56.313256Z","stageTimestamp":"2021-03-17T10:38:56.322259Z","annotations":{"authorization.k8s.io/decision":"allow","authorization.k8s.io/reason":"RBAC: allowed by ClusterRoleBinding \"system:openshift:operator:authentication\" of ClusterRole \"cluster-admin\" to ServiceAccount \"authentication-operator/openshift-authentication-operator\""}}`,
		`{"kind":"Event","apiVersion":"audit.k8s.io/v1","level":"Metadata","auditID":"3519ad9f-920b-401b-8672-eb40a0a31105","stage":"ResponseComplete","requestURI":"/apis/route.openshift.io/v1/namespaces/openshift-ingress/routes?allowWatchBookmarks=true\u0026resourceVersion=1111233\u0026timeoutSeconds=560\u0026watch=true","verb":"watch","user":{"username":"system:serviceaccount:openshift-ingress-operator:ingress-operator","groups":["system:serviceaccounts","system:serviceaccounts:openshift-ingress-operator","system:authenticated"]},"sourceIPs":["10.217.0.5","10.217.0.1"],"userAgent":"ingress-operator/v0.0.0 (linux/amd64) kubernetes/$Format","objectRef":{"resource":"routes","namespace":"openshift-ingress","apiGroup":"route.openshift.io","apiVersion":"v1"},"responseStatus":{"metadata":{},"status":"Success","message":"Connection closed early","code":200},"requestReceivedTimestamp":"2021-03-17T10:29:52.835070Z","stageTimestamp":"2021-03-17T10:39:12.846248Z","annotations":{"authorization.k8s.io/decision":"allow","authorization.k8s.io/reason":"RBAC: allowed by ClusterRoleBinding \"openshift-ingress-operator\" of ClusterRole \"openshift-ingress-operator\" to ServiceAccount \"ingress-operator/openshift-ingress-operator\""}}`,
		`{"kind":"Event","apiVersion":"audit.k8s.io/v1","level":"Metadata","auditID":"03c80b7c-96da-41e4-8530-2771ef2da540","stage":"ResponseStarted","requestURI":"/apis/template.openshift.io/v1/brokertemplateinstances?allowWatchBookmarks=true\u0026resourceVersion=1111236\u0026timeout=8m40s\u0026timeoutSeconds=520\u0026watch=true","verb":"watch","user":{"username":"system:kube-controller-manager","groups":["system:authenticated"]},"sourceIPs":["192.168.130.11","10.217.0.1"],"userAgent":"kube-controller-manager/v1.20.0+bd9e442 (linux/amd64) kubernetes/bd9e442/kube-controller-manager","objectRef":{"resource":"brokertemplateinstances","apiGroup":"template.openshift.io","apiVersion":"v1"},"responseStatus":{"metadata":{},"status":"Success","message":"Connection closed early","code":200},"requestReceivedTimestamp":"2021-03-17T10:30:33.448881Z","stageTimestamp":"2021-03-17T10:39:13.468046Z","annotations":{"authorization.k8s.io/decision":"allow","authorization.k8s.io/reason":"RBAC: allowed by ClusterRoleBinding \"system:kube-controller-manager\" of ClusterRole \"system:kube-controller-manager\" to User \"system:kube-controller-manager\""}}`,
		`{"kind":"Event","apiVersion":"audit.k8s.io/v1","level":"Metadata","auditID":"03c80b7c-96da-41e4-8530-2771ef2da540","stage":"ResponseComplete","requestURI":"/apis/template.openshift.io/v1/brokertemplateinstances?allowWatchBookmarks=true\u0026resourceVersion=1111236\u0026timeout=8m40s\u0026timeoutSeconds=520\u0026watch=true","verb":"watch","user":{"username":"system:kube-controller-manager","groups":["system:authenticated"]},"sourceIPs":["192.168.130.11","10.217.0.1"],"userAgent":"kube-controller-manager/v1.20.0+bd9e442 (linux/amd64) kubernetes/bd9e442/kube-controller-manager","objectRef":{"resource":"brokertemplateinstances","apiGroup":"template.openshift.io","apiVersion":"v1"},"responseStatus":{"metadata":{},"status":"Success","message":"Connection closed early","code":200},"requestReceivedTimestamp":"2021-03-17T10:30:33.448881Z","stageTimestamp":"2021-03-17T10:39:13.468207Z","annotations":{"authorization.k8s.io/decision":"allow","authorization.k8s.io/reason":"RBAC: allowed by ClusterRoleBinding \"system:kube-controller-manager\" of ClusterRole \"system:kube-controller-manager\" to User \"system:kube-controller-manager\""}}
`,
	}
)
