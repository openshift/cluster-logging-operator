#!/bin/bash

set -eou pipefail

repoDir="$(dirname "${BASH_SOURCE[0]}" )/.."

RUN_DURATION=10m


options=""
if [ "${CONF:-}" != "" ]; then
  options="--collector-config=$CONF"
fi

for i in 1000 1500 2000 2500; do 
  echo "Running $i lines/sec"
  $repoDir/bin/functional-benchmarker --run-duration=$RUN_DURATION --lines-per-sec=$i $options &
  pids[${i}]=$!
done

for pid in ${pids[*]}; do
    wait $pid
done
