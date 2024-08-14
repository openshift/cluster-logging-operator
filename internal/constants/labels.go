package constants

const (

	// K8s recommended label names: https://kubernetes.io/docs/concepts/overview/working-with-objects/common-labels/

	LabelK8sName      = "app.kubernetes.io/name"       // The name of the application (string)
	LabelK8sInstance  = "app.kubernetes.io/instance"   // A unique name identifying the instance of an application (string)
	LabelK8sVersion   = "app.kubernetes.io/version"    // The current version of the application (e.g., a semantic version, revision hash, etc.) (string)
	LabelK8sComponent = "app.kubernetes.io/component"  // The component within the architecture (string)
	LabelK8sPartOf    = "app.kubernetes.io/part-of"    // The name of a higher level application this one is part of (string)
	LabelK8sManagedBy = "app.kubernetes.io/managed-by" // The tool being used to manage the operation of an application (string)

	LabelLoggingServiceType      = "logging.observability.openshift.io/service-type"
	LabelLoggingInputServiceType = "logging.observability.openshift.io/input-service-type"

	ServiceTypeMetrics = "metrics"
	ServiceTypeInput   = "input"
)
