#!/bin/bash

source "$(dirname $0)/common"

oc delete ns openshift-logging --force --ignore-not-found --grace-period=0 ||:
oc delete -n openshift is origin-cluster-logging-operator --force --ignore-not-found --grace-period=0 ||:
oc delete -n openshift bc cluster-logging-operator --force --ignore-not-found --grace-period=0||:
