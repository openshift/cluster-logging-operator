#!/bin/bash

source "$(dirname $0)/common"

for repo in ${repo_dir} ${ELASTICSEARCH_OP_REPO}; do
  oc delete -f ${repo}/manifests --ignore-not-found
done
oc label node --all logging-infra-fluentd- || :
oc label node --all logging-infra-rsyslog- || :
