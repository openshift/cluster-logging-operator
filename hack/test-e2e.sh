#!/bin/bash

current_dir=$(dirname "${BASH_SOURCE[0]}" )
source "${current_dir}/lib/init.sh"

for test in $( find "${current_dir}/testing" -type f -name 'test-*.sh' | sort); do
	os::log::info "======================================"
	os::log::info "running e2e $test "
	os::log::info "======================================"
	if [ "$(basename $test)" == "test-020-olm-upgrade.sh" ] ; then
		os::log::warning "Intentionally skipping $test"
		continue
	fi
	if "${test}" ; then
		os::log::info "e2e $test succeeded at $( date )"
	else
		os::log::warning "e2e $test failed at $( date )"
		failed="true"
	fi
done

if [[ -n "${failed:-}" ]]; then
    exit 1
fi