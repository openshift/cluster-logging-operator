#!/bin/bash

CFG_DIR=/etc/rsyslog.d
ENABLE_PROMETHEUS_ENDPOINT=${ENABLE_PROMETHEUS_ENDPOINT:-"false"}

rsyslogargs="-f /etc/rsyslog/conf/rsyslog.conf -n"
if [[ $VERBOSE ]]; then
  set -ex
  rsyslogargs="$rsyslogargs -d"
  echo ">>>>>> ENVIRONMENT VARS <<<<<"
  env | sort
  echo ">>>>>>>>>>>>><<<<<<<<<<<<<<<<"
else
  set -e
fi

issue_deprecation_warnings() {
    : # none at the moment
}

if type -p python3 > /dev/null 2>&1 ; then
    pycmd=python3
else
    pycmd=python
fi

if [ -f /etc/docker-hostname ] ; then
    HOSTNAME=`cat /etc/docker-hostname`
    IPADDR4=`getent ahostsv4 $HOSTNAME | awk '/ STREAM / {print $1}'`
    IPADDR6=`getent ahostsv6 $HOSTNAME | awk '/ STREAM / {print $1}'`
elif [ -n "${NODE_IPV4:-}" ] ; then
    IPADDR4=$NODE_IPV4
    HOSTNAME=`getent hosts $IPADDR4 | awk '{print $2}'`
    IPADDR6=`getent ahostsv6 $HOSTNAME | awk '/ STREAM / {print $1}'`
elif [ -x /usr/sbin/ip ] ; then
    HOSTNAME=`$pycmd -c 'import socket; print(socket.gethostname())'`
    IPADDR4=`/usr/sbin/ip -4 addr show dev eth0 | grep inet | sed -e "s/[ \t]*inet \([0-9.]*\).*/\1/"`
    IPADDR6=`/usr/sbin/ip -6 addr show dev eth0 | grep inet6 | sed "s/[ \t]*inet6 \([a-f0-9:]*\).*/\1/"`
else
    HOSTNAME=`$pycmd -c 'import socket; print(socket.gethostname())'`
    IPADDR4=`getent ahostsv4 $HOSTNAME | awk '/ STREAM / {print $1}'`
    IPADDR6=`getent ahostsv6 $HOSTNAME | awk '/ STREAM / {print $1}'`
fi
RSYSLOG_VERSION=`/usr/sbin/rsyslogd -v | awk -F'[ ,]+' '/^rsyslogd / {print $2}'`
PIPELINE_VERSION="${RSYSLOG_VERSION} ${DATA_VERSION}"
export IPADDR4 IPADDR6 PIPELINE_VERSION HOSTNAME

BUFFER_SIZE_LIMIT=${BUFFER_SIZE_LIMIT:-16777216}

export PIPELINE_TYPE=collector

if [ -z $ES_HOST ]; then
    echo "ERROR: Environment variable ES_HOST for Elasticsearch host name is not set."
    exit 1
fi
if [ -z $ES_PORT ]; then
    echo "ERROR: Environment variable ES_PORT for Elasticsearch port number is not set."
    exit 1
fi

OPS_HOST=${OPS_HOST:-$ES_HOST}
OPS_PORT=${OPS_PORT:-$ES_PORT}
OPS_CA=${OPS_CA:-$ES_CA}
OPS_CLIENT_CERT=${OPS_CLIENT_CERT:-$ES_CLIENT_CERT}
OPS_CLIENT_KEY=${OPS_CLIENT_KEY:-$ES_CLIENT_KEY}
export OPS_HOST OPS_PORT OPS_CA OPS_CLIENT_CERT OPS_CLIENT_KEY

# How many outputs?
# check ES_HOST vs. OPS_HOST; ES_PORT vs. OPS_PORT
if [ "$ES_HOST" = ${OPS_HOST:-""} -a $ES_PORT -eq ${OPS_PORT:-0} ]; then
    # There is one output Elasticsearch
    NUM_OUTPUTS=1
    ES_OUTPUT_NAME=elasticsearch
    ES_OPS_OUTPUT_NAME=elasticsearch
else
    NUM_OUTPUTS=2
    ES_OUTPUT_NAME=elasticsearch-app
    ES_OPS_OUTPUT_NAME=elasticsearch-infra
fi
export ES_OUTPUT_NAME ES_OPS_OUTPUT_NAME

# If FILE_BUFFER_PATH exists and it is not a directory, mkdir fails with the error.
RSYSLOG_WORKDIRECTORY=${RSYSLOG_WORKDIRECTORY:-/var/lib/rsyslog.pod}
if [ ! -d $RSYSLOG_WORKDIRECTORY ] ; then
    mkdir -p $RSYSLOG_WORKDIRECTORY
fi
RSYSLOG_SPOOLDIRECTORY=${RSYSLOG_SPOOLDIRECTORY:-$RSYSLOG_WORKDIRECTORY}
RSYSLOG_BULK_ERRORS=${RSYSLOG_BULK_ERRORS:-"$RSYSLOG_WORKDIRECTORY/es-bulk-errors.log"}
RSYSLOG_IMJOURNAL_STATE=${RSYSLOG_IMJOURNAL_STATE:-"$RSYSLOG_WORKDIRECTORY/imjournal.state"}
RSYSLOG_IMPSTATS_FILE=${RSYSLOG_IMPSTATS_FILE:-"$RSYSLOG_WORKDIRECTORY/impstats.json"}
export RSYSLOG_WORKDIRECTORY RSYSLOG_SPOOLDIRECTORY RSYSLOG_BULK_ERRORS \
  RSYSLOG_IMJOURNAL_STATE RSYSLOG_IMPSTATS_FILE
FILE_BUFFER_PATH=${FILE_BUFFER_PATH:-$RSYSLOG_WORKDIRECTORY}
mkdir -p $FILE_BUFFER_PATH

# Get the available disk size.
DF_LIMIT=$(df -B1 $FILE_BUFFER_PATH | grep -v Filesystem | awk '{print $2}')
DF_LIMIT=${DF_LIMIT:-0}
if [ "${MUX_FILE_BUFFER_STORAGE_TYPE:-}" = "hostmount" ]; then
    # Use 1/4 of the disk space for hostmount.
    DF_LIMIT=$(expr $DF_LIMIT / 4) || :
fi
if [ $DF_LIMIT -eq 0 ]; then
    echo "ERROR: No disk space is available for file buffer in $FILE_BUFFER_PATH."
    exit 1
fi

cnvt_to_bytes() {
    local byteval=$1
    echo "$1" | \
    sed -e "s/[Kk]/*1024/g;s/[Mm]/*1024*1024/g;s/[Gg]/*1024*1024*1024/g;s/i//g" | \
    $pycmd -c 'import sys; print(eval(sys.stdin.read()))'
}

# Determine final total given the number of outputs we have.
TOTAL_LIMIT=$(cnvt_to_bytes ${FILE_BUFFER_LIMIT:-2Gi}) || :
if [ $TOTAL_LIMIT -le 0 ]; then
    echo "ERROR: Invalid file buffer limit ($FILE_BUFFER_LIMIT) is given.  Failed to convert to bytes."
    exit 1
fi

# If forward and secure-forward outputs are configured, add them to NUM_OUTPUTS.
forward_files=$( grep -l "@type .*forward" ${CFG_DIR}/*/* 2> /dev/null || : )
for afile in ${forward_files} ; do
    file=$( basename $afile )
    if [ "$file" != "${mux_client_filename:-}" ]; then
        grep "@type .*forward" $afile | while read -r line; do
            if [ $( expr "$line" : "^ *#" ) -eq 0 ]; then
                NUM_OUTPUTS=$( expr $NUM_OUTPUTS + 1 )
            fi
        done
    fi
done

TOTAL_LIMIT=$(expr $TOTAL_LIMIT \* $NUM_OUTPUTS) || :
if [ $DF_LIMIT -lt $TOTAL_LIMIT ]; then
    echo "WARNING: Available disk space ($DF_LIMIT bytes) is less than the user specified file buffer limit ($FILE_BUFFER_LIMIT times $NUM_OUTPUTS)."
    TOTAL_LIMIT=$DF_LIMIT
fi

BUFFER_SIZE_LIMIT=$(cnvt_to_bytes $BUFFER_SIZE_LIMIT)
BUFFER_SIZE_LIMIT=${BUFFER_SIZE_LIMIT:-16777216}

# TOTAL_BUFFER_SIZE_LIMIT per buffer
TOTAL_BUFFER_SIZE_LIMIT=$(expr $TOTAL_LIMIT / $NUM_OUTPUTS) || :
if [ -z $TOTAL_BUFFER_SIZE_LIMIT -o $TOTAL_BUFFER_SIZE_LIMIT -eq 0 ]; then
    echo "ERROR: Calculated TOTAL_BUFFER_SIZE_LIMIT is 0. TOTAL_LIMIT $TOTAL_LIMIT is too small compared to NUM_OUTPUTS $NUM_OUTPUTS. Please increase FILE_BUFFER_LIMIT $FILE_BUFFER_LIMIT and/or the volume size of $FILE_BUFFER_PATH."
    exit 1
fi
BUFFER_QUEUE_LIMIT=$(expr $TOTAL_BUFFER_SIZE_LIMIT / $BUFFER_SIZE_LIMIT) || :
if [ -z $BUFFER_QUEUE_LIMIT -o $BUFFER_QUEUE_LIMIT -eq 0 ]; then
    echo "ERROR: Calculated BUFFER_QUEUE_LIMIT is 0. TOTAL_BUFFER_SIZE_LIMIT $TOTAL_BUFFER_SIZE_LIMIT is too small compared to BUFFER_SIZE_LIMIT $BUFFER_SIZE_LIMIT. Please increase FILE_BUFFER_LIMIT $FILE_BUFFER_LIMIT and/or the volume size of $FILE_BUFFER_PATH."
    exit 1
fi
export BUFFER_QUEUE_LIMIT BUFFER_SIZE_LIMIT

ES_QUEUE_TYPE=${ES_QUEUE_TYPE:-Disk}
OPS_QUEUE_TYPE=${OPS_QUEUE_TYPE:-$ES_QUEUE_TYPE}
ES_QUEUE_MAXDISKSPACE=${ES_QUEUE_MAXDISKSPACE:-$TOTAL_BUFFER_SIZE_LIMIT}
OPS_QUEUE_MAXDISKSPACE=${OPS_QUEUE_MAXDISKSPACE:-$ES_QUEUE_MAXDISKSPACE}
ES_QUEUE_MAXFILESIZE=${ES_QUEUE_MAXFILESIZE:-$BUFFER_SIZE_LIMIT}
OPS_QUEUE_MAXFILESIZE=${OPS_QUEUE_MAXFILESIZE:-$ES_QUEUE_MAXFILESIZE}
ES_QUEUE_CHECKPOINTINTERVAL=${ES_QUEUE_CHECKPOINTINTERVAL:-1000}
OPS_QUEUE_CHECKPOINTINTERVAL=${OPS_QUEUE_CHECKPOINTINTERVAL:-$ES_QUEUE_CHECKPOINTINTERVAL}
export ES_QUEUE_TYPE OPS_QUEUE_TYPE ES_QUEUE_MAXDISKSPACE OPS_QUEUE_MAXDISKSPACE ES_QUEUE_MAXFILESIZE \
    OPS_QUEUE_MAXFILESIZE ES_QUEUE_CHECKPOINTINTERVAL OPS_QUEUE_CHECKPOINTINTERVAL

# bug https://bugzilla.redhat.com/show_bug.cgi?id=1437952
# pods unable to be terminated because collector has them busy
if [ -d /var/lib/docker/containers ] ; then
    # If oci-umount is fixed, we can remove this. 
    if [ -n "${VERBOSE:-}" ] ; then
        echo "umounts of dead containers will fail. Ignoring..."
        umount /var/lib/docker/containers/*/shm || :
    else
        umount /var/lib/docker/containers/*/shm > /dev/null 2>&1 || :
    fi
fi

if [[ "${USE_REMOTE_SYSLOG:-}" = "true" ]] ; then
    if [[ $REMOTE_SYSLOG_HOST ]] ; then
        ruby generate_syslog_config.rb
    fi
fi

if [ "${ENABLE_PROMETHEUS_ENDPOINT}" != "true" ] ; then
  echo "INFO: Disabling Prometheus endpoint"
  rm -f ${CFG_DIR}/openshift/input-pre-prometheus-metrics.conf
fi

# Create a directory for log files
mkdir -p /var/log/rsyslog/

issue_deprecation_warnings

if [[ $DEBUG ]] ; then
    exec /usr/sbin/rsyslogd $rsyslogargs > /var/log/rsyslog.log 2>&1
else
    exec /usr/sbin/rsyslogd $rsyslogargs
fi
