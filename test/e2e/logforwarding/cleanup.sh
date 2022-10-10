#!/bin/bash
source "$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )/../../../hack/testing-olm/utils"
artifact_dir=$1
GENERATOR_NS=$2
LOGGING_NS=${3:-openshift-logging}
runtime=$(date +%s)
mkdir -p "$artifact_dir/$runtime" ||:
gather_logging_resources "$LOGGING_NS" "$artifact_dir" "$runtime"

if [ "$(oc -n "$GENERATOR_NS" describe deployment/log-generator --ignore-not-found --no-headers)" != "" ] ; then
  oc -n "$GENERATOR_NS" describe deployment/log-generator  > "$artifact_dir/$runtime/log-generator.describe" ||:
  oc -n "$GENERATOR_NS" logs deployment/log-generator  > "$artifact_dir/$runtime/log-generator.logs" ||:
  oc -n "$GENERATOR_NS" get deployment/log-generator -o yaml > "$artifact_dir/$runtime/log-generator.deployment.yaml" ||:
  oc -n "$GENERATOR_NS" get pods -o yaml > "$artifact_dir/$runtime/log-generator.pods.yaml" ||:
fi

if [ "$(oc -n $LOGGING_NS get pods -l component=syslog-receiver -o name --ignore-not-found --no-headers)" != "" ] ; then
  pod_name=$(oc -n $LOGGING_NS get pods -l component=syslog-receiver -o name| sed 's/pod\///')
  oc -n $LOGGING_NS exec $pod_name -- tail -n 20000 /var/log/infra.log > "$artifact_dir/$runtime/syslog-receiver.log" ||:
  oc -n $LOGGING_NS exec $pod_name -- cat /rsyslog/etc/rsyslog.conf > "$artifact_dir/$runtime/syslog-receiver.conf" ||:
fi

for pod in $(oc -n $LOGGING_NS pods -llogging-infra=kafka -oname| sed 's/pod\///')
do
    oc -n $LOGGING_NS exec $pod -- tail -n 5000 /shared/consumed.logs > "$artifact_dir/$runtime/$pod.logs" ||:
done
