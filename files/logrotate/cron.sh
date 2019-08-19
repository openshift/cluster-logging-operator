#!/bin/sh
# Wrapper to start cronjob to invoke logrotate periodically

echo "export LOGGING_FILE_PATH=${LOGGING_FILE_PATH:-}" > /tmp/.logrotate
echo "export LOGGING_FILE_SIZE=${LOGGING_FILE_SIZE:-}" >> /tmp/.logrotate
echo "export LOGGING_FILE_AGE=${LOGGING_FILE_AGE:-}" >> /tmp/.logrotate
echo "export RSYSLOG_WORKDIRECTORY=${RSYSLOG_WORKDIRECTORY:-/var/lib/rsyslog.pod}" >> /tmp/.logrotate

exec /usr/sbin/crond -n $CROND_OPTIONS
