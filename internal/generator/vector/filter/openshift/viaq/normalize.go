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
if exists(.kubernetes.pod_name) && starts_with(string!(.kubernetes.pod_name), "eventrouter-") {
  parsed, err = parse_json(.message)
  if err != null {
    log("Unable to process EventRouter log: " + err, level: "info")
  } else {
    ._internal.event = parsed
    if exists(._internal.event.event) && is_object(._internal.event.event) {
      ._internal.kubernetes.event = del(._internal.event.event)
      ._internal.kubernetes.event.verb = del(._internal.event.verb)
      # escape 'new line' symbol see: LOG-8090
      msg = to_string!(del(._internal.kubernetes.event.message))
      ._internal.message = replace(msg, "\n", s'\n')
      # Determine event timestamp: prefer lastTimestamp, then firstTimestamp (v1 events),
      # fall back to eventTime (events.k8s.io/v1), then creationTimestamp
      # lastTimestamp -> firstTimestamp -> eventTime -> creationTimestamp
      zero_ts = "0001-01-01T00:00:00Z"
      candidate_timestamps = filter([
        ._internal.kubernetes.event.lastTimestamp,
        ._internal.kubernetes.event.firstTimestamp,
        ._internal.kubernetes.event.eventTime,
        ._internal.kubernetes.event.metadata.creationTimestamp
      ]) -> |_index, value| {
        value != null && value != "" && value != zero_ts
      }
      ts = candidate_timestamps[0]
      ."@timestamp" = ts
      .timestamp = ts
      .message = ._internal.message
      .kubernetes.event = del(._internal.kubernetes.event)
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
