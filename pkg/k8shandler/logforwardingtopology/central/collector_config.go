package central

const (
	fluentbitConf = `
[SERVICE]
    Log_Level ${LOG_LEVEL}
    HTTP_Server  On
    HTTP_Listen  ${POD_IP}
    HTTP_PORT    2020
    Parsers_file /etc/fluent-bit/parsers.conf

[INPUT]
    Name systemd
    Path /var/log/journal
    Tag journal
    DB /var/lib/fluent-bit/journal.pos.db
    Read_From_Tail On

[INPUT]
    Name tail
    Path /var/log/containers/*_openshift*_*.log, /var/log/containers/*_kube*_*.log, /var/log/containers/*_default_*.log
    Path_Key filename
    Parser containerd
    Exclude_Path /var/log/containers/*_openshift-logging*_*.log
    Tag kubernetes.*
    DB /var/lib/fluent-bit/infra-containers.pos.db
    Refresh_Interval 5

[INPUT]
    Name tail
    Path /var/log/containers/*.log
    Path_Key filename
    Parser containerd
    Exclude_Path /var/log/containers/*_openshift*_*.log, /var/log/containers/*_kube*_*.log, /var/log/containers/*_default_*.log
    Tag kubernetes.*
    DB /var/lib/fluent-bit/app-containers.pos.db
    Refresh_Interval 5

[INPUT]
    Name tail
    Path /var/log/audit/audit.log
    Path_Key filename
    Parser json
    Tag linux-audit.log
    DB  /var/lib/fluent-bit/audit-linux.pos.db
    Refresh_Interval 5

[INPUT]
    Name tail
    Path /var/log/kube-apiserver/audit.log
    Path_Key filename
    Parser json
    Tag k8s-audit.log
    DB /var/lib/fluent-bit/audit-k8s.pos.db
    Refresh_Interval 5

[INPUT]
    Name tail
    Path /var/log/oauth-apiserver/audit.log, /var/log/openshift-apiserver/audit.log
    Path_Key filename
    Parser json
    Tag openshift-audit.log
    DB /var/lib/fluent-bit/audit-oauth.pos.db
    Refresh_Interval 5

[INPUT]
    Name systemd
    Path /var/log/journal
    Tag journal
    DB /var/lib/fluent-bit/journal.pos.db

[FILTER]
    Name    lua
    Match   kubernetes.*
    script  /etc/fluent-bit/concat-crio.lua
    call reassemble_cri_logs

[OUTPUT]
    Name forward
    Host normalizer.openshift-logging.svc
    Match *
`
)
