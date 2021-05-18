#!/bin/bash

source "$(dirname $0)/common"

# Delete concurrently to reduce total wait time.
oc delete --wait ns/openshift-logging --ignore-not-found ||: &
oc delete --wait -n openshift is/origin-cluster-logging-operator --ignore-not-found ||: &
oc delete --wait -n openshift bc/cluster-logging-operator --ignore-not-found ||: &
wait
