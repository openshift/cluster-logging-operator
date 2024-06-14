package controller

// This file collects all the "kubebuilder rbac annotations" that the controllers contained
// in this operator need to function.

// +kubebuilder:rbac:groups=apps,resources=deployments;daemonsets;replicasets;statefulsets,verbs=*
// +kubebuilder:rbac:groups=authentication.k8s.io,resources=tokenreviews,verbs=create
// +kubebuilder:rbac:groups=authorization.k8s.io,resources=subjectaccessreviews,verbs=create
// +kubebuilder:rbac:groups=batch,resources=cronjobs,verbs=*
// +kubebuilder:rbac:groups=config.openshift.io,resources=proxies;infrastructures,verbs=get;list;watch
// +kubebuilder:rbac:groups=console.openshift.io,resources=consolelinks;consoleexternalloglinks;consoleplugins;consoleplugins/finalizers,verbs=get;create;update;delete
// +kubebuilder:rbac:groups=core,resources=pods;pods/exec;services;endpoints;persistentvolumeclaims;events;configmaps;secrets;serviceaccounts;serviceaccounts/finalizers;services/finalizers;namespaces,verbs=*
// +kubebuilder:rbac:groups=logging.openshift.io,resources=*,verbs=*
// +kubebuilder:rbac:groups=monitoring.coreos.com,resources=prometheusrules;servicemonitors,verbs=*
// +kubebuilder:rbac:groups=networking.k8s.io,resources=networkpolicies,verbs=create;delete
// +kubebuilder:rbac:groups=oauth.openshift.io,resources=oauthclients,verbs=*
// +kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=clusterroles;clusterrolebindings;roles;rolebindings,verbs=*
// +kubebuilder:rbac:groups=route.openshift.io,resources=routes;routes/custom-host,verbs=*
// +kubebuilder:rbac:groups=scheduling.k8s.io,resources=priorityclasses,verbs=*
// +kubebuilder:rbac:groups=security.openshift.io,resources=securitycontextconstraints,verbs=create;use;get;list;watch
// +kubebuilder:rbac:urls=/metrics,verbs=get

// +kubebuilder:rbac:groups=config.openshift.io,resources=apiservers;clusterversions,verbs=get;list;watch

// +kubebuilder:rbac:groups=observability.openshift.io,resources=clusterlogforwarders,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=observability.openshift.io,resources=clusterlogforwarders/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=observability.openshift.io,resources=clusterlogforwarders/finalizers,verbs=update
