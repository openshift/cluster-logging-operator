#!/bin/bash
source "$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )/../../../hack/testing/utils"
$artifact_dir=$1
$GENERATOR_NS=$2
get_all_logging_pod_logs $artifact_dir
oc -n openshift-logging exec -c elasticsearch \
	$(oc -n openshift-logging get pods -l component=elasticsearch -o jsonpath={.items[0].metadata.name}) \
	-- indices > $artifact_dir/indices.txt||:
for p in $(oc -n openshift-logging get pods -l component=fluentd -o jsonpath={.items[*].metadata.name}); do
	oc -n openshift-logging exec -- ls -l /var/lib/fluentd/clo_default_output_es > $artifact_dir/$p.buffers.txt||:
	oc -n openshift-logging exec -- ls -l /var/lib/fluentd/retry_clo_default_output_es > $artifact_dir/$p.buffers.retry.txt||:
done
oc -n openshift-logging get configmap fluentd -o jsonpath={.data} > $artifact_dir/fluent-configmap.yaml||:
oc -n openshift-logging get configmap secure-forward -o jsonpath={.data} > $artifact_dir/secure-forward-configmap.yaml||:
oc -n openshift-logging get secret secure-forward -o yaml > $artifact_dir/secure-forward-secret.yaml||:
oc -n openshift-logging extract secret/elasticsearch --to=$artifact_dir||:
oc -n $GENERATOR_NS describe deployment/log-generator  > $artifact_dir/log-generator.describe||:
oc -n $GENERATOR_NS logs deployment/log-generator  > $artifact_dir/log-generator.logs||:
oc -n $GENERATOR_NS get deployment/log-generator -o yaml > $artifact_dir/log-generator.deployment.yaml||:
