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

	// Common Resource attributes
	OpenshiftClusterUID  = "openshift.cluster.uid"
	OpenshiftLogSource   = "openshift.log.source"
	OpenshiftLogType     = "openshift.log.type"
	OpenshiftLabelPrefix = "openshift.label."

	// Common to container, node(journal), auditd
	NodeName = "k8s.node.name"
	Level    = "level"

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
	SystemTPrefix = "systemd.t."
	SystemUPrefix = "systemd.u."

	// Audit (Kubernetes Events) LogRecord Attributes
	K8sEventLevel                           = "k8s.audit.event.level"
	K8sEventStage                           = "k8s.audit.event.stage"
	K8sEventUserAgent                       = "k8s.audit.event.user_agent"
	K8sEventRequestURI                      = "k8s.audit.event.request.uri"
	K8sEventResponseCode                    = "k8s.audit.event.response.code"
	K8sEventAnnotationPrefix                = "k8s.audit.event.annotation."
	K8sEventObjectRefResource               = "k8s.audit.event.object_ref.resource"
	K8sEventObjectRefName                   = "k8s.audit.event.object_ref.name"
	K8sEventObjectRefNamespace              = "k8s.audit.event.object_ref.namespace"
	K8sEventObjectRefAPIGroup               = "k8s.audit.event.object_ref.api_group"
	K8sEventObjectRefAPIVersion             = "k8s.audit.event.object_ref.api_version"
	K8sUserUsername                         = "k8s.user.username"
	K8sUserGroups                           = "k8s.user.groups"
	K8sEventAnnotationAuthorizationDecision = "k8s.audit.event.annotation.authorization.k8s.io/decision"
	K8sEventAnnotationAuthorizationReason   = "k8s.audit.event.annotation.authorization.k8s.io/reason"

	// Audit OVN LogRecord Attributes
	K8sOVNSequence  = "k8s.ovn.sequence"
	K8sOVNComponent = "k8s.ovn.component"

	// Auditd LogRecord Attributes
	AuditdSequence = "auditd.sequence"
	AuditdType     = "auditd.type"
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
