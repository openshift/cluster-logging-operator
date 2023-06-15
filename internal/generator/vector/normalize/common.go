package normalize

const (
	ClusterID             = `.openshift.cluster_id = "${OPENSHIFT_CLUSTER_ID:-}"`
	FixTimestampField     = `ts = del(.timestamp); if !exists(."@timestamp") {."@timestamp" = ts}`
	MergeJsonMessageField = `if is_json(string!(.message)) {
    log(">>>>>>>>>>>>>>>>> 1", level: "error")
    msj = parse_json!(string!(.message))
    if (is_object(msj.event) && is_string(msj.verb)) {
      log(">>>>>>>>>>>>>>>>> 2", level: "error")
      log(., level: "error")
      ., err = merge(., msj)
      log(">>>>>>>>>>>>>>>>> 3", level: "error")
      log(., level: "error")
      if err == null {
        log(err, level: "error")
      }
    }
  }`
)
