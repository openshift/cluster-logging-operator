package controller

// This file collects all the "kubebuilder rbac annotations" that the controllers contained
// in this operator need to function.

// +kubebuilder:rbac:groups=apps,resources=deployments;daemonsets,verbs=*
// +kubebuilder:rbac:groups=authorization.k8s.io,resources=subjectaccessreviews,verbs=create
// +kubebuilder:rbac:groups=config.openshift.io,resources=proxies;infrastructures,verbs=get;list;watch
// +kubebuilder:rbac:groups=core,resources=pods;services;events;configmaps;secrets;serviceaccounts;serviceaccounts/finalizers;services/finalizers;namespaces,verbs=*
// +kubebuilder:rbac:groups=core,namespaces,verbs=get;list;watch
// +kubebuilder:rbac:groups=logging.openshift.io,resources=*,verbs=*
// +kubebuilder:rbac:groups=monitoring.coreos.com,resources=prometheusrules;servicemonitors,verbs=*
// +kubebuilder:rbac:groups=networking.k8s.io,resources=networkpolicies,verbs=create;delete
// +kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=clusterroles;clusterrolebindings;roles;rolebindings,verbs=*
// +kubebuilder:rbac:groups=scheduling.k8s.io,resources=priorityclasses,verbs=delete
// +kubebuilder:rbac:groups=security.openshift.io,resources=securitycontextconstraints,verbs=create;use;get;list;watch

// +kubebuilder:rbac:groups=config.openshift.io,resources=clusterversions,verbs=get;list;watch

// +kubebuilder:rbac:groups=observability.openshift.io,resources=clusterlogforwarders,verbs=*
// +kubebuilder:rbac:groups=observability.openshift.io,resources=clusterlogforwarders/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=observability.openshift.io,resources=clusterlogforwarders/finalizers,verbs=update
