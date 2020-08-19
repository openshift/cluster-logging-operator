package normalizer

const (
	fluentbitConf = `
[SERVICE]
    Log_Level info
    Parsers_file /etc/fluent-bit/parsers.conf

[INPUT]
    Name tail
    Path /var/log/containers/*_openshift*_*.log, /var/log/containers/*_kube*_*.log, /var/log/containers/*_default_*.log
    Path_Key filename
    Parser containerd
    Exclude_Path /var/log/containers/collector*_openshift*_*.log, /var/log/containers/elasticsearch*_openshift*_*.log,/var/log/containers/kibana*_openshift*_*.log,/var/log/containers/normalizer*_openshift*_*.log
    Tag kubernetes.*
    DB /var/log/infra-containers.pos.db
    Refresh_Interval 5

[INPUT]
    Name tail
    Path /var/log/containers/*.log
    Path_Key filename
    Parser containerd
    Exclude_Path /var/log/containers/*_openshift*_*.log, /var/log/containers/*_kube*_*.log, /var/log/containers/*_default_*.log
    Tag kubernetes.*
    DB /var/log/app-containers.pos.db
    Refresh_Interval 5

[INPUT]
    Name tail
    Path /var/log/audit/audit.log
    Path_Key filename
    Parser json
    Tag linux-audit.log
    DB /var/log/audit-linux.pos.db
    Refresh_Interval 5

[INPUT]
    Name tail
    Path /var/log/kube-apiserver/audit.log
    Path_Key filename
    Parser json
    Tag k8s-audit.log
    DB /var/log/audit-k8s.pos.db
    Refresh_Interval 5

[INPUT]
    Name tail
    Path /var/log/oauth-apiserver/audit.log, /var/log/openshift-apiserver/audit.log
    Path_Key filename
    Parser json
    Tag openshift-audit.log
    DB /var/log/audit-oauth.pos.db
    Refresh_Interval 5

[INPUT]
    Name systemd
    Path /var/log/journal
    Tag journal
    DB /var/log/journal.pos.db

[OUTPUT]
    Name forward
    Host normalizer.openshift-logging.svc
    Match *
`
	fluetnbitParserConf = `
[PARSER]
    Name containerd
    Format regex
    Regex /^(?<time>.+) (?<stream>\w+) (?<logtag>[FP]) (?<log>.+)$/
    Time_Key time
    Time_Keep on
    Time_Format %Y-%m-%dT%H:%M:%S.%L

[PARSER]
    Name json
    Format json
    Time_Key requestReceivedTimestamp
    Time_Format %Y-%m-%dT%H:%M:%S.%L
    Time_Keep on
`
)
