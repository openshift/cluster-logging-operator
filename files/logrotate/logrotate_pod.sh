#!/bin/sh

set -eu

. /tmp/.logrotate

cat > /tmp/logrotate_pod.conf << EOF
"${RSYSLOG_WORKDIRECTORY}/*.log"
"${RSYSLOG_WORKDIRECTORY}/impstats.json"
{
  create 0644 root
  dateext
  dateformat -%Y%m%d-%s
  missingok
  notifempty
  size ${LOGGING_FILE_SIZE:-1024000}
  rotate ${LOGGING_FILE_AGE:-10}
  postrotate
    # rsyslogd is "exec"ed in the first script rsyslog.sh
    kill -HUP $( cat /var/run/rsyslogd.pid )
  endscript
}
EOF

exec /usr/sbin/logrotate --log ${RSYSLOG_WORKDIRECTORY}/logrotate.log /tmp/logrotate_pod.conf
