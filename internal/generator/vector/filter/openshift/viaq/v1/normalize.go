package v1

const (
	HandleEventRouterLog = `

if exists(._internal.kubernetes.pod_name) && starts_with(string!(._internal.kubernetes.pod_name), "eventrouter-") {
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
	MergeStructuredIntoRoot = `
if ._internal.log_type == "audit" && exists(._internal.structured) {. = merge!(.,._internal.structured) }
if ._internal.log_source == "syslog" && exists(._internal.structured) {. = merge!(.,._internal.structured) }
if ._internal.log_source == "container" && exists(._internal.structured) {.structured = ._internal.structured }
`
	SetLogLevel = `

if !exists(._internal.level) {
  level = null
  message = ._internal.message

  # attempt 1: parse as logfmt (e.g. level=error msg="Failed to connect")
  
  parsed_logfmt, err = parse_logfmt(message)  
  if err == null && is_string(parsed_logfmt.level) {
    level = downcase!(parsed_logfmt.level)
  }

  # attempt 2: parse as klog (e.g. I0920 14:22:00.089385 1 scheduler.go:592] "Successfully bound pod to node")
  if level == null {
    parsed_klog, err = parse_klog(message)  
    if err == null && is_string(parsed_klog.level) {
      level = parsed_klog.level
    }
  }

  # attempt 3: parse with groks template (if previous attempts failed) for classic text logs like Logback, Log4j etc. 
  
  if level == null {
    parsed_grok, err = parse_groks(message,
      patterns: [
        "%{common_prefix} %{_message}"
      ],
      aliases: {
        "common_prefix": "%{_timestamp} %{_loglevel}",
        "_timestamp": "%{TIMESTAMP_ISO8601:timestamp}",
        "_loglevel": "%{LOGLEVEL:level}",
        "_message": "%{GREEDYDATA:message}"
      }
    )
    
    if err == null && is_string(parsed_grok.level) {
      level = downcase!(parsed_grok.level)      
    }
  }

  if level == null {
    level = "default"
 
    # attempt 4: Match on well known structured patterns
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
  
    # attempt 5: Match on the keyword that appears earliest in the message
	if level == "default" {
	  level_patterns = r'(?i)(?<emergency>emergency|<emergency>)|(?<alert>alert|<alert>)|(?<critical>critical|<critical>)|(?<error>error|<error>)|(?<warn>warn(?:ing)?|<warn>)|(?<notice>notice|<notice>)|(?:\b(?<info>info)\b|<info>)|(?<debug>debug|<debug>)|(?<trace>trace|<trace>)'
	  parsed, err = parse_regex(message, level_patterns)
	  if err == null {
		if is_string(parsed.emergency) {
		  level = "emergency"
		} else if is_string(parsed.alert) {
		  level = "alert"
		} else if is_string(parsed.critical) {
		  level = "critical"
		} else if is_string(parsed.error) {
		  level = "error"
		} else if is_string(parsed.warn) {
		  level = "warn"
		} else if is_string(parsed.notice) {
		  level = "notice"
		} else if is_string(parsed.info) {
		  level = "info"
		} else if is_string(parsed.debug) {
		  level = "debug"
		} else if is_string(parsed.trace) {
		  level = "trace"
		}
	  }
	}
  }
  ._internal.level = level
}
`
	SetLogLevelOnRoot = `
if ._internal.log_type != "audit" && exists(._internal.level) {
  .level = ._internal.level
}
`

	SetLogTypeOnRoot    = ".log_type = ._internal.log_type"
	SetHostnameOnRoot   = ` if exists(._internal.hostname) { .hostname = ._internal.hostname }`
	SetLogSourceOnRoot  = ".log_source = ._internal.log_source"
	SetKubernetesOnRoot = `
.kubernetes = ._internal.kubernetes
.kubernetes.container_iostream = ._internal.stream
if exists(._internal.dedot_labels) {.kubernetes.labels = del(._internal.dedot_labels) }
if exists(._internal.dedot_namespace_labels) {.kubernetes.namespace_labels = del(._internal.dedot_namespace_labels) }
del(.kubernetes.node_labels)
del(.kubernetes.container_image_id)
del(.kubernetes.pod_ips)
`
	SetMessageOnRoot = `
if !exists(._internal.structured) {
  .message = ._internal.message
}
`
	SetOpenShiftOnRoot = `
if exists(._internal.openshift) {.openshift = ._internal.openshift}
if exists(._internal.dedot_openshift_labels) {.openshift.labels = del(._internal.dedot_openshift_labels) }
`
	SetOpenShiftSequence = `if ._internal.log_type != "receiver" { ._internal.openshift.sequence = to_unix_timestamp(now(), unit: "nanoseconds")}`
)
