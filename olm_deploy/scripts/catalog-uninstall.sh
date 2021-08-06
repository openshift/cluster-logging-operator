#!/bin/sh
set -eou pipefail

source $(dirname "${BASH_SOURCE[0]}")/env.sh

oc delete --wait --ignore-not-found project ${CLUSTER_LOGGING_OPERATOR_NAMESPACE}
