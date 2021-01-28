package types

import (
	"bytes"
	"encoding/json"
	"errors"
	logger "github.com/ViaQ/logerr/log"
	"github.com/openshift/cluster-logging-operator/pkg/utils"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"math/rand"
	"strings"
	"time"
)

var ErrParse = errors.New("logs could not be parsed")

type Logs []AllLog

// ContainerLog
type ContainerLog struct {
	Timestamp        time.Time        `json:"@timestamp"`
	Docker           Docker           `json:"docker"`
	Kubernetes       Kubernetes       `json:"kubernetes"`
	Message          string           `json:"message"`
	Level            string           `json:"level"`
	Hostname         string           `json:"hostname"`
	PipelineMetadata PipelineMetadata `json:"pipeline_metadata"`
	ViaqIndexName    string           `json:"viaq_index_name"`
	ViaqMsgID        string           `json:"viaq_msg_id"`
	OpenshiftLabels  OpenshiftMeta    `json:"openshift"`
}

type Docker struct {
	ContainerID string `json:"container_id"`
}

type Kubernetes struct {
	ContainerName     string            `json:"container_name"`
	NamespaceName     string            `json:"namespace_name"`
	PodName           string            `json:"pod_name"`
	ContainerImage    string            `json:"container_image"`
	ContainerImageID  string            `json:"container_image_id"`
	PodID             string            `json:"pod_id"`
	Host              string            `json:"host"`
	MasterURL         string            `json:"master_url"`
	NamespaceID       string            `json:"namespace_id"`
	FlatLabels        []string          `json:"flat_labels"`
	Labels            map[string]string `json:"labels"`
	OrphanedNamespace string            `json:"orphaned_namespace"`
}

type Collector struct {
	Ipaddr4    string    `json:"ipaddr4"`
	Inputname  string    `json:"inputname"`
	Name       string    `json:"name"`
	ReceivedAt time.Time `json:"received_at"`
	Version    string    `json:"version"`
}

// EventData encodes an eventrouter event and previous event, with a verb for
// whether the event is created or updated.
type EventData struct {
	Verb     string    `json:"verb"`
	Event    *v1.Event `json:"event"`
	OldEvent *v1.Event `json:"old_event,omitempty"`
}

type PipelineMetadata struct {
	Collector Collector `json:"collector"`
}

type OpenshiftMeta struct {
	Labels map[string]string `json:"labels"`
}

// Application Logs are container logs from all namespaces except "openshift" and "openshift-*" namespaces
type ApplicationLog ContainerLog

// Infrastructure logs are
// - Journal logs
// - logs from "openshift" and "openshift-*" namespaces

// InfraContainerLog
// InfraContainerLog logs are container logs from "openshift" and "openshift-*" namespaces
type InfraContainerLog ContainerLog

// JournalLog is linux journal logs
type JournalLog struct {
	STREAMID            string           `json:"_STREAM_ID"`
	SYSTEMDINVOCATIONID string           `json:"_SYSTEMD_INVOCATION_ID"`
	Systemd             Systemd          `json:"systemd"`
	Level               string           `json:"level"`
	Message             string           `json:"message"`
	Hostname            string           `json:"hostname"`
	PipelineMetadata    PipelineMetadata `json:"pipeline_metadata"`
	Timestamp           time.Time        `json:"@timestamp"`
	ViaqIndexName       string           `json:"viaq_index_name"`
	ViaqMsgID           string           `json:"viaq_msg_id"`
	Kubernetes          Kubernetes       `json:"kubernetes"`
}

type T struct {
	BOOTID              string `json:"BOOT_ID"`
	CAPEFFECTIVE        string `json:"CAP_EFFECTIVE"`
	CMDLINE             string `json:"CMDLINE"`
	COMM                string `json:"COMM"`
	EXE                 string `json:"EXE"`
	GID                 string `json:"GID"`
	MACHINEID           string `json:"MACHINE_ID"`
	PID                 string `json:"PID"`
	SELINUXCONTEXT      string `json:"SELINUX_CONTEXT"`
	STREAMID            string `json:"STREAM_ID"`
	SYSTEMDCGROUP       string `json:"SYSTEMD_CGROUP"`
	SYSTEMDINVOCATIONID string `json:"SYSTEMD_INVOCATION_ID"`
	SYSTEMDSLICE        string `json:"SYSTEMD_SLICE"`
	SYSTEMDUNIT         string `json:"SYSTEMD_UNIT"`
	TRANSPORT           string `json:"TRANSPORT"`
	UID                 string `json:"UID"`
}

type U struct {
	SYSLOGIDENTIFIER string `json:"SYSLOG_IDENTIFIER"`
}

type Systemd struct {
	T T `json:"t"`
	U U `json:"u"`
}

// InfraLog is union of JournalLog and InfraContainerLog
type InfraLog struct {
	Docker              Docker           `json:"docker,omitempty"`
	Kubernetes          Kubernetes       `json:"kubernetes,omitempty"`
	Message             string           `json:"message"`
	Level               string           `json:"level"`
	Hostname            string           `json:"hostname"`
	PipelineMetadata    PipelineMetadata `json:"pipeline_metadata"`
	Timestamp           time.Time        `json:"@timestamp"`
	ViaqIndexName       string           `json:"viaq_index_name"`
	ViaqMsgID           string           `json:"viaq_msg_id"`
	STREAMID            string           `json:"_STREAM_ID,omitempty"`
	SYSTEMDINVOCATIONID string           `json:"_SYSTEMD_INVOCATION_ID,omitempty"`
	Systemd             Systemd          `json:"systemd,omitempty"`
	OpenshiftLabels     OpenshiftMeta    `json:"openshift"`
}

/*
Audit logs are
 - Audit logs generated by linux
 - Audit logs generated by kubernetes
 - Audit logs generated by openshift
*/

// LinuxAuditLog is generated by linux operating system
type LinuxAuditLog struct {
	Hostname         string           `json:"hostname"`
	AuditLinux       AuditLinux       `json:"audit.linux"`
	Message          string           `json:"message"`
	PipelineMetadata PipelineMetadata `json:"pipeline_metadata"`
	Timestamp        time.Time        `json:"@timestamp"`
	ViaqIndexName    string           `json:"viaq_index_name"`
	ViaqMsgID        string           `json:"viaq_msg_id"`
	Kubernetes       Kubernetes       `json:"kubernetes"`
	OpenshiftLabels  OpenshiftMeta    `json:"openshift"`
}

type AuditLinux struct {
	Type     string `json:"type"`
	RecordID string `json:"record_id"`
}

// AuditLogCommon is common to k8s and openshift auditlogs
type AuditLogCommon struct {
	Kind                     string           `json:"kind"`
	APIVersion               string           `json:"apiVersion"`
	Level                    string           `json:"level"`
	AuditID                  string           `json:"auditID"`
	Stage                    string           `json:"stage"`
	RequestURI               string           `json:"requestURI"`
	Verb                     string           `json:"verb"`
	User                     User             `json:"user"`
	SourceIPs                []string         `json:"sourceIPs"`
	UserAgent                string           `json:"userAgent"`
	ObjectRef                ObjectRef        `json:"objectRef"`
	ResponseStatus           ResponseStatus   `json:"responseStatus"`
	RequestReceivedTimestamp time.Time        `json:"requestReceivedTimestamp"`
	StageTimestamp           time.Time        `json:"stageTimestamp"`
	Annotations              Annotations      `json:"annotations"`
	Message                  interface{}      `json:"message"`
	Hostname                 string           `json:"hostname"`
	PipelineMetadata         PipelineMetadata `json:"pipeline_metadata"`
	Timestamp                time.Time        `json:"@timestamp"`
	ViaqIndexName            string           `json:"viaq_index_name"`
	ViaqMsgID                string           `json:"viaq_msg_id"`
	Kubernetes               Kubernetes       `json:"kubernetes"`
	OpenshiftLabels          OpenshiftMeta    `json:"openshift"`
}

// EventRouterLog is generated by event router
type EventRouterLog struct {
	Docker           Docker           `json:"docker"`
	Kubernetes       Kubernetes       `json:"kubernetes"`
	Message          string           `json:"message"`
	Level            string           `json:"level"`
	PipelineMetadata PipelineMetadata `json:"pipeline_metadata"`
	Timestamp        time.Time        `json:"@timestamp"`
	ViaqIndexName    string           `json:"viaq_index_name"`
	ViaqMsgID        string           `json:"viaq_msg_id"`
	OpenshiftLabels  OpenshiftMeta    `json:"openshift"`
}

type User struct {
	Username string   `json:"username"`
	UID      string   `json:"uid"`
	Groups   []string `json:"groups"`
}
type ObjectRef struct {
	Resource        string `json:"resource"`
	ResourceVersion string `json:"resourceVersion"`
	Name            string `json:"name"`
	Namespace       string `json:"namespace"`
	APIGroup        string `json:"apiGroup"`
	APIVersion      string `json:"apiVersion"`
	UID             string `json:"uid"`
}
type ResponseStatus struct {
	Code int `json:"code"`
}
type Annotations struct {
	AuthorizationK8SIoDecision string `json:"authorization.k8s.io/decision"`
	AuthorizationK8SIoReason   string `json:"authorization.k8s.io/reason"`
}

// OpenshiftAuditLog is audit log generated by openshift-apiserver
type OpenshiftAuditLog struct {
	AuditLogCommon
	OpenshiftAuditLevel string `json:"openshift_audit_level"`
}

// K8sAuditLog is audit logs generated by kube-apiserver
type K8sAuditLog struct {
	AuditLogCommon
	K8SAuditLevel string `json:"k8s_audit_level"`
}

// AuditLog is a union of LinuxAudit, K8sAudit, OpenshiftAudit logs
type AuditLog struct {
	Hostname                 string           `json:"hostname"`
	AuditLinux               AuditLinux       `json:"audit.linux"`
	Message                  string           `json:"message"`
	PipelineMetadata         PipelineMetadata `json:"pipeline_metadata"`
	Timestamp                time.Time        `json:"@timestamp"`
	ViaqIndexName            string           `json:"viaq_index_name"`
	ViaqMsgID                string           `json:"viaq_msg_id"`
	Kubernetes               Kubernetes       `json:"kubernetes"`
	Kind                     string           `json:"kind"`
	APIVersion               string           `json:"apiVersion"`
	Level                    string           `json:"level"`
	AuditID                  string           `json:"auditID"`
	Stage                    string           `json:"stage"`
	RequestURI               string           `json:"requestURI"`
	Verb                     string           `json:"verb"`
	User                     User             `json:"user"`
	SourceIPs                []string         `json:"sourceIPs"`
	UserAgent                string           `json:"userAgent"`
	ObjectRef                ObjectRef        `json:"objectRef"`
	ResponseStatus           ResponseStatus   `json:"responseStatus"`
	RequestReceivedTimestamp time.Time        `json:"requestReceivedTimestamp"`
	StageTimestamp           time.Time        `json:"stageTimestamp"`
	Annotations              Annotations      `json:"annotations"`
	K8SAuditLevel            string           `json:"k8s_audit_level"`
	OpenshiftAuditLevel      string           `json:"openshift_audit_level"`
}

// AllLog is a union of all log types
type AllLog struct {
	Docker                   Docker           `json:"docker"`
	Kubernetes               Kubernetes       `json:"kubernetes"`
	Message                  string           `json:"message"`
	Level                    string           `json:"level"`
	Hostname                 string           `json:"hostname"`
	PipelineMetadata         PipelineMetadata `json:"pipeline_metadata"`
	Timestamp                time.Time        `json:"@timestamp"`
	ViaqIndexName            string           `json:"viaq_index_name"`
	ViaqMsgID                string           `json:"viaq_msg_id"`
	STREAMID                 string           `json:"_STREAM_ID"`
	SYSTEMDINVOCATIONID      string           `json:"_SYSTEMD_INVOCATION_ID"`
	Systemd                  Systemd          `json:"systemd"`
	AuditLinux               AuditLinux       `json:"audit.linux"`
	Kind                     string           `json:"kind"`
	APIVersion               string           `json:"apiVersion"`
	AuditID                  string           `json:"auditID"`
	Stage                    string           `json:"stage"`
	RequestURI               string           `json:"requestURI"`
	Verb                     string           `json:"verb"`
	User                     User             `json:"user"`
	SourceIPs                []string         `json:"sourceIPs"`
	UserAgent                string           `json:"userAgent"`
	ObjectRef                ObjectRef        `json:"objectRef"`
	ResponseStatus           ResponseStatus   `json:"responseStatus"`
	RequestReceivedTimestamp time.Time        `json:"requestReceivedTimestamp"`
	StageTimestamp           time.Time        `json:"stageTimestamp"`
	Annotations              Annotations      `json:"annotations"`
	K8SAuditLevel            string           `json:"k8s_audit_level"`
	OpenshiftAuditLevel      string           `json:"openshift_audit_level"`
	OpenshiftLabels          OpenshiftMeta    `json:"openshift"`
}

func StrictlyParseLogs(in string, logs interface{}) error {
	logger.V(3).Info("ParseLogs", "content", in)
	if in == "" {
		return nil
	}
	dec := json.NewDecoder(bytes.NewBufferString(in))
	dec.DisallowUnknownFields()
	err := dec.Decode(&logs)
	if err != nil {
		return err
	}

	return nil
}

func ParseLogs(in string) (Logs, error) {
	logs := Logs{}
	if in == "" {
		return logs, nil
	}

	err := json.Unmarshal([]byte(in), &logs)
	if err != nil {
		return nil, err
	}

	return logs, nil
}

func (l Logs) ByIndex(prefix string) Logs {
	filtered := Logs{}
	for _, entry := range l {
		if strings.HasPrefix(entry.ViaqIndexName, prefix) {
			filtered = append(filtered, entry)
		}
	}
	return filtered
}

func (l Logs) ByPod(name string) Logs {
	filtered := Logs{}
	for _, entry := range l {
		if entry.Kubernetes.PodName == name {
			filtered = append(filtered, entry)
		}
	}
	return filtered
}

func (l Logs) NonEmpty() bool {
	if l == nil {
		return false
	}
	return len(l) > 0
}

func NewMockPod() *v1.Pod {
	uuidStr := string(utils.GetRandomWord(16))
	podName := string(utils.GetRandomWord(8))
	namespace := string(utils.GetRandomWord(8))
	fakePod := v1.Pod{
		TypeMeta: metav1.TypeMeta{
			Kind: "Pod",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      podName,
			Namespace: namespace,
			UID:       types.UID(uuidStr),
		},
		Spec: v1.PodSpec{},
	}
	return &fakePod
}

func NewMockEvent(ref *v1.ObjectReference, eventType, reason, message string) *v1.Event {
	tm := metav1.Time{
		Time: time.Now(),
	}
	return &v1.Event{
		ObjectMeta: metav1.ObjectMeta{
			Name:      string(utils.GetRandomWord(8)),
			Namespace: ref.Namespace,
		},
		InvolvedObject: *ref,
		Reason:         reason,
		Message:        message,
		FirstTimestamp: tm,
		LastTimestamp:  tm,
		Count:          rand.Int31(),
		Type:           eventType,
	}
}
