#!/bin/sh

set -eu

. /tmp/.logrotate

dirname=$( dirname ${LOGGING_FILE_PATH:-"/var/log/rsyslog/rsyslog.log"} )
cat > /tmp/logrotate.conf << EOF
"$dirname/*.log"
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

exec /usr/sbin/logrotate --log $dirname/logrotate.log /tmp/logrotate.conf
