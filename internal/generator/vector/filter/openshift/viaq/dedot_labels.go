package viaq

const (
	VRLDedotLabels = `
if ._internal.log_source == "container" {
  if exists(._internal.kubernetes.namespace_labels) {
    for_each(object!(._internal.kubernetes.namespace_labels)) -> |key,value| { 
      newkey = replace(key, r'[\./]', "_") 
      ._internal.kubernetes.namespace_labels = set!(._internal.kubernetes.namespace_labels,[newkey],value)
      if newkey != key {._internal.kubernetes.namespace_labels = remove!(._internal.kubernetes.namespace_labels,[key],true)}
    }
  }
  if exists(._internal.kubernetes.labels) {
    for_each(object!(._internal.kubernetes.labels)) -> |key,value| { 
      newkey = replace(key, r'[\./]', "_") 
      ._internal.kubernetes.labels = set!(._internal.kubernetes.labels,[newkey],value)
      if newkey != key {._internal.kubernetes.labels = remove!(._internal.kubernetes.labels,[key],true)}
    }
  }
}
if exists(._internal.openshift.labels) {for_each(object!(._internal.openshift.labels)) -> |key,value| {
  newkey = replace(key, r'[\./]', "_") 
  ._internal.openshift.labels = set!(._internal.openshift.labels,[newkey],value)
  if newkey != key {._internal.openshift.labels = remove!(._internal.openshift.labels,[key],true)}
}}
`
)
