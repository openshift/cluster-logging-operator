package types

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	_ "github.com/onsi/ginkgo" // Accept ginkgo command line options
)

const (
	ApplicationContainerLogStr = `
{
  "docker": {
    "container_id": "9694d9dcde514f37fe865bb4752114939490f28080f630231223953eb18fff57"
  },
  "kubernetes": {
    "container_name": "log-generator",
    "namespace_name": "clo-test-21151",
    "pod_name": "log-generator-595c967f99-59v44",
    "container_image": "quay.io/quay/busybox:latest",
    "container_image_id": "quay.io/quay/busybox@sha256:9f1c79411e054199210b4d489ae600a061595967adb643cd923f8515ad8123d2",
    "pod_id": "b902950c-90df-4ca0-b296-614471c38fd0",
    "host": "crc-j55b9-master-0",
    "master_url": "https://kubernetes.default.svc",
    "namespace_id": "86e19672-1671-472e-9b2e-6ca8c304f32b",
    "flat_labels": [
      "component=test",
      "logging-infra=log-generator",
      "pod-template-hash=595c967f99",
      "provider=openshift"
    ]
  },
  "message": "8316: My life is my message",
  "level": "unknown",
  "hostname": "crc-j55b9-master-0",
  "pipeline_metadata": {
    "collector": {
      "ipaddr4": "192.168.126.11",
      "inputname": "fluent-plugin-systemd",
      "name": "fluentd",
      "received_at": "2020-11-27T19:55:52.158588+00:00",
      "version": "1.7.4 1.6.0"
    }
  },
  "@timestamp": "2020-11-27T18:32:57.600159+00:00",
  "viaq_index_name": "app-write",
  "viaq_msg_id": "M2QxNzM0MmQtMmVmMy00NjM1LWE1YzAtYjE1MWMxOWE5MTM2"
}
	`

	InfraContainerLogStr = `
{
  "docker": {
    "container_id": "0b25616ba52ca96f0230779a72ea6abf541a7a06b586f93799052a7186dcb5c8"
  },
  "kubernetes": {
    "container_name": "download-server",
    "namespace_name": "openshift-console",
    "pod_name": "downloads-55f4ff79-gnrsw",
    "pod_id": "90831bef-2045-48d6-9d69-bb5156de67ff",
    "host": "crc-j55b9-master-0",
    "master_url": "https://kubernetes.default.svc",
    "namespace_id": "213a73f1-1b3d-4d32-9f68-ff6b7010dcfe",
    "flat_labels": [
      "app=console",
      "component=downloads",
      "pod-template-hash=55f4ff79"
    ]
  },
  "message": "::ffff:10.116.0.1 - - [29/Nov/2020 13:27:47] \"GET / HTTP/1.1\" 200 -",
  "level": "unknown",
  "hostname": "crc-j55b9-master-0",
  "pipeline_metadata": {
    "collector": {
      "ipaddr4": "192.168.126.11",
      "inputname": "fluent-plugin-systemd",
      "name": "fluentd",
      "received_at": "2020-11-29T13:27:48.886552+00:00",
      "version": "1.7.4 1.6.0"
    }
  },
  "@timestamp": "2020-11-29T13:27:47.955953+00:00",
  "viaq_index_name": "infra-write",
  "viaq_msg_id": "MjFlMWIxNGItMTljMi00NjA2LWFhNzUtNDg2OTYzYjQxYzUx"
}
	`
	JournalLogStr = `
{
  "_STREAM_ID": "827e6000d4cb4e72ba99af117275cffb",
  "_SYSTEMD_INVOCATION_ID": "1d6941944a444c199fbfa8497bbfcff1",
  "systemd": {
    "t": {
      "BOOT_ID": "c5b940038cdf487d9661b1aef19d26dc",
      "CAP_EFFECTIVE": "3fffffffff",
      "CMDLINE": "kubelet --node-ip=192.168.126.11 --config=/etc/kubernetes/kubelet.conf --bootstrap-kubeconfig=/etc/kubernetes/kubeconfig --kubeconfig=/var/lib/kubelet/kubeconfig --container-runtime=remote --container-runtime-endpoint=/var/run/crio/crio.sock --runtime-cgroups=/system.slice/crio.service --node-labels=node-role.kubernetes.io/master,node.openshift.io/os_id=rhcos --minimum-container-ttl-duration=6m0s --cloud-provider= --volume-plugin-dir=/etc/kubernetes/kubelet-plugins/volume/exec --register-with-taints=node-role.kubernetes.io/master=:NoSchedule --pod-infra-container-image=quay.io/openshift-release-dev/ocp-v4.0-art-dev@sha256:cc0eb0534a361cffb7c392bf889ef061fc5390a7b21f6c92710f7a53ac4edd85 --v=4",
      "COMM": "kubelet",
      "EXE": "/usr/bin/kubelet",
      "GID": "0",
      "MACHINE_ID": "2b2d5b75dec04ce889fb7ae3690bdca7",
      "PID": "2612",
      "SELINUX_CONTEXT": "system_u:system_r:unconfined_service_t:s0",
      "STREAM_ID": "827e6000d4cb4e72ba99af117275cffb",
      "SYSTEMD_CGROUP": "/system.slice/kubelet.service",
      "SYSTEMD_INVOCATION_ID": "1d6941944a444c199fbfa8497bbfcff1",
      "SYSTEMD_SLICE": "system.slice",
      "SYSTEMD_UNIT": "kubelet.service",
      "TRANSPORT": "stdout",
      "UID": "0"
    },
    "u": {
      "SYSLOG_IDENTIFIER": "hyperkube"
    }
  },
  "level": "info",
  "message": "I1128 18:18:17.286513    2612 desired_state_of_world_populator.go:344] Added volume \"cluster-samples-operator-token-rz7g2\" (volSpec=\"cluster-samples-operator-token-rz7g2\") for pod \"efe875e1-05bf-4169-b7d4-0891e856d4eb\" to desired state.",
  "hostname": "crc-j55b9-master-0",
  "pipeline_metadata": {
    "collector": {
      "ipaddr4": "192.168.126.11",
      "inputname": "fluent-plugin-systemd",
      "name": "fluentd",
      "received_at": "2020-11-28T18:18:17.845698+00:00",
      "version": "1.7.4 1.6.0"
    }
  },
  "@timestamp": "2020-11-28T18:18:17.286517+00:00",
  "viaq_index_name": "infra-write",
  "viaq_msg_id": "ZDY0ZjE4NTAtNGU3ZC00YmQ4LWJjYjctNzVjMTEwMzdlYWIz"
}
	`
	LinuxAuditLogStr = `
{
  "hostname": "crc-j55b9-master-0",
  "audit.linux": {
    "type": "ANOM_PROMISCUOUS",
    "record_id": "5964"
  },
  "message": "type=ANOM_PROMISCUOUS msg=audit(1606655808.785:5964): dev=veth823df183 prom=256 old_prom=0 auid=4294967295 uid=998 gid=996 ses=4294967295\u001dAUID=\"unset\" UID=\"etcd\" GID=\"cgred\"",
  "pipeline_metadata": {
    "collector": {
      "ipaddr4": "192.168.126.11",
      "inputname": "fluent-plugin-systemd",
      "name": "fluentd",
      "received_at": "2020-11-29T13:16:49.021493+00:00",
      "version": "1.7.4 1.6.0"
    }
  },
  "@timestamp": "2020-11-29T13:16:48.785000+00:00",
  "viaq_index_name": "audit-write",
  "viaq_msg_id": "Y2M1NThmYzUtODYxYS00MzY5LWJmZDQtN2FkYjk4ZDlmYjE3",
  "kubernetes": {}
}
   `
	K8sAuditLogStr = `
{
  "kind": "Event",
  "apiVersion": "audit.k8s.io/v1",
  "level": "info",
  "auditID": "e7b84ee2-04c0-4b3f-bf8a-51290816680e",
  "stage": "ResponseComplete",
  "requestURI": "/api/v1/namespaces/openshift-sdn/configmaps/openshift-network-controller",
  "verb": "update",
  "user": {
    "username": "system:serviceaccount:openshift-sdn:sdn-controller",
    "uid": "abcb0401-1236-4dce-ab7c-ba1d2948da46",
    "groups": [
      "system:serviceaccounts",
      "system:serviceaccounts:openshift-sdn",
      "system:authenticated"
    ]
  },
  "sourceIPs": [
    "192.168.130.11"
  ],
  "userAgent": "openshift-sdn-controller/v0.0.0 (linux/amd64) kubernetes/$Format",
  "objectRef": {
    "resource": "configmaps",
    "namespace": "openshift-sdn",
    "name": "openshift-network-controller",
    "uid": "a07dbbb1-8b53-4241-9153-c935c93c25b7",
    "apiVersion": "v1",
    "resourceVersion": "558761"
  },
  "responseStatus": {
    "code": 200
  },
  "requestReceivedTimestamp": "2020-11-27T19:55:01.798728Z",
  "stageTimestamp": "2020-11-27T19:55:01.801820Z",
  "annotations": {
    "authorization.k8s.io/decision": "allow",
    "authorization.k8s.io/reason": "RBAC: allowed by RoleBinding \"openshift-sdn-controller-leaderelection/openshift-sdn\" of Role \"openshift-sdn-controller-leaderelection\" to ServiceAccount \"sdn-controller/openshift-sdn\""
  },
  "k8s_audit_level": "Metadata",
  "message": null,
  "hostname": "crc-j55b9-master-0",
  "pipeline_metadata": {
    "collector": {
      "ipaddr4": "192.168.126.11",
      "inputname": "fluent-plugin-systemd",
      "name": "fluentd",
      "received_at": "2020-11-27T19:55:17.529590+00:00",
      "version": "1.7.4 1.6.0"
    }
  },
  "@timestamp": "2020-11-27T19:55:01.798728+00:00",
  "viaq_index_name": "audit-write",
  "viaq_msg_id": "OWU1OGU0MzYtOGQ4YS00MTBhLWIwZGQtMzM1ZDc3ZmIzOTc4",
  "kubernetes": {}
}
	`
	OpenshiftAuditLogStr = `
{
  "kind": "Event",
  "apiVersion": "audit.k8s.io/v1",
  "level": "info",
  "auditID": "e1c9ce68-dcde-4465-89f2-c9dd65da8139",
  "stage": "ResponseComplete",
  "requestURI": "/apis/security.openshift.io/v1/rangeallocations/scc-uid",
  "verb": "update",
  "user": {
    "username": "system:serviceaccount:openshift-infra:namespace-security-allocation-controller",
    "groups": [
      "system:serviceaccounts",
      "system:serviceaccounts:openshift-infra",
      "system:authenticated"
    ]
  },
  "sourceIPs": [
    "192.168.130.11",
    "10.116.0.1"
  ],
  "userAgent": "cluster-policy-controller/v0.0.0 (linux/amd64) kubernetes/$Format/system:serviceaccount:openshift-infra:namespace-security-allocation-controller",
  "objectRef": {
    "resource": "rangeallocations",
    "name": "scc-uid",
    "uid": "c51f9ef8-f6ac-4e14-be47-6322a37c6070",
    "apiGroup": "security.openshift.io",
    "apiVersion": "v1",
    "resourceVersion": "440262"
  },
  "responseStatus": {
    "code": 200
  },
  "requestReceivedTimestamp": "2020-11-29T13:26:56.978921Z",
  "stageTimestamp": "2020-11-29T13:26:56.992462Z",
  "annotations": {
    "authorization.k8s.io/decision": "allow",
    "authorization.k8s.io/reason": "RBAC: allowed by ClusterRoleBinding \"system:openshift:controller:namespace-security-allocation-controller\" of ClusterRole \"system:openshift:controller:namespace-security-allocation-controller\" to ServiceAccount \"namespace-security-allocation-controller/openshift-infra\""
  },
  "openshift_audit_level": "Metadata",
  "message": null,
  "hostname": "crc-j55b9-master-0",
  "pipeline_metadata": {
    "collector": {
      "ipaddr4": "192.168.126.11",
      "inputname": "fluent-plugin-systemd",
      "name": "fluentd",
      "received_at": "2020-11-29T13:26:57.026736+00:00",
      "version": "1.7.4 1.6.0"
    }
  },
  "@timestamp": "2020-11-29T13:26:56.978921+00:00",
  "viaq_index_name": "audit-write",
  "viaq_msg_id": "ZTRjMDQ4ZWEtMmVkMy00YTJmLTk0NTUtYTk4YzNjNDFlYjM5",
  "kubernetes": {}
}
`
	OVNAuditLogStr = `
{
  "hostname": "crc-j55b9-master-0",
  "message": "2021-07-06T08:26:58.687Z|00004|acl_log(ovn_pinctrl0)|INFO|name=verify-audit-logging_deny-all, verdict=drop",
  "pipeline_metadata": {
    "collector": {
      "ipaddr4": "192.168.126.11",
      "inputname": "fluent-plugin-systemd",
      "name": "fluentd",
      "received_at": "2020-11-29T13:16:49.021493+00:00",
      "version": "1.7.4 1.6.0"
    }
  },
  "@timestamp": "2021-07-06T08:26:58.687000+00:00",
  "viaq_index_name": "audit-write",
  "viaq_msg_id": "Y2M1NThmYzUtODYxYS00MzY5LWJmZDQtN2FkYjk4ZDlmYjE3",
  "kubernetes": {}
}
   `
)

func join(log ...string) string {
	return "[" + strings.Join(log, ",") + "]"
}

func TestDecodeApplicationLogs(t *testing.T) {
	var logs []ApplicationLog
	in := join(ApplicationContainerLogStr)
	dec := json.NewDecoder(strings.NewReader(in))
	dec.DisallowUnknownFields()
	err := dec.Decode(&logs)
	if err != nil {
		fmt.Printf("%#v", err)
		t.Fail()
	}
}

func TestDecodeInfraContainerLogs(t *testing.T) {
	var logs []InfraContainerLog
	in := join(InfraContainerLogStr)
	dec := json.NewDecoder(strings.NewReader(in))
	dec.DisallowUnknownFields()
	err := dec.Decode(&logs)
	if err != nil {
		fmt.Printf("%#v", err)
		t.Fail()
	}
}

func TestDecodeJournalLogs(t *testing.T) {
	var logs []JournalLog
	in := join(JournalLogStr)
	dec := json.NewDecoder(strings.NewReader(in))
	dec.DisallowUnknownFields()
	err := dec.Decode(&logs)
	if err != nil {
		fmt.Printf("%#v", err)
		t.Fail()
	}
}

// decode Infra Container and Journal logs together
func TestDecodeInfraLogs(t *testing.T) {
	var logs []InfraLog
	in := join(InfraContainerLogStr, JournalLogStr)
	dec := json.NewDecoder(strings.NewReader(in))
	dec.DisallowUnknownFields()
	err := dec.Decode(&logs)
	if err != nil {
		fmt.Printf("%#v", err)
		t.Fail()
	}
	if len(logs) != 2 {
		t.Fail()
	}
}

func TestDecodeLinuxAuditLogs(t *testing.T) {
	var logs []LinuxAuditLog
	in := join(LinuxAuditLogStr)
	dec := json.NewDecoder(strings.NewReader(in))
	dec.DisallowUnknownFields()
	err := dec.Decode(&logs)
	if err != nil {
		fmt.Printf("%#v", err)
		t.Fail()
	}
}

// Decode Ovn Audit logs
func TestDecodeOvnAuditLogs(t *testing.T) {
	var logs []OVNAuditLog
	in := join(OVNAuditLogStr)
	dec := json.NewDecoder(strings.NewReader(in))
	dec.DisallowUnknownFields()
	err := dec.Decode(&logs)
	if err != nil {
		fmt.Printf("%#v", err)
		t.Fail()
	}
}

func TestDecodeK8sAuditLogs(t *testing.T) {
	var logs []K8sAuditLog
	in := join(K8sAuditLogStr)
	dec := json.NewDecoder(strings.NewReader(in))
	dec.DisallowUnknownFields()
	err := dec.Decode(&logs)
	if err != nil {
		fmt.Printf("%#v", err)
		t.Fail()
	}
}

func TestDecodeOpenshiftAuditLogs(t *testing.T) {
	var logs []OpenshiftAuditLog
	in := join(OpenshiftAuditLogStr)
	dec := json.NewDecoder(strings.NewReader(in))
	dec.DisallowUnknownFields()
	err := dec.Decode(&logs)
	if err != nil {
		fmt.Printf("%#v", err)
		t.Fail()
	}
}

// decode Linux, K8s, Openshift Audit logs together
func TestDecodeAuditLogs(t *testing.T) {
	var logs []AuditLog
	in := join(
		LinuxAuditLogStr,
		K8sAuditLogStr,
		OpenshiftAuditLogStr)
	dec := json.NewDecoder(strings.NewReader(in))
	dec.DisallowUnknownFields()
	err := dec.Decode(&logs)
	if err != nil {
		fmt.Printf("%#v", err)
		t.Fail()
	}
	if len(logs) != 3 {
		t.Fail()
	}
}

func TestDecodeAllLogs(t *testing.T) {
	var logs []AllLog
	in := join(
		ApplicationContainerLogStr,
		JournalLogStr,
		InfraContainerLogStr,
		LinuxAuditLogStr,
		K8sAuditLogStr,
		OpenshiftAuditLogStr)
	dec := json.NewDecoder(strings.NewReader(in))
	dec.DisallowUnknownFields()
	err := dec.Decode(&logs)
	if err != nil {
		fmt.Printf("%#v", err)
		t.Fail()
	}
	if len(logs) != 6 {
		t.Fail()
	}
}
