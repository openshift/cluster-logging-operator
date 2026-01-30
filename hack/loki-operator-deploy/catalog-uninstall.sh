#!/bin/sh
set -eou pipefail
current_dir=$(dirname "${BASH_SOURCE[0]}" )
source "${current_dir}/env.sh"
oc delete  --wait --ignore-not-found ${LOKI_OPERATOR_CATALOG} -n ${LOKI_OPERATOR_CATALOG_NAMESPACE}

