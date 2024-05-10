package viaq

const (
	FixLogLevel = `
if !exists(.level) {
  .level = "default"
  if match!(.message, r'Warning|WARN|^W[0-9]+|level=warn|Value:warn|"level":"warn"|<warn>') {
    .level = "warn"
  } else if match!(.message, r'Error|ERROR|^E[0-9]+|level=error|Value:error|"level":"error"|<error>') {
    .level = "error"
  } else if match!(.message, r'Critical|CRITICAL|^C[0-9]+|level=critical|Value:critical|"level":"critical"|<critical>') {
    .level = "critical"
  } else if match!(.message, r'Debug|DEBUG|^D[0-9]+|level=debug|Value:debug|"level":"debug"|<debug>') {
    .level = "debug"
  } else if match!(.message, r'Notice|NOTICE|^N[0-9]+|level=notice|Value:notice|"level":"notice"|<notice>') {
    .level = "notice"
  } else if match!(.message, r'Alert|ALERT|^A[0-9]+|level=alert|Value:alert|"level":"alert"|<alert>') {
    .level = "alert"
  } else if match!(.message, r'Emergency|EMERGENCY|^EM[0-9]+|level=emergency|Value:emergency|"level":"emergency"|<emergency>') {
    .level = "emergency"
  } else if match!(.message, r'(?i)\b(?:info)\b|^I[0-9]+|level=info|Value:info|"level":"info"|<info>') {
    .level = "info"
	}
}
`
	RemoveSourceType     = `del(.source_type)`
	HandleEventRouterLog = `
pod_name = string!(.kubernetes.pod_name)
if starts_with(pod_name, "eventrouter-") {
  parsed, err = parse_json(.message)
  if err != null {
    log("Unable to process EventRouter log: " + err, level: "info")
  } else {
    ., err = merge(.,parsed)
    if err == null && exists(.event) && is_object(.event) {
        if exists(.verb) {
          .event.verb = .verb
          del(.verb)
        }
        .kubernetes.event = del(.event)
        .message = del(.kubernetes.event.message)
        . = set!(., ["@timestamp"], .kubernetes.event.metadata.creationTimestamp)
        del(.kubernetes.event.metadata.creationTimestamp)
		. = compact(., nullish: true)
    } else {
      log("Unable to merge EventRouter log message into record: " + err, level: "info")
    }
  }
}
`
	RemoveStream       = `del(.stream)`
	RemovePodIPs       = `del(.kubernetes.pod_ips)`
	RemoveNodeLabels   = `del(.kubernetes.node_labels)`
	RemoveTimestampEnd = `del(.timestamp_end)`

	ParseAndFlatten = `. = merge(., parse_json!(string!(.message))) ?? .
del(.message)
`
	FixHostname = `.hostname = get_env_var("VECTOR_SELF_NODE_NAME") ?? ""`
)
