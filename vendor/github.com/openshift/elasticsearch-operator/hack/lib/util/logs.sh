#!/bin/bash

get_logging_pod_logs() {
  local node
  for node in $(oc get nodes -o jsonpath='{.items[*].metadata.name}') ; do
      for logfile in $(oc adm node-logs "$node" --path=fluentd/); do
          oc adm node-logs "$node" --path="fluentd/$logfile" > $ARTIFACT_DIR/$node-$logfile
      done
  done
}
