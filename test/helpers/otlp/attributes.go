package otlp

const (
	NodeName           = "k8s.node.name"
	ClusterID          = "k8s.cluster.uid"
	K8sNamespaceName   = "k8s.namespace.name"
	K8sPodName         = "k8s.pod.name"
	K8sContainerName   = "k8s.container.name"
	OpenshiftLogSource = "openshift.log.source"
	OpenshiftLogType   = "openshift.log.type"
)

func (r Resource) NodeNameAttribute() string {
	return NodeName
}
func (r Resource) ClusterIDAttribute() string {
	return ClusterID
}
func (r Resource) NamespaceNameAttribute() string {
	return K8sNamespaceName
}
func (r Resource) PodNameAttribute() string {
	return K8sPodName
}
func (r Resource) ContainerNameAttribute() string {
	return K8sContainerName
}
func (r Resource) LogSourceAttribute() string {
	return OpenshiftLogSource
}
func (l LogRecord) LogTypeAttribute() string {
	return OpenshiftLogType
}

func (r Resource) FindStringValue(key string) string {
	for _, attr := range r.Attributes {
		if attr.Key == key {
			return attr.Value.StringValue
		}
	}
	return ""
}

func (l LogRecord) FindStringValue(key string) string {
	for _, attr := range l.Attributes {
		if attr.Key == key {
			return attr.Value.StringValue
		}
	}
	return ""
}
