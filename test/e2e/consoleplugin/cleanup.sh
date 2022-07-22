#!/bin/bash
source "$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )/../../../hack/testing-olm/utils"
artifact_dir=$1
GENERATOR_NS=$2
runtime=$(date +%s)
mkdir -p "$artifact_dir/$runtime" ||:
gather_logging_resources "openshift-logging" "$artifact_dir" "$runtime"

name=logging-view-plugin

get_describe() {
    kind=$1; shift
    oc describe $kind/$name $*  > "$artifact_dir/$runtime/$name.$kind.describe" ||:
    oc get $kind/$name $* -o yaml  > "$artifact_dir/$runtime/$name.$kind.yaml" ||:
}

for kind in configmap service deployment; do
    get_describe $kind -n openshift-logging
done
get_describe consoleplugin

kind=deployment
oc -n openshift-logging logs $kind/$name  > "$artifact_dir/$runtime/$name.logs" ||:
