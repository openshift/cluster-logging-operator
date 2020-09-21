package fluentbit

const (
	Parsers = `
## CLO GENERATED CONFIGURATION ###
# This file is a generated fluentbit configuration
# supplied in a configmap.
[PARSER]
    Name containerd
    Format regex
    Regex /^(?<time>.+) (?<stream>\w+) (?<logtag>[FP]) (?<message>.+)$/
    Time_Key time
    Time_Keep on
    Time_Format %Y-%m-%dT%H:%M:%S.%L
[PARSER]
    Name        json
    Format      json
    Time_Key    time
    Time_Format %Y-%m-%dT%H:%M:%S.%L
`
)
