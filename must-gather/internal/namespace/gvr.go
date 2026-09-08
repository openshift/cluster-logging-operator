package namespace

import "k8s.io/apimachinery/pkg/runtime/schema"

const (
	groupApps       = "apps"
	groupK8sOvn     = "k8s.ovn.org"
	groupMonitoring = "monitoring.coreos.com"
	groupOperators  = "operators.coreos.com"
	groupRbac       = "rbac.authorization.k8s.io"

	v1       = "v1"
	v2       = "v2"
	v1Alpha1 = "v1alpha1"
)

var (
	nsGVR = schema.GroupVersionResource{Group: "", Version: v1, Resource: "namespaces"}

	defaultNamespacedResources = []schema.GroupVersionResource{
		// Core resources
		{Group: "", Version: v1, Resource: "pods"},
		{Group: "", Version: v1, Resource: "services"},
		{Group: "", Version: v1, Resource: "configmaps"},
		{Group: "", Version: v1, Resource: "secrets"},
		{Group: "", Version: v1, Resource: "serviceaccounts"},
		{Group: "", Version: v1, Resource: "events"},
		{Group: "", Version: v1, Resource: "endpoints"},
		{Group: "", Version: v1, Resource: "persistentvolumeclaims"},
		{Group: "", Version: v1, Resource: "replicationcontrollers"},

		// Apps
		{Group: groupApps, Version: v1, Resource: "deployments"},
		{Group: groupApps, Version: v1, Resource: "daemonsets"},
		{Group: groupApps, Version: v1, Resource: "statefulsets"},
		{Group: groupApps, Version: v1, Resource: "replicasets"},

		// RBAC
		{Group: groupRbac, Version: v1, Resource: "roles"},
		{Group: groupRbac, Version: v1, Resource: "rolebindings"},

		// Networking
		{Group: "networking.k8s.io", Version: v1, Resource: "networkpolicies"},
		{Group: "discovery.k8s.io", Version: v1, Resource: "endpointslices"},

		// OpenShift Routes
		{Group: "route.openshift.io", Version: v1, Resource: "routes"},

		// Autoscaling
		{Group: "autoscaling", Version: v2, Resource: "horizontalpodautoscalers"},

		// Policy
		{Group: "policy", Version: v1, Resource: "poddisruptionbudgets"},

		// OpenShift Monitoring
		{Group: groupMonitoring, Version: v1, Resource: "servicemonitors"},
		{Group: groupMonitoring, Version: v1, Resource: "podmonitors"},
		{Group: groupMonitoring, Version: v1, Resource: "prometheusrules"},

		// OVN Kubernetes
		{Group: groupK8sOvn, Version: v1, Resource: "egressfirewalls"},
		{Group: groupK8sOvn, Version: v1, Resource: "egressqoses"},
		{Group: groupK8sOvn, Version: v1, Resource: "userdefinednetworks"},

		// Operators
		{Group: groupOperators, Version: v1Alpha1, Resource: "installplans"},
		{Group: groupOperators, Version: v1Alpha1, Resource: "subscriptions"},
		{Group: groupOperators, Version: v1Alpha1, Resource: "clusterserviceversions"},
	}
)
