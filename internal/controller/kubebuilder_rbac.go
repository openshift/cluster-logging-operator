package controller

// This file collects all the "kubebuilder rbac annotations" that the controllers contained
// in this operator need to function.

// +kubebuilder:rbac:groups=apps,resources=deployments;daemonsets,verbs=get;list;watch;create;update;delete
// +kubebuilder:rbac:groups=authorization.k8s.io,resources=subjectaccessreviews,verbs=create
// +kubebuilder:rbac:groups=config.openshift.io,resources=proxies;infrastructures,verbs=get;list;watch
// +kubebuilder:rbac:groups=core,resources=configmaps;services,verbs=get;list;watch;create;update;delete
// +kubebuilder:rbac:groups=core,resources=namespaces,verbs=get;list;watch
// +kubebuilder:rbac:groups=core,resources=pods;nodes,verbs=get;list;watch
// +kubebuilder:rbac:groups=core,resources=serviceaccounts;serviceaccounts/finalizers,verbs=get;list;watch
// +kubebuilder:rbac:groups=core,resources=serviceaccounts;serviceaccounts/finalizers,verbs=create,namespace=openshift-logging
// +kubebuilder:rbac:groups=core,resources=serviceaccounts;serviceaccounts/finalizers,verbs=update,resourceNames=logfilesmetricexporter,namespace=openshift-logging
// +kubebuilder:rbac:groups=core,resources=secrets,verbs=get;list;watch;create
// +kubebuilder:rbac:groups=core,namespaces,verbs=get;list;watch
// +kubebuilder:rbac:groups=logging.openshift.io,resources=logfilemetricexporters,verbs=get;list;watch
// +kubebuilder:rbac:groups=logging.openshift.io,resources=logfilemetricexporters/status,verbs=update
// +kubebuilder:rbac:groups=monitoring.coreos.com,resources=servicemonitors,verbs=get;list;watch;create;update
// +kubebuilder:rbac:groups=networking.k8s.io,resources=networkpolicies,verbs=get;list;watch;create;update;delete
// +kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=clusterrolebindings,verbs=get;list;watch;create;update;delete
// +kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=roles;rolebindings,verbs=get;list;watch;create;update;delete
// +kubebuilder:rbac:groups=scheduling.k8s.io,resources=priorityclasses,verbs=delete
// +kubebuilder:rbac:groups=security.openshift.io,resources=securitycontextconstraints,verbs=create
// +kubebuilder:rbac:groups=security.openshift.io,resources=securitycontextconstraints,resourceNames=logging-scc,verbs=get;update;use

// +kubebuilder:rbac:groups=config.openshift.io,resources=clusterversions,verbs=get;list;watch

// +kubebuilder:rbac:groups=observability.openshift.io,resources=clusterlogforwarders,verbs=get;list;watch
// +kubebuilder:rbac:groups=observability.openshift.io,resources=clusterlogforwarders/status,verbs=get;update;patch
