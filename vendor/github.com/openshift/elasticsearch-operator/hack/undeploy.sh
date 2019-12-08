#!/bin/bash

set -euxo pipefail

oc delete ns $NAMESPACE --force --grace-period=1 ||:
oc delete -n openshift is origin-elasticsearch-operator || :
oc delete -n openshift bc elasticsearch-operator || :
