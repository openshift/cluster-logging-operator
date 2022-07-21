package functional

import (
	"fmt"
	"strings"
	"time"
)

var (
	//Timestamp = 2021-07-06T08:26:58.687Z
	OVNLogTemplate            = "%s|00004|acl_log(ovn_pinctrl0)|INFO|name=verify-audit-logging_deny-all, verdict=drop"
	KubeAuditLogTemplate      = `{"kind":"Event","apiVersion":"audit.k8s.io/v1","level":"Metadata","auditID":"a6299d35-5759-4f67-9bed-2b962cf21cf3","stage":"ResponseComplete","requestURI":"/api/v1/namespaces/openshift-kube-storage-version-migrator/serviceaccounts/kube-storage-version-migrator-sa","verb":"get","user":{"username":"system:serviceaccount:openshift-kube-storage-version-migrator-operator:kube-storage-version-migrator-operator","uid":"d40a1a15-8b96-4ffa-a56b-5a834583532e","groups":["system:serviceaccounts","system:serviceaccounts:openshift-kube-storage-version-migrator-operator","system:authenticated"]},"sourceIPs":["10.128.0.16"],"userAgent":"cluster-kube-storage-version-migrator-operator/v0.0.0 (linux/amd64) kubernetes/$Format","objectRef":{"resource":"serviceaccounts","namespace":"openshift-kube-storage-version-migrator","name":"kube-storage-version-migrator-sa","apiVersion":"v1"},"responseStatus":{"metadata":{},"code":200},"requestReceivedTimestamp":"%s","stageTimestamp":"%s","annotations":{"authentication.k8s.io/legacy-token":"system:serviceaccount:openshift-kube-storage-version-migrator-operator:kube-storage-version-migrator-operator","authorization.k8s.io/decision":"allow","authorization.k8s.io/reason":"RBAC: allowed by ClusterRoleBinding \"system:openshift:operator:kube-storage-version-migrator-operator\" of ClusterRole \"cluster-admin\" to ServiceAccount \"kube-storage-version-migrator-operator/openshift-kube-storage-version-migrator-operator\""}}`
	OpenShiftAuditLogTemplate = `{"kind":"Event","apiVersion":"audit.k8s.io/v1","level":"Metadata","auditID":"19f44b1a-e4fb-4c9a-bc2f-068dc94be8fb","stage":"ResponseComplete","requestURI":"/","verb":"get","user":{"username":"system:anonymous","groups":["system:unauthenticated"]},"sourceIPs":["10.128.0.1"],"userAgent":"Go-http-client/1.1","responseStatus":{"metadata":{},"status":"Failure","reason":"Forbidden","code":403},"requestReceivedTimestamp":"%s","stageTimestamp":"%s","annotations":{"authorization.k8s.io/decision":"forbid","authorization.k8s.io/reason":""}}`
)

func NewKubeAuditLog(eventTime time.Time) string {
	now := CRIOTime(eventTime)
	return fmt.Sprintf(KubeAuditLogTemplate, now, now)
}

func NewAuditHostLog(eventTime time.Time) string {
	now := fmt.Sprintf("%.3f", float64(eventTime.UnixNano())/float64(time.Second))
	return fmt.Sprintf(`type=DAEMON_START msg=audit(%s:2914): op=start ver=3.0 format=enriched kernel=4.18.0-240.15.1.el8_3.x86_64 auid=4294967295 pid=1396 uid=0 ses=4294967295 subj=system_u:system_r:auditd_t:s0 res=successAUID="unset" UID="root"`, now)
}
func NewOVNAuditLog(eventTime time.Time) string {
	now := CRIOTime(eventTime)
	return fmt.Sprintf(OVNLogTemplate, now)
}

func NewPartialCRIOLogMessage(timestamp, message string) string {
	return NewCRIOLogMessage(timestamp, message, true)
}

func NewFullCRIOLogMessage(timestamp, message string) string {
	return NewCRIOLogMessage(timestamp, message, false)
}

func NewCRIOLogMessage(timestamp, message string, partial bool) string {
	fullOrPartial := "F"
	if partial {
		fullOrPartial = "P"
	}
	return fmt.Sprintf("%s stdout %s %s", timestamp, fullOrPartial, message)
}

// CRIOTime returns the CRIO string format of time t.
func CRIOTime(t time.Time) string { return t.UTC().Format(time.RFC3339Nano) }

func CreateAppLogFromJson(jsonstr string) string {
	jsonMsg := strings.ReplaceAll(jsonstr, "\n", "")
	timestamp := "2020-11-04T18:13:59.061892+00:00"

	return fmt.Sprintf("%s stdout F %s", timestamp, jsonMsg)
}
