#!/bin/bash

if [ "${DEBUG:-}" = "true" ]; then
  set -x
fi
set -euo pipefail

current_dir=$(dirname "${BASH_SOURCE[0]}" )
source "${current_dir}/lib/init.sh"
source "${current_dir}/lib/util/logs.sh"

for test in $( find "${current_dir}/testing" -type f -name 'test-*.sh' | sort); do
	os::log::info "==============================================================="
	os::log::info "running e2e $test "
	os::log::info "==============================================================="
	if "${test}" ; then
		os::log::info "==========================================================="
		os::log::info "e2e $test succeeded at $( date )"
		os::log::info "==========================================================="
	else

		os::log::error "============= FAILED FAILED ============= "
		os::log::error "e2e $test failed at $( date )"
		os::log::error "============= FAILED FAILED ============= "
		failed="true"
	fi
done

get_logging_pod_logs

if [[ -n "${failed:-}" ]]; then
    exit 1
fi
