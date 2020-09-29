#!/bin/bash
source "$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )/../../../hack/testing-olm/utils"
artifact_dir=$1
GENERATOR_NS=$2
runtime=$(date +%s)
mkdir -p "$artifact_dir/$runtime" ||:
gather_logging_resources "openshift-logging" "$artifact_dir" "$runtime"

oc -n "$GENERATOR_NS" describe deployment/log-generator  > "$artifact_dir/$runtime/log-generator.describe" ||:
oc -n "$GENERATOR_NS" logs deployment/log-generator  > "$artifact_dir/$runtime/log-generator.logs" ||:
oc -n "$GENERATOR_NS" get deployment/log-generator -o yaml > "$artifact_dir/$runtime/log-generator.deployment.yaml" ||:
oc -n "$GENERATOR_NS" get pods -o yaml > "$artifact_dir/$runtime/log-generator.pods.yaml" ||:

oc -n openshift-logging exec $(oc -n openshift-logging get pods -l component=syslog-receiver -o name| sed 's/pod\///') -- tail -n 20000 /var/log/infra.log > "$artifact_dir/$runtime/syslog-receiver.log" ||:
oc -n openshift-logging exec $(oc -n openshift-logging get pods -l component=syslog-receiver -o name| sed 's/pod\///') -- cat /rsyslog/etc/rsyslog.conf > "$artifact_dir/$runtime/syslog-receiver.conf" ||:
oc -n openshift-logging exec $(oc -n openshift-logging get pods -l component=kafka-consumer-clo-topic -o name| sed 's/pod\///') -- tail -n 5000 /shared/consumed.logs > "$artifact_dir/$runtime/kafka-consumer-clo-topic.log" ||:
