package types

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"strings"
	"time"

	rand2 "github.com/openshift/cluster-logging-operator/test/helpers/rand"

	log "github.com/ViaQ/logerr/v2/log/static"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	"github.com/openshift/cluster-logging-operator/test"
)

var ErrParse = errors.New("logs could not be parsed")

const (
	OptionalString = "**optional**"
	AnyString      = "*"
)

// ContainerLog
type ContainerLog struct {
	ViaQCommon `json:",inline,omitempty"`

	// +optional
	// +deprecated
	Docker Docker `json:"docker,omitempty"`

	// The Kubernetes-specific metadata
	Kubernetes Kubernetes `json:"kubernetes,omitempty"`

	// Original log entry as a structured object.
	//
	//Example:
	// `{"pid":21631,"ppid":21618,"worker":0,"message":"starting fluentd worker pid=21631 ppid=21618 worker=0"}`
	//
	// This field may be present if the forwarder was configured to parse structured JSON logs.
	// If the original log entry was a valid structured log, this field will contain an equivalent JSON structure.
	// Otherwise this field will be empty or absent, and the `message` field will contain the original log message.
	// The `structured` field includes the same sub-fields as the original log message.
	// +optional
	Structured map[string]interface{} `json:"structured,omitempty"`
}

type Docker struct {

	// ContainerID is the id of the container producing the log
	ContainerID string `json:"container_id"`
}

type Kubernetes struct {

	// Annotations associated with the Kubernetes pod
	// +optional
	Annotations map[string]string `json:"annotations,omitempty"`

	// ContainerName of the the pod container that produced the log
	ContainerName string `json:"container_name,omitempty"`

	//NamespaceName where the pod is deployed
	NamespaceName string `json:"namespace_name,omitempty"`

	// PodName is the name of the pod
	PodName string `json:"pod_name,omitempty"`

	// +optional
	ContainerID string `json:"container_id,omitempty"`
	// +optional
	ContainerImage string `json:"container_image,omitempty"`
	// +optional
	ContainerImageID string `json:"container_image_id,omitempty"`

	//PodID is the unique uuid of the pod
	// +optional
	PodID string `json:"pod_id,omitempty"`

	// +docgen:ignore
	// +optional
	PodIP string `json:"pod_ip,omitempty"`

	//Host is the kubernetes node name that hosts the pod
	// +optional
	Host string `json:"host,omitempty"`

	//MasterURL is the url to the apiserver
	// +deprecated
	MasterURL string `json:"master_url,omitempty"`

	//NamespaceID is the unique uuid of the namespace
	// +optional
	NamespaceID string `json:"namespace_id,omitempty"`

	//FlatLabels is an array of the pod labels joined as key=value
	// +optional
	// +deprecated
	// +docgen:type=array
	FlatLabels []string `json:"flat_labels,omitempty"`

	// Labels present on the Pod at time the log was generated
	// +optional
	Labels map[string]string `json:"labels,omitempty"`

	// +docgen:ignore
	// +deprecated
	// +optional
	OrphanedNamespace string `json:"orphaned_namespace,omitempty"`

	// NamespaceLabels are the labels present on the pod namespace
	// +optional
	NamespaceLabels map[string]string `json:"namespace_labels,omitempty"`

	// The name of the stream the log line was submitted to (e.g.: stdout, stderr)
	// +optional
	ContainerStream string `json:"container_iostream,omitempty"`
}

type Collector struct {
	//Ipaddr4 is the ipV4 address of the collector
	//+optional
	Ipaddr4 string `json:"ipaddr4,omitempty"`

	//+deprecated
	Inputname string `json:"inputname,omitempty"`

	//Name is the implementation of the collector agent
	Name string `json:"name,omitempty"`

	//ReceivedAt the time the collector received the log entry
	ReceivedAt time.Time `json:"received_at,omitempty"`

	//Version is collector version information
	Version string `json:"version,omitempty"`

	//OriginalRawMessage captures the original message for eventrouter logs
	OriginalRawMessage string `json:"original_raw_message,omitempty"`
}

// PipelineMetadata is metadata related to ViaQ log collection pipeline. Everything about log collector, normalizers, mappings goes here.
// Data in this subgroup is forwarded for troubleshooting and tracing purposes.  This is only present when deploying
// fluentd collector implementations
// +deprecated
// +optional
type PipelineMetadata struct {

	//Collector metadata
	Collector Collector `json:"collector,omitempty"`
}

type OpenshiftMeta struct {

	//ClusterID is the unique id of the cluster where the workload is deployed
	ClusterID string `json:"cluster_id,omitempty"`

	//Labels is a set of common, static labels that were spec'd for log forwarding
	//to be sent with the log Records
	//+optional
	Labels map[string]string `json:"labels,omitempty"`

	//Sequence is increasing id used in conjunction with the timestamp to estblish a linear timeline
	//of log records.  This was added as a workaround for logstores that do not have nano-second precision.
	Sequence OptionalInt `json:"sequence,omitempty"`
}

// Application Logs are container logs from all namespaces except "openshift" and "openshift-*" namespaces
type ApplicationLog ContainerLog

// Infrastructure logs are
// - Journal logs
// - logs from "openshift" and "openshift-*" namespaces

// InfraContainerLog
// InfraContainerLog logs are container logs from "openshift" and "openshift-*" namespaces
type InfraContainerLog ContainerLog

type ViaQCommon struct {

	// A UTC value that marks when the log payload was created.
	//
	// If the creation time is not known when the log payload was first collected. The “@” prefix denotes a
	// field that is reserved for a particular use.
	//
	// format:
	//
	// * yyyy-MM-dd HH:mm:ss,SSSZ
	// * yyyy-MM-dd'T'HH:mm:ss.SSSSSSZ
	// * yyyy-MM-dd'T'HH:mm:ssZ
	// * dateOptionalTime
	//
	// example: `2015-01-24 14:06:05.071000000 Z`
	Timestamp time.Time `json:"@timestamp,omitempty"`

	// Original log entry text, UTF-8 encoded
	//
	// This field may be absent or empty if a non-empty `structured` field is present.
	// See the description of `structured` for additional details.
	// +optional
	Message string `json:"message,omitempty"`

	// The normalized log level
	//
	// The logging level from various sources, including `rsyslog(severitytext property)`, python's logging module, and others.
	//        The following values come from link:http://sourceware.org/git/?p=glibc.git;a=blob;f=misc/sys/syslog.h;h=ee01478c4b19a954426a96448577c5a76e6647c0;hb=HEAD#l74[`syslog.h`], and are preceded by their http://sourceware.org/git/?p=glibc.git;a=blob;f=misc/sys/syslog.h;h=ee01478c4b19a954426a96448577c5a76e6647c0;hb=HEAD#l51[numeric equivalents]:
	//
	//        * `0` = `emerg`, system is unusable.
	//        * `1` = `alert`, action must be taken immediately.
	//        * `2` = `crit`, critical conditions.
	//        * `3` = `err`, error conditions.
	//        * `4` = `warn`, warning conditions.
	//        * `5` = `notice`, normal but significant condition.
	//        * `6` = `info`, informational.
	//        * `7` = `debug`, debug-level messages.
	//        The two following values are not part of `syslog.h` but are widely used:
	//        * `8` = `trace`, trace-level messages, which are more verbose than `debug` messages.
	//        * `9` = `unknown`, when the logging system gets a value it doesn't recognize.
	//        Map the log levels or priorities of other logging systems to their nearest match in the preceding list. For example, from link:https://docs.python.org/2.7/library/logging.html#logging-levels[python logging], you can match `CRITICAL` with `crit`, `ERROR` with `err`, and so on.
	Level string `json:"level,omitempty"`

	// The name of the host where this log message originated. In a Kubernetes cluster, this is the same as `kubernetes.host`.
	Hostname string `json:"hostname,omitempty"`

	// Metadata related to ViaQ log collection pipeline. Everything about log collector, normalizers, mappings goes here.
	// Data in this subgroup is forwarded for troubleshooting and tracing purposes.  This is only present when deploying
	// fluentd collector implementations
	// +deprecated
	// +optional
	PipelineMetadata PipelineMetadata `json:"pipeline_metadata,omitempty"`

	// LogSource is the source of a log used along with the LogType to distinguish a subcategory of the LogType.
	// Application logs are always sourced from containers
	// Infrastructure logs are sourced from containers or journal logs from the node
	// Audit logs are sourced from: kubernetes and openshift API servers, node auditd, and OVN
	LogSource string `json:"log_source,omitempty"`

	//The source type of the log. The `log_type` field may contain one of these strings, or may have additional dot-separated components, for example "infrastructure.container" or "infrastructure.node".
	//
	// * "application": Container logs generated by user applications running in the cluster, except infrastructure containers.
	// * "infrastructure": Node logs (such as syslog or journal logs), and container logs from pods in the openshift*, kube*, or default projects.
	// * "audit":
	// ** Node logs from auditd (/var/log/audit/audit.log)
	// ** Kubernetes and OpenShift apiservers audit logs.
	// ** OVN audit logs
	//
	LogType string `json:"log_type,omitempty"`

	// ViaqIndexName used with Elasticsearch 6.x and later, this is a name of a write index alias (e.g. app-write).
	//
	// The value depends on the log type of this message. Detailed documentation is found at https://github.com/openshift/enhancements/blob/master/enhancements/cluster-logging/cluster-logging-es-rollover-data-design.md#data-model.
	// +optional
	ViaqIndexName string `json:"viaq_index_name,omitempty"`

	// ViaqMessageId is a unique ID assigned to each message. The format is not specified.
	//
	// It may be a UUID or a Base64 (e.g. 82f13a8e-882a-4344-b103-f0a6f30fd218),
	// or some other ASCII value and is used as the `_id` of the document when sending to Elasticsearch. The intended use of this field is that if you use another
	// logging store or application other than Elasticsearch, but you still need to correlate data with the data stored
	// in Elasticsearch, this field will give you the exact document corresponding to the record.
	//
	// This is only present when deploying fluentd collector implementations
	// +optional
	ViaqMsgID string `json:"viaq_msg_id,omitempty"`

	// Openshift specific metadata
	Openshift OpenshiftMeta `json:"openshift,omitempty"`
}

// JournalLog is linux journal logs
type JournalLog struct {
	ViaQCommon          `json:",inline,omitempty"`
	STREAMID            string  `json:"_STREAM_ID,omitempty"`
	SYSTEMDINVOCATIONID string  `json:"_SYSTEMD_INVOCATION_ID,omitempty"`
	Systemd             Systemd `json:"systemd,omitempty"`
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
	LogType             string           `json:"log_type,omitempty"`
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
 - Audit logs generated by Openshift virtual network
*/

// LinuxAuditLog is generated by linux operating system
type LinuxAuditLog struct {
	Hostname         string           `json:"hostname"`
	AuditLinux       AuditLinux       `json:"audit.linux"`
	Message          string           `json:"message,omitempty"`
	PipelineMetadata PipelineMetadata `json:"pipeline_metadata"`
	Timestamp        time.Time        `json:"@timestamp"`
	LogSource        string           `json:"log_source,omitempty"`
	LogType          string           `json:"log_type,omitempty"`
	ViaqIndexName    string           `json:"viaq_index_name"`
	ViaqMsgID        string           `json:"viaq_msg_id"`
	Kubernetes       Kubernetes       `json:"kubernetes"`
	Openshift        OpenshiftMeta    `json:"openshift"`
	Timing           `json:",inline"`
	Level            string `json:"level,omitempty"`
}

type AuditLinux struct {
	Type     string `json:"type,omitempty"`
	RecordID string `json:"record_id,omitempty"`
}

// OVN Audit log
type OVNAuditLog struct {
	Hostname         string           `json:"hostname"`
	Message          string           `json:"message,omitempty"`
	PipelineMetadata PipelineMetadata `json:"pipeline_metadata"`
	Timestamp        time.Time        `json:"@timestamp"`
	LogType          string           `json:"log_type,omitempty"`
	LogSource        string           `json:"log_source,omitempty"`
	ViaqIndexName    string           `json:"viaq_index_name"`
	ViaqMsgID        string           `json:"viaq_msg_id"`
	Kubernetes       Kubernetes       `json:"kubernetes"`
	Openshift        OpenshiftMeta    `json:"openshift"`
	Level            string           `json:"level,omitempty"`
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
	LogSource                string           `json:"log_source,omitempty"`
	LogType                  string           `json:"log_type,omitempty"`
	ViaqIndexName            string           `json:"viaq_index_name,omitempty"`
	ViaqMsgID                string           `json:"viaq_msg_id,omitempty"`
	Kubernetes               Kubernetes       `json:"kubernetes,omitempty"`
	OpenshiftLabels          OpenshiftMeta    `json:"openshift,omitempty"`
	Timing                   `json:",inline"`
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
	Docker                   Docker           `json:"docker,omitempty"`
	LogType                  string           `json:"log_type,omitempty"`
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
	ContainerLog             `json:",inline,omitempty"`
	STREAMID                 string         `json:"_STREAM_ID,omitempty"`
	SYSTEMDINVOCATIONID      string         `json:"_SYSTEMD_INVOCATION_ID,omitempty"`
	Systemd                  Systemd        `json:"systemd,omitempty"`
	AuditLinux               AuditLinux     `json:"audit.linux,omitempty"`
	Kind                     string         `json:"kind,omitempty"`
	APIVersion               string         `json:"apiVersion,omitempty"`
	AuditID                  string         `json:"auditID,omitempty"`
	Stage                    string         `json:"stage,omitempty"`
	RequestURI               string         `json:"requestURI,omitempty"`
	Verb                     string         `json:"verb,omitempty"`
	User                     User           `json:"user,omitempty"`
	SourceIPs                []string       `json:"sourceIPs,omitempty"`
	UserAgent                string         `json:"userAgent,omitempty"`
	ObjectRef                ObjectRef      `json:"objectRef,omitempty"`
	ResponseStatus           ResponseStatus `json:"responseStatus,omitempty"`
	RequestReceivedTimestamp time.Time      `json:"requestReceivedTimestamp,omitempty"`
	StageTimestamp           time.Time      `json:"stageTimestamp,omitempty"`
	Annotations              Annotations    `json:"annotations,omitempty"`
	K8SAuditLevel            string         `json:"k8s_audit_level,omitempty"`
	OpenshiftAuditLevel      string         `json:"openshift_audit_level,omitempty"`
}

func (l AllLog) StreamName() string {
	if l.Kubernetes.NamespaceName != "" {
		return fmt.Sprintf("%s_%s_%s", l.Kubernetes.NamespaceName, l.Kubernetes.PodName, l.Kubernetes.ContainerName)
	}
	return l.STREAMID
}

func StrictlyParseLogsFromSlice(in []string, logs interface{}) error {
	jsonString := fmt.Sprintf("[%s]", strings.Join(in, ","))
	return StrictlyParseLogs(jsonString, logs)
}

func StrictlyParseLogs(in string, logs interface{}) error {
	return ParseLogsFrom(in, logs, true)
}

func ParseLogsFrom(in string, logs interface{}, strict bool) error {
	log.V(3).Info("ParseLogs", "content", in)
	if in == "" {
		return nil
	}
	dec := json.NewDecoder(bytes.NewBufferString(in))
	if strict {
		dec.DisallowUnknownFields()
	}
	err := dec.Decode(&logs)
	if err != nil {
		log.V(1).Error(err, "Error decoding", "log", in)
		return err
	}

	return nil
}

type Logs []AllLog

type Timing struct {
	// EpocIn is only added during benchmark testing
	EpocIn float64 `json:"epoc_in,omitempty"`
	// EpocOut is only added during benchmark testing
	EpocOut float64 `json:"epoc_out,omitempty"`
}

// Bloat is the ratio of overall size / Message size
func (l *AllLog) Bloat() float64 {
	return float64(len(l.String())) / float64(len(l.Message))
}

func (l *AllLog) String() string {
	return test.JSONLine(l)
}

func ParseJournalLogs[t any](in string) ([]t, error) {
	logs := []t{}
	if in == "" {
		return logs, nil
	}

	err := json.Unmarshal([]byte(in), &logs)
	if err != nil {
		return nil, err
	}

	return logs, nil
}

// TODO: replace with generic
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
	uuidStr := string(rand2.Word(16))
	podName := string(rand2.Word(8))
	namespace := string(rand2.Word(8))
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

func NewEvent(ref *v1.ObjectReference, eventType, reason, message string) *v1.Event {
	tm := metav1.Time{
		Time: time.Time{},
	}
	return &v1.Event{
		ObjectMeta: metav1.ObjectMeta{
			Name:      ref.Name,
			Namespace: ref.Namespace,
		},
		InvolvedObject: *ref,
		Reason:         reason,
		Message:        message,
		FirstTimestamp: tm,
		LastTimestamp:  tm,
		Count:          rand.Int31(), // nolint:gosec
		Type:           eventType,
	}
}
