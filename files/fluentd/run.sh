#!/bin/bash

CFG_DIR=/etc/fluent/configs.d

fluentdargs="--no-supervisor"
# find the sniffer class file
sniffer=$( gem contents fluent-plugin-elasticsearch|grep elasticsearch_simple_sniffer.rb )
if [ -z "$sniffer" ] ; then
    sniffer=$( rpm -ql rubygem-fluent-plugin-elasticsearch|grep elasticsearch_simple_sniffer.rb )
fi
if [ -n "$sniffer" -a -f "$sniffer" ] ; then
    fluentdargs="$fluentdargs -r $sniffer"
fi

set -e
fluentdargs="--suppress-config-dump $fluentdargs"


issue_deprecation_warnings() {
    : # none at the moment
}

IPADDR4=${NODE_IPV4:-$( /usr/sbin/ip -4 addr show dev eth0 | grep inet | sed -e "s/[ \t]*inet \([0-9.]*\).*/\1/" )}
IPADDR6=${NODE_IPV6:-$( /usr/sbin/ip -6 addr show dev eth0 | grep inet | sed -e "s/[ \t]*inet6 \([a-z0-9::]*\).*/\1/" | grep -v ^fe80 | grep -v ^::1 || echo "")}

export IPADDR4 IPADDR6

# Check bearer_token_file for fluent-plugin-kubernetes_metadata_filter.
if [ ! -s /var/run/secrets/kubernetes.io/serviceaccount/token ] ; then
    echo "ERROR: Bearer_token_file (/var/run/secrets/kubernetes.io/serviceaccount/token) to access the Kubernetes API server is missing or empty."
    exit 1
fi

# If FILE_BUFFER_PATH exists and it is not a directory, mkdir fails with the error.
FILE_BUFFER_PATH=/var/lib/fluentd
mkdir -p $FILE_BUFFER_PATH
FLUENT_CONF=$CFG_DIR/user/fluent.conf
if [ ! -f "$FLUENT_CONF" ] ; then
    echo "ERROR: The configuration $FLUENT_CONF does not exist"
    exit 1
fi

###
# Calculate the max allowed for each output buffer given the number of
# buffer file paths
###

NUM_OUTPUTS=$(grep "path.*'$FILE_BUFFER_PATH" $FLUENT_CONF | wc -l)
if [ $NUM_OUTPUTS -eq 0 ]; then
    # Reset to default single output if log forwarding outputs all invalid
    NUM_OUTPUTS=1
fi

# Get the available disk size.
DF_LIMIT=$(df -B1 $FILE_BUFFER_PATH | grep -v Filesystem | awk '{print $2}')
DF_LIMIT=${DF_LIMIT:-0}
if [ $DF_LIMIT -eq 0 ]; then
    echo "ERROR: No disk space is available for file buffer in $FILE_BUFFER_PATH."
    exit 1
fi

# Default to 15% of disk which is approximately 18G
ALLOWED_PERCENT_OF_DISK=${ALLOWED_PERCENT_OF_DISK:-15}
if [ $ALLOWED_PERCENT_OF_DISK -gt 100 ] || [ $ALLOWED_PERCENT_OF_DISK -le 0 ] ; then
  ALLOWED_PERCENT_OF_DISK=15
  echo ALLOWED_PERCENT_OF_DISK is out of the allowed range. Setting to ${ALLOWED_PERCENT_OF_DISK}%
fi
# Determine allowed total given the number of outputs we have.
ALLOWED_DF_LIMIT=$(expr $DF_LIMIT \* $ALLOWED_PERCENT_OF_DISK / 100) || :

# TOTAL_LIMIT_SIZE per buffer
TOTAL_LIMIT_SIZE=$(expr $ALLOWED_DF_LIMIT / $NUM_OUTPUTS) || :
echo "Setting each total_size_limit for $NUM_OUTPUTS buffers to $TOTAL_LIMIT_SIZE bytes"
export TOTAL_LIMIT_SIZE

##
# Calculate the max number of queued chunks given the size of each chunk
# and the max allowed space per buffer
##
BUFFER_SIZE_LIMIT=$(echo ${BUFFER_SIZE_LIMIT:-8m} |  sed -e "s/[Kk]/*1024/g;s/[Mm]/*1024*1024/g;s/[Gg]/*1024*1024*1024/g;s/i//g" | bc)
BUFFER_QUEUE_LIMIT=$(expr $TOTAL_LIMIT_SIZE / $BUFFER_SIZE_LIMIT)
echo "Setting queued_chunks_limit_size for each buffer to $BUFFER_QUEUE_LIMIT"
export BUFFER_QUEUE_LIMIT
echo "Setting chunk_limit_size for each buffer to $BUFFER_SIZE_LIMIT"
export BUFFER_SIZE_LIMIT

issue_deprecation_warnings

# this should be the last thing before launching fluentd so as not to use
# jemalloc with any other processes
if type -p jemalloc-config > /dev/null 2>&1 ; then
    export LD_PRELOAD=$( jemalloc-config --libdir )/libjemalloc.so.$( jemalloc-config --revision )
    export LD_BIND_NOW=1 # workaround for https://bugzilla.redhat.com/show_bug.cgi?id=1544815
fi
if [ -f /var/log/openshift-apiserver/audit.log.pos ] ; then
  #https://bugzilla.redhat.com/show_bug.cgi?id=1867687
  mv /var/log/openshift-apiserver/audit.log.pos /var/log/oauth-apiserver.audit.log
fi

exec fluentd $fluentdargs
