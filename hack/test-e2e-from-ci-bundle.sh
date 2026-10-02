#!/bin/bash

set -euo pipefail

repo_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$repo_dir"

CLUSTER_LOGGING_OPERATOR_NAMESPACE=${CLUSTER_LOGGING_OPERATOR_NAMESPACE:-openshift-logging}

oc label ns/"${CLUSTER_LOGGING_OPERATOR_NAMESPACE}" openshift.io/cluster-monitoring=true --overwrite
oc label ns/"${CLUSTER_LOGGING_OPERATOR_NAMESPACE}" pod-security.kubernetes.io/enforce=privileged --overwrite
oc label ns/"${CLUSTER_LOGGING_OPERATOR_NAMESPACE}" pod-security.kubernetes.io/audit=privileged --overwrite
oc label ns/"${CLUSTER_LOGGING_OPERATOR_NAMESPACE}" pod-security.kubernetes.io/warn=privileged --overwrite

# Run with the ginkgo CLI so specs can execute in parallel. Ginkgo v2 only
# parallelizes when driven by its own CLI (go test cannot). Suites marked
# `Serial` (all e2e suites except input_selection) still run serially on a
# single process; only parallel-safe suites spread across the worker processes.
# --procs is capped to keep the shared, claimed cluster from being overwhelmed.
GOFLAGS=-mod=mod go run github.com/onsi/ginkgo/v2/ginkgo \
   -p --procs=4 \
   -v --trace --no-color \
   --skip="FlowControl" \
   --poll-progress-after=300s \
   --poll-progress-interval=30s \
   --timeout=90m \
   ./test/e2e/...
