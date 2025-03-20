#!/bin/bash

# based on monitoring_common.sh script from standard must-gather

# safeguards
set -o nounset
set -o errexit
set -o pipefail

get_first_ready_prom_pod() {
  readarray -t READY_PROM_PODS < <(
    oc get pods -n openshift-monitoring  -l prometheus=k8s --field-selector=status.phase==Running \
      --no-headers -o custom-columns=":metadata.name"
  )
  echo "${READY_PROM_PODS[0]}"
}
