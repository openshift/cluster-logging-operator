package v1

const (
	VRLDedotLabels = `
if ._internal.log_source == "container" {
  if exists(._internal.kubernetes.namespace_labels) {
    ._internal.dedot_namespace_labels = {}
    for_each(object!(._internal.kubernetes.namespace_labels)) -> |key,value| { 
      newkey = replace(key, r'[\./]', "_") 
      ._internal.dedot_namespace_labels = set!(._internal.dedot_namespace_labels,[newkey],value)
    }
  }
  if exists(._internal.kubernetes.labels) {
    ._internal.dedot_labels = {}
    for_each(object!(._internal.kubernetes.labels)) -> |key,value| { 
      newkey = replace(key, r'[\./]', "_") 
      ._internal.dedot_labels = set!(._internal.dedot_labels,[newkey],value)
    }
  }
}
if exists(._internal.openshift.labels) {for_each(object!(._internal.openshift.labels)) -> |key,value| {
  ._internal.dedot_openshift_labels = {}
  newkey = replace(key, r'[\./]', "_") 
  ._internal.dedot_openshift_labels = set!(._internal.dedot_openshift_labels,[newkey],value)
}}
`
)
