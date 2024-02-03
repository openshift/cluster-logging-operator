// package v2beta custom resorces for the cluster logging operator.
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
package v2beta1

/* FIXME: NOTES and OPEN QUESTIONS

   # Filter types

   Not documented yet;
   { name: x, type: kubeapiaudit, kubeapiaudit: { ... already defined } }
   { name: x, type: multiLineErrors } # No parameters? List of languages to enable?
   { name: x, type: parseJSON } # No parameters.

   Future:
   { name: x, type: prune, prune: {...} } # See separate proposal
   { name: x, type: drop, drop: {...} } # See separate proposal
   { name: x, type: add, add: { "field.name": value, "field2.name": value2... }} # In place of "labels", add any fiel

   # Templated names

   Need to document (or use special type) to mark strings that can contain "{{field.name}}"
   templates.

*/
