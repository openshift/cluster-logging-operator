package viaq

const (
	FixLogLevel = `
if !exists(.level) {
  .level = "default"

  # Match on well known structured patterns
  # Order: emergency, alert, critical, error, warn, notice, info, debug, trace

  if match!(.message, r'^EM[0-9]+|level=emergency|Value:emergency|"level":"emergency"') {
    .level = "emergency"
  } else if match!(.message, r'^A[0-9]+|level=alert|Value:alert|"level":"alert"') {
    .level = "alert"
  } else if match!(.message, r'^C[0-9]+|level=critical|Value:critical|"level":"critical"') {
    .level = "critical"
  } else if match!(.message, r'^E[0-9]+|level=error|Value:error|"level":"error"') {
    .level = "error"
  } else if match!(.message, r'^W[0-9]+|level=warn|Value:warn|"level":"warn"') {
    .level = "warn"
  } else if match!(.message, r'^N[0-9]+|level=notice|Value:notice|"level":"notice"') {
    .level = "notice"
  } else if match!(.message, r'^I[0-9]+|level=info|Value:info|"level":"info"') {
    .level = "info"
  } else if match!(.message, r'^D[0-9]+|level=debug|Value:debug|"level":"debug"') {
    .level = "debug"
  } else if match!(.message, r'^T[0-9]+|level=trace|Value:trace|"level":"trace"') {
    .level = "trace"
  }

  # Match on unstructured keywords in same order

  if .level == "default" {
    if match!(.message, r'Emergency|EMERGENCY|<emergency>') {
      .level = "emergency"
    } else if match!(.message, r'Alert|ALERT|<alert>') {
      .level = "alert"
    } else if match!(.message, r'Critical|CRITICAL|<critical>') {
      .level = "critical"
    } else if match!(.message, r'Error|ERROR|<error>') {
      .level = "error"
    } else if match!(.message, r'Warning|WARN|<warn>') {
      .level = "warn"
    } else if match!(.message, r'Notice|NOTICE|<notice>') {
      .level = "notice"
    } else if match!(.message, r'(?i)\b(?:info)\b|<info>') {
      .level = "info"
    } else if match!(.message, r'Debug|DEBUG|<debug>') {
      .level = "debug"
    } else if match!(.message, r'Trace|TRACE|<trace>') {
      .level = "trace"
    }
  }
}
`
	RemovePartial        = `del(._partial)`
	RemoveFile           = `del(.file)`
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
	HandleStream       = `.kubernetes.container_iostream = del(.stream)`
	RemovePodIPs       = `del(.kubernetes.pod_ips)`
	RemoveNodeLabels   = `del(.kubernetes.node_labels)`
	RemoveTimestampEnd = `del(.timestamp_end)`

	ParseAndFlatten = `. = merge(., parse_json!(string!(.message))) ?? .
del(.message)
`
	FixHostname = `.hostname = get_env_var("VECTOR_SELF_NODE_NAME") ?? ""`
)
