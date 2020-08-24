#!/bin/bash

source "$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )/../../../hack/testing-olm/utils"

artifact_dir="$1"
GENERATOR_NS="$2"
runtime="$(date +%s)"

mkdir -p "$artifact_dir/$runtime"
gather_logging_resources "openshift-logging" "$artifact_dir" "$runtime"

oc -n "$GENERATOR_NS" describe deployment/log-generator  > "$artifact_dir/$runtime/log-generator.describe" ||:
oc -n "$GENERATOR_NS" logs deployment/log-generator  > "$artifact_dir/$runtime/log-generator.logs" ||:
oc -n "$GENERATOR_NS" get deployment/log-generator -o yaml > "$artifact_dir/$runtime/log-generator.deployment.yaml" ||:
