// package v2alpha1 custom resorces for the cluster logging operator.
//
// NOTE: This API is under construction and subject to change.
//
// NOTE: Doc comments are written using JSON names (lowerCase) rather than Go UpperCase stlye.
//
// IMPORTANT: Run "make generate" to regenerate code after modifying any files in this package.
// Add custom validation using kubebuilder tags: https://book.kubebuilder.io/beyond_basics/generating_crd.html
//
// +groupName=logging.openshift.io
// +k8s:deepcopy-gen=package,register
// +k8s:openapi-gen=true
// +kubebuilder:object:generate=true
package v2alpha1

// FIXME Still includes fluentd as well as vector settings. This is not to say we should support fluentd with v1 (separate discussion),
// just to show what the APIs look like with multiple options.
