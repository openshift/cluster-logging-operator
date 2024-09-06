package viaq

const (
	HandleEventRouterLog = `
pod_name = string!(._internal.kubernetes.pod_name)
if starts_with(pod_name, "eventrouter-") {
  parsed, err = parse_json(._internal.message)
  if err != null {
    log("Unable to process EventRouter log: " + err, level: "info")
  } else {
    ._internal.event = parsed
    if exists(._internal.event.event) && is_object(._internal.event.event) {
        ._internal.kubernetes.event = del(._internal.event.event)
		._internal.kubernetes.event.verb = ._internal.event.verb
        ._internal.message = del(._internal.kubernetes.event.message)
        ._internal."@timestamp" = .kubernetes.event.metadata.creationTimestamp
    } else {
      log("Unable to merge EventRouter log message into record: " + err, level: "info")
    }
  }
}
`
	MergeStructuredIntoRoot = "if exists(._internal.structured) {. = merge!(.,._internal.structured) }"
	SetLogLevel             = `
if !exists(._internal.level) {
  level = "default"
  message = ._internal.message

  # Match on well known structured patterns
  # Order: emergency, alert, critical, error, warn, notice, info, debug, trace

  if match!(message, r'^EM[0-9]+|level=emergency|Value:emergency|"level":"emergency"') {
    level = "emergency"
  } else if match!(message, r'^A[0-9]+|level=alert|Value:alert|"level":"alert"') {
    level = "alert"
  } else if match!(message, r'^C[0-9]+|level=critical|Value:critical|"level":"critical"') {
    level = "critical"
  } else if match!(message, r'^E[0-9]+|level=error|Value:error|"level":"error"') {
    level = "error"
  } else if match!(message, r'^W[0-9]+|level=warn|Value:warn|"level":"warn"') {
    level = "warn"
  } else if match!(message, r'^N[0-9]+|level=notice|Value:notice|"level":"notice"') {
    level = "notice"
  } else if match!(message, r'^I[0-9]+|level=info|Value:info|"level":"info"') {
    level = "info"
  } else if match!(message, r'^D[0-9]+|level=debug|Value:debug|"level":"debug"') {
    level = "debug"
  } else if match!(message, r'^T[0-9]+|level=trace|Value:trace|"level":"trace"') {
    level = "trace"
  }

  # Match on unstructured keywords in same order

  if level == "default" {
    if match!(message, r'Emergency|EMERGENCY|<emergency>') {
      level = "emergency"
    } else if match!(message, r'Alert|ALERT|<alert>') {
      level = "alert"
    } else if match!(message, r'Critical|CRITICAL|<critical>') {
      level = "critical"
    } else if match!(message, r'Error|ERROR|<error>') {
      level = "error"
    } else if match!(message, r'Warning|WARN|<warn>') {
      level = "warn"
    } else if match!(message, r'Notice|NOTICE|<notice>') {
      level = "notice"
    } else if match!(message, r'(?i)\b(?:info)\b|<info>') {
      level = "info"
    } else if match!(message, r'Debug|DEBUG|<debug>') {
      .level = "debug"
    } else if match!(message, r'Trace|TRACE|<trace>') {
      .level = "trace"
    }
  }
  ._internal.level = level
}
`

	SetLogTypeOnRoot    = ".log_type = ._internal.log_type"
	SetHostnameOnRoot   = `.hostname = ._internal.hostname`
	SetLogSourceOnRoot  = ".log_source = ._internal.log_source"
	SetKubernetesOnRoot = `
.kubernetes = ._internal.kubernetes
del(.kubernetes.node_labels)
del(.kubernetes.container_image_id)
del(.kubernetes.pod_ips)
`
	SetMessageOnRoot = `
if !exists(._internal.structured) {
  .message = ._internal.message
}
`
	SetOpenShiftOnRoot   = `if exists(._internal.openshift) {.openshift = ._internal.openshift}`
	SetOpenShiftSequence = `._internal.openshift.sequence = to_unix_timestamp(now(), unit: "nanoseconds")`
)
