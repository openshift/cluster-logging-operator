#!/bin/bash

set -euo pipefail

repo_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$repo_dir"

CLUSTER_LOGGING_OPERATOR_NAMESPACE=${CLUSTER_LOGGING_OPERATOR_NAMESPACE:-openshift-logging}

oc label ns/"${CLUSTER_LOGGING_OPERATOR_NAMESPACE}" openshift.io/cluster-monitoring=true --overwrite
oc label ns/"${CLUSTER_LOGGING_OPERATOR_NAMESPACE}" pod-security.kubernetes.io/enforce=privileged --overwrite
oc label ns/"${CLUSTER_LOGGING_OPERATOR_NAMESPACE}" pod-security.kubernetes.io/audit=privileged --overwrite
oc label ns/"${CLUSTER_LOGGING_OPERATOR_NAMESPACE}" pod-security.kubernetes.io/warn=privileged --overwrite

# Only the input_selection suite is parallel-safe; it is driven by the ginkgo
# CLI (Ginkgo v2 only parallelizes when driven by its own CLI; go test cannot)
# with --procs capped to keep the shared, claimed cluster from being
# overwhelmed. Every other suite runs serially: several of them deploy
# cluster-scoped resources (e.g. the lokistack suite's reader ClusterRole and
# LokiStacks) that parallel specs would delete out from under each other, and
# the claimed cluster cannot sustain more than one LokiStack at a time.
GOFLAGS=-mod=mod go run github.com/onsi/ginkgo/v2/ginkgo \
   -v --trace --no-color \
   --skip="FlowControl|InputSelection" \
   --poll-progress-after=300s \
   --poll-progress-interval=30s \
   --timeout=90m \
   ./test/e2e/...

GOFLAGS=-mod=mod go run github.com/onsi/ginkgo/v2/ginkgo \
   -p --procs=4 \
   -v --trace --no-color \
   --poll-progress-after=300s \
   --poll-progress-interval=30s \
   --timeout=90m \
   ./test/e2e/input_selection/
