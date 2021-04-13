package types

import (
	"bytes"
	"encoding/json"
	"errors"
	"math/rand"
	"strings"
	"time"

	logger "github.com/ViaQ/logerr/log"
	"github.com/openshift/cluster-logging-operator/pkg/utils"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	"github.com/ViaQ/logerr/log"

	"github.com/openshift/cluster-logging-operator/test"
)

var ErrParse = errors.New("logs could not be parsed")

// ContainerLog
type ContainerLog struct {
	Timestamp        time.Time              `json:"@timestamp"`
	Docker           Docker                 `json:"docker"`
	Kubernetes       Kubernetes             `json:"kubernetes"`
	Message          string                 `json:"message"`
	Level            string                 `json:"level"`
	Hostname         string                 `json:"hostname"`
	PipelineMetadata PipelineMetadata       `json:"pipeline_metadata"`
	ViaqIndexName    string                 `json:"viaq_index_name"`
	ViaqMsgID        string                 `json:"viaq_msg_id"`
	OpenshiftLabels  OpenshiftMeta          `json:"openshift"`
	Structured       map[string]interface{} `json:"structured"`
}

type Docker struct {
	ContainerID string `json:"container_id"`
}

type Kubernetes struct {
	ContainerName     string            `json:"container_name,omitempty"`
	NamespaceName     string            `json:"namespace_name,omitempty"`
	PodName           string            `json:"pod_name,omitempty"`
	ContainerImage    string            `json:"container_image,omitempty"`
	ContainerImageID  string            `json:"container_image_id,omitempty"`
	PodID             string            `json:"pod_id,omitempty"`
	Host              string            `json:"host,omitempty"`
	MasterURL         string            `json:"master_url,omitempty"`
	NamespaceID       string            `json:"namespace_id,omitempty"`
	FlatLabels        []string          `json:"flat_labels,omitempty"`
	Labels            map[string]string `json:"labels,omitempty"`
	OrphanedNamespace string            `json:"orphaned_namespace,omitempty"`
	NamespaceLabels   map[string]string `json:"namespace_labels,omitempty"`
}

type Collector struct {
	Ipaddr4    string    `json:"ipaddr4,omitempty"`
	Inputname  string    `json:"inputname,omitempty"`
	Name       string    `json:"name,omitempty"`
	ReceivedAt time.Time `json:"received_at,omitempty"`
	Version    string    `json:"version,omitempty"`
}

// EventData encodes an eventrouter event and previous event, with a verb for
// whether the event is created or updated.
type EventData struct {
	Verb     string    `json:"verb"`
	Event    *v1.Event `json:"event"`
	OldEvent *v1.Event `json:"old_event,omitempty"`
}

type PipelineMetadata struct {
	Collector Collector `json:"collector,omitempty"`
}

type OpenshiftMeta struct {
	Labels map[string]string `json:"labels,omitempty"`
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
	STREAMID            string           `json:"_STREAM_ID,omitempty"`
	SYSTEMDINVOCATIONID string           `json:"_SYSTEMD_INVOCATION_ID,omitempty"`
	Systemd             Systemd          `json:"systemd,omitempty"`
	Level               string           `json:"level,omitempty"`
	Message             string           `json:"message,omitempty"`
	Hostname            string           `json:"hostname,omitempty"`
	PipelineMetadata    PipelineMetadata `json:"pipeline_metadata,omitempty"`
	Timestamp           time.Time        `json:"@timestamp,omitempty"`
	ViaqIndexName       string           `json:"viaq_index_name,omitempty"`
	ViaqMsgID           string           `json:"viaq_msg_id,omitempty"`
	Kubernetes          Kubernetes       `json:"kubernetes,omitempty"`
}

type T struct {
	BOOTID              string `json:"BOOT_ID,omitempty"`
	CAPEFFECTIVE        string `json:"CAP_EFFECTIVE,omitempty"`
	CMDLINE             string `json:"CMDLINE,omitempty"`
	COMM                string `json:"COMM,omitempty"`
	EXE                 string `json:"EXE,omitempty"`
	GID                 string `json:"GID,omitempty"`
	MACHINEID           string `json:"MACHINE_ID,omitempty"`
	PID                 string `json:"PID,omitempty"`
	SELINUXCONTEXT      string `json:"SELINUX_CONTEXT,omitempty"`
	STREAMID            string `json:"STREAM_ID,omitempty"`
	SYSTEMDCGROUP       string `json:"SYSTEMD_CGROUP,omitempty"`
	SYSTEMDINVOCATIONID string `json:"SYSTEMD_INVOCATION_ID,omitempty"`
	SYSTEMDSLICE        string `json:"SYSTEMD_SLICE,omitempty"`
	SYSTEMDUNIT         string `json:"SYSTEMD_UNIT,omitempty"`
	TRANSPORT           string `json:"TRANSPORT,omitempty"`
	UID                 string `json:"UID,omitempty"`
}

type U struct {
	SYSLOGIDENTIFIER string `json:"SYSLOG_IDENTIFIER,omitempty"`
}

type Systemd struct {
	T T `json:"t,omitempty"`
	U U `json:"u,omitempty"`
}

// InfraLog is union of JournalLog and InfraContainerLog
type InfraLog struct {
	Docker              Docker           `json:"docker,omitempty"`
	Kubernetes          Kubernetes       `json:"kubernetes,omitempty"`
	Message             string           `json:"message,omitempty"`
	Level               string           `json:"level,omitempty"`
	Hostname            string           `json:"hostname,omitempty"`
	PipelineMetadata    PipelineMetadata `json:"pipeline_metadata,omitempty"`
	Timestamp           time.Time        `json:"@timestamp,omitempty"`
	ViaqIndexName       string           `json:"viaq_index_name,omitempty"`
	ViaqMsgID           string           `json:"viaq_msg_id,omitempty"`
	STREAMID            string           `json:"_STREAM_ID,omitempty"`
	SYSTEMDINVOCATIONID string           `json:"_SYSTEMD_INVOCATION_ID,omitempty"`
	Systemd             Systemd          `json:"systemd,omitempty"`
	OpenshiftLabels     OpenshiftMeta    `json:"openshift,omitempty"`
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
	Timing           `json:",inline"`
	Level            string `json:"level,omitempty"`
}

type AuditLinux struct {
	Type     string `json:"type,omitempty"`
	RecordID string `json:"record_id,omitempty"`
}

// AuditLogCommon is common to k8s and openshift auditlogs
type AuditLogCommon struct {
	Kind                     string           `json:"kind,omitempty"`
	APIVersion               string           `json:"apiVersion,omitempty"`
	Level                    string           `json:"level,omitempty"`
	AuditID                  string           `json:"auditID,omitempty"`
	Stage                    string           `json:"stage,omitempty"`
	RequestURI               string           `json:"requestURI,omitempty"`
	Verb                     string           `json:"verb,omitempty"`
	User                     User             `json:"user,omitempty"`
	SourceIPs                []string         `json:"sourceIPs,omitempty"`
	UserAgent                string           `json:"userAgent,omitempty"`
	ObjectRef                ObjectRef        `json:"objectRef,omitempty"`
	ResponseStatus           ResponseStatus   `json:"responseStatus,omitempty"`
	RequestReceivedTimestamp time.Time        `json:"requestReceivedTimestamp,omitempty"`
	StageTimestamp           time.Time        `json:"stageTimestamp,omitempty"`
	Annotations              Annotations      `json:"annotations,omitempty"`
	Message                  interface{}      `json:"message,omitempty"`
	Hostname                 string           `json:"hostname,omitempty"`
	PipelineMetadata         PipelineMetadata `json:"pipeline_metadata,omitempty"`
	Timestamp                time.Time        `json:"@timestamp,omitempty"`
	ViaqIndexName            string           `json:"viaq_index_name,omitempty"`
	ViaqMsgID                string           `json:"viaq_msg_id,omitempty"`
	Kubernetes               Kubernetes       `json:"kubernetes,omitempty"`
	OpenshiftLabels          OpenshiftMeta    `json:"openshift,omitempty"`
	Timing                   `json:",inline"`
}

// EventRouterLog is generated by event router
type EventRouterLog struct {
	Docker           Docker           `json:"docker"`
	Kubernetes       Kubernetes       `json:"kubernetes"`
	Message          string           `json:"message"`
	Level            string           `json:"level"`
	Hostname         string           `json:"hostname,omitempty"`
	PipelineMetadata PipelineMetadata `json:"pipeline_metadata"`
	Timestamp        time.Time        `json:"@timestamp"`
	ViaqIndexName    string           `json:"viaq_index_name"`
	ViaqMsgID        string           `json:"viaq_msg_id"`
	OpenshiftLabels  OpenshiftMeta    `json:"openshift"`
	Timing           `json:",inline"`
}

type User struct {
	Username string   `json:"username,omitempty"`
	UID      string   `json:"uid,omitempty"`
	Groups   []string `json:"groups,omitempty"`
}
type ObjectRef struct {
	Resource        string `json:"resource,omitempty"`
	ResourceVersion string `json:"resourceVersion,omitempty"`
	Name            string `json:"name,omitempty"`
	Namespace       string `json:"namespace,omitempty"`
	APIGroup        string `json:"apiGroup,omitempty"`
	APIVersion      string `json:"apiVersion,omitempty"`
	UID             string `json:"uid,omitempty"`
}
type ResponseStatus struct {
	Code int `json:"code,omitempty"`
}
type Annotations struct {
	AuthorizationK8SIoDecision string `json:"authorization.k8s.io/decision"`
	AuthorizationK8SIoReason   string `json:"authorization.k8s.io/reason"`
}

// OpenshiftAuditLog is audit log generated by openshift-apiserver
type OpenshiftAuditLog struct {
	AuditLogCommon
	OpenshiftAuditLevel string `json:"openshift_audit_level,omitempty"`
}

// K8sAuditLog is audit logs generated by kube-apiserver
type K8sAuditLog struct {
	AuditLogCommon
	K8SAuditLevel string `json:"k8s_audit_level,omitempty"`
}

// AuditLog is a union of LinuxAudit, K8sAudit, OpenshiftAudit logs
type AuditLog struct {
	Hostname                 string           `json:"hostname,omitempty"`
	AuditLinux               AuditLinux       `json:"audit.linux,omitempty"`
	Message                  string           `json:"message,omitempty"`
	PipelineMetadata         PipelineMetadata `json:"pipeline_metadata"`
	Timestamp                time.Time        `json:"@timestamp,omitempty"`
	ViaqIndexName            string           `json:"viaq_index_name,omitempty"`
	ViaqMsgID                string           `json:"viaq_msg_id,omitempty"`
	Kubernetes               Kubernetes       `json:"kubernetes,omitempty"`
	Kind                     string           `json:"kind,omitempty"`
	APIVersion               string           `json:"apiVersion,omitempty"`
	Level                    string           `json:"level,omitempty"`
	AuditID                  string           `json:"auditID,omitempty"`
	Stage                    string           `json:"stage,omitempty"`
	RequestURI               string           `json:"requestURI,omitempty"`
	Verb                     string           `json:"verb,omitempty"`
	User                     User             `json:"user,omitempty"`
	SourceIPs                []string         `json:"sourceIPs,omitempty"`
	UserAgent                string           `json:"userAgent,omitempty"`
	ObjectRef                ObjectRef        `json:"objectRef,omitempty"`
	ResponseStatus           ResponseStatus   `json:"responseStatus,omitempty"`
	RequestReceivedTimestamp time.Time        `json:"requestReceivedTimestamp,omitempty"`
	StageTimestamp           time.Time        `json:"stageTimestamp,omitempty"`
	Annotations              Annotations      `json:"annotations,omitempty"`
	K8SAuditLevel            string           `json:"k8s_audit_level,omitempty"`
	OpenshiftAuditLevel      string           `json:"openshift_audit_level,omitempty"`
}

// AllLog is a union of all log types
type AllLog struct {
	Docker                   Docker           `json:"docker,omitempty"`
	Kubernetes               Kubernetes       `json:"kubernetes,omitempty"`
	Message                  string           `json:"message,omitempty"`
	Level                    string           `json:"level,omitempty"`
	Hostname                 string           `json:"hostname,omitempty"`
	PipelineMetadata         PipelineMetadata `json:"pipeline_metadata,omitempty"`
	Timestamp                time.Time        `json:"@timestamp,omitempty"`
	ViaqIndexName            string           `json:"viaq_index_name,omitempty"`
	ViaqMsgID                string           `json:"viaq_msg_id,omitempty"`
	STREAMID                 string           `json:"_STREAM_ID,omitempty"`
	SYSTEMDINVOCATIONID      string           `json:"_SYSTEMD_INVOCATION_ID,omitempty"`
	Systemd                  Systemd          `json:"systemd,omitempty"`
	AuditLinux               AuditLinux       `json:"audit.linux,omitempty"`
	Kind                     string           `json:"kind,omitempty"`
	APIVersion               string           `json:"apiVersion,omitempty"`
	AuditID                  string           `json:"auditID,omitempty"`
	Stage                    string           `json:"stage,omitempty"`
	RequestURI               string           `json:"requestURI,omitempty"`
	Verb                     string           `json:"verb,omitempty"`
	User                     User             `json:"user,omitempty"`
	SourceIPs                []string         `json:"sourceIPs,omitempty"`
	UserAgent                string           `json:"userAgent,omitempty"`
	ObjectRef                ObjectRef        `json:"objectRef,omitempty"`
	ResponseStatus           ResponseStatus   `json:"responseStatus,omitempty"`
	RequestReceivedTimestamp time.Time        `json:"requestReceivedTimestamp,omitempty"`
	StageTimestamp           time.Time        `json:"stageTimestamp,omitempty"`
	Annotations              Annotations      `json:"annotations,omitempty"`
	K8SAuditLevel            string           `json:"k8s_audit_level,omitempty"`
	OpenshiftAuditLevel      string           `json:"openshift_audit_level,omitempty"`
	OpenshiftLabels          OpenshiftMeta    `json:"openshift,omitempty"`
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
		log.V(1).Error(err, "Error decoding", "log", in)
		return err
	}

	return nil
}

type Logs []AllLog

type PerfLog struct {
	AllLog
	Timing `json:",inline"`
}

type Timing struct {
	//EpocIn is only added during benchmark testing
	EpocIn float64 `json:"epoc_in,omitempty"`
	//EpocOut is only added during benchmark testing
	EpocOut float64 `json:"epoc_out,omitempty"`
}

type PerfLogs []PerfLog

func (t *PerfLog) ElapsedEpoc() float64 {
	return t.EpocOut - t.EpocIn
}

//Bloat is the ratio of overall size / Message size
func (l *AllLog) Bloat() float64 {
	return float64(len(l.String())) / float64(len(l.Message))
}

func (l *AllLog) String() string {
	return test.JSONLine(l)
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
