package fluentd

// RunScript is the run.sh script for launching the fluentd container process
const RunScript = `
#!/bin/bash

CFG_DIR=/etc/fluent/configs.d

fluentdargs="--umask 0077 --no-supervisor"
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

export PROM_BIND_IP=$(
ruby <<EOF
  require 'ipaddr'
  def get_ip (ipstrs)
	isIPV4 = false
	isIPV6 = false
	for ipstr in ipstrs do
	  begin
		ipaddr = IPAddr.new ipstr.strip
	  rescue => e
		return "#{ipstr} is invalid ip. exception #{e.class} occurred with message #{e.message}"
	  else
		isIPV4 |= ipaddr.ipv4?
		isIPV6 |= ipaddr.ipv6?
	  end
	end
	return "[::]" if isIPV6
	return "0.0.0.0" if isIPV4
	return "invalid-ip"
  end
  puts get_ip("${POD_IPS}".split(","))
EOF
)

echo "POD_IPS: ${POD_IPS}, PROM_BIND_IP: ${PROM_BIND_IP}"



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

# total limit size allowed  per buffer
TOTAL_LIMIT_SIZE_ALLOWED_PER_BUFFER=$(expr $ALLOWED_DF_LIMIT / $NUM_OUTPUTS) || :

TOTAL_LIMIT_SIZE_ALLOWED_PER_BUFFER=${TOTAL_LIMIT_SIZE_ALLOWED_PER_BUFFER:-0}
TOTAL_LIMIT_SIZE_PER_BUFFER=${TOTAL_LIMIT_SIZE_PER_BUFFER:-0}
TOTAL_LIMIT_SIZE_PER_BUFFER=$(echo $TOTAL_LIMIT_SIZE_PER_BUFFER |  sed -e "s/[Kk]/*1024/g;s/[Mm]/*1024*1024/g;s/[Gg]/*1024*1024*1024/g;s/i//g" | bc) || :
if [[ $TOTAL_LIMIT_SIZE_PER_BUFFER -lt $TOTAL_LIMIT_SIZE_ALLOWED_PER_BUFFER ]];
then
   if [[ $TOTAL_LIMIT_SIZE_PER_BUFFER -eq 0 ]]; then
       TOTAL_LIMIT_SIZE_PER_BUFFER=$TOTAL_LIMIT_SIZE_ALLOWED_PER_BUFFER
   fi
else
    echo "Requested buffer size per output $TOTAL_LIMIT_SIZE_PER_BUFFER for $NUM_OUTPUTS buffers exceeds maximum available size  $TOTAL_LIMIT_SIZE_ALLOWED_PER_BUFFER bytes per output"
    TOTAL_LIMIT_SIZE_PER_BUFFER=$TOTAL_LIMIT_SIZE_ALLOWED_PER_BUFFER
fi
echo "Setting each total_size_limit for $NUM_OUTPUTS buffers to $TOTAL_LIMIT_SIZE_PER_BUFFER bytes"
export TOTAL_LIMIT_SIZE_PER_BUFFER

##
# Calculate the max number of queued chunks given the size of each chunk
# and the max allowed space per buffer
##
BUFFER_SIZE_LIMIT=$(echo ${BUFFER_SIZE_LIMIT:-8388608})
BUFFER_QUEUE_LIMIT=$(expr $TOTAL_LIMIT_SIZE_PER_BUFFER / $BUFFER_SIZE_LIMIT)
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

# In case of an update to secure fluentd container, copy the fluentd pos files to their new
# locations under /var/lib/fluentd/pos. Moving old pos files is not possible since /var/log
# is mounted read-only in the secure fluentd container.
#
POS_FILES_DIR=${FILE_BUFFER_PATH}/pos
mkdir -p $POS_FILES_DIR
if [ -f /var/log/openshift-apiserver/audit.log.pos -a ! -f ${POS_FILES_DIR}/oauth-apiserver.audit.log ] ; then
    cp /var/log/openshift-apiserver/audit.log.pos ${POS_FILES_DIR}/oauth-apiserver.audit.log
fi
declare -A POS_FILES_FROM_TO=( [/var/log/audit/audit.log.pos]=${POS_FILES_DIR}/audit.log.pos [/var/log/kube-apiserver/audit.log.pos]=${POS_FILES_DIR}/kube-apiserver.audit.log.pos )
for POS_FILE in es-containers.log.pos journal_pos.json oauth-apiserver.audit.log
do
  POS_FILES_FROM_TO["/var/log/$pos_file"]="${POS_FILES_DIR}/$pos_file"
done
for FROM in "${!POS_FILES_FROM_TO[@]}"
do
    TO=${POS_FILES_FROM_TO[$FROM]}
    if [ -f "$FROM" -a ! -f "$TO" ] ; then
      cp "$FROM" "$TO"
    fi
done

FILE="/var/lib/fluentd/pos/journal_pos.json"

if test -f "$FILE"; then
    echo "$FILE exists, checking if yajl parser able to parse this json file without any error."

    ruby -v /etc/fluent/configs.d/user/cleanInValidJson.rb  $FILE || 
    if [ $? -ne 0 ]; then
      echo "$FILE contains invalid json content so removing it as leads to crashloopbackoff error in fluentd pod"
      rm $FILE
    fi
fi

exec fluentd $fluentdargs

`

const CleanInValidJson = `
#!/usr/bin/env ruby

require 'yajl'
require 'json'


#example pos file where issue was reported - FILE = "/var/lib/fluentd/pos/journal_pos.json"

ARGV.each do |filename|

 input = File.read(filename)

 puts "checking if #{filename} a valid json by calling yajl parser"

 @default_options ||= {:symbolize_keys => false}

 begin
   Yajl::Parser.parse(input, @default_options )
 rescue Yajl::ParseError => e
   raise e.message
 end

end

`
