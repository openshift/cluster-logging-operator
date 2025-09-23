package otlp

import "sort"

const (
	// (DEPRECATED) Common Resource attributes
	LogSource               = "log_source"
	LogType                 = "log_type"
	KubernetesContainerName = "kubernetes.container_name"
	KubernetesHost          = "kubernetes.host"
	KubernetesPodName       = "kubernetes.pod_name"
	KubernetesNamespaceName = "kubernetes.namespace_name"
	OpenshiftClusterID      = "openshift.cluster_id"
	Level                   = "level"

	// Common Resource attributes
	OpenshiftClusterUID  = "openshift.cluster.uid"
	OpenshiftLogSource   = "openshift.log.source"
	OpenshiftLogType     = "openshift.log.type"
	OpenshiftLabelPrefix = "openshift.label."

	// Common to container, node(journal), auditd
	K8sNodeName = "k8s.node.name"

	// Container Resource Attributes
	K8sNamespaceName = "k8s.namespace.name"
	K8sContainerName = "k8s.container.name"
	K8sPodName       = "k8s.pod.name"
	K8sPodID         = "k8s.pod.uid"

	// Container LogRecord Resource Attributes
	LogIOStream = "log.iostream"

	// Journal Resource Attributes
	ProcessExeName     = "process.executable.name"
	ProcessExePath     = "process.executable.path"
	ProcessCommandLine = "process.command_line"
	ProcessPID         = "process.pid"
	ServiceName        = "service.name"

	// Journal LogRecord Attributes
	SystemdPrefix = "systemd."

	// Audit OVN LogRecord Attributes
	K8sOVNComponent = "k8s.ovn.component"

	// Auditd LogRecord Attributes
	AuditdType = "auditd.type"

	// Common LogRecord Attributes
	LogSequence = "log.sequence"
)

func (r Resource) Attribute(name string) AttributeValue {
	for _, a := range r.Attributes {
		if a.Key == name {
			return a.Value
		}
	}
	return AttributeValue{}
}

func (l LogRecord) Attribute(name string) AttributeValue {
	for _, a := range l.Attributes {
		if a.Key == name {
			return a.Value
		}
	}
	return AttributeValue{}
}

func CollectNames(attrs []Attribute) (names []string) {
	for _, a := range attrs {
		names = append(names, a.Key)
	}
	sort.Strings(names)
	return names
}
