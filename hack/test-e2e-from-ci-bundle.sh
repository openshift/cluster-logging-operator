#!/bin/bash

set -euo pipefail

repo_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$repo_dir"

CLUSTER_LOGGING_OPERATOR_NAMESPACE=${CLUSTER_LOGGING_OPERATOR_NAMESPACE:-openshift-logging}

oc label ns/"${CLUSTER_LOGGING_OPERATOR_NAMESPACE}" openshift.io/cluster-monitoring=true --overwrite
oc label ns/"${CLUSTER_LOGGING_OPERATOR_NAMESPACE}" pod-security.kubernetes.io/enforce=privileged --overwrite
oc label ns/"${CLUSTER_LOGGING_OPERATOR_NAMESPACE}" pod-security.kubernetes.io/audit=privileged --overwrite
oc label ns/"${CLUSTER_LOGGING_OPERATOR_NAMESPACE}" pod-security.kubernetes.io/warn=privileged --overwrite

GOFLAGS=-mod=mod go test -p 1 -v -timeout=90m ./test/e2e/... \
   -ginkgo.v -ginkgo.trace -ginkgo.no-color \
   -ginkgo.skip="FlowControl" \
   -ginkgo.poll-progress-after=300s \
   -ginkgo.poll-progress-interval=30s
