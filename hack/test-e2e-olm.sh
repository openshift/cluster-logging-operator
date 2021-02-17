#!/bin/bash

current_dir=$(dirname "${BASH_SOURCE[0]}" )
source "${current_dir}/lib/init.sh"
source "${current_dir}/lib/util/logs.sh"

pushd ../elasticsearch-operator
# install the catalog containing the elasticsearch operator csv
ELASTICSEARCH_OPERATOR_NAMESPACE=openshift-operators-redhat olm_deploy/scripts/catalog-deploy.sh
# install the elasticsearch operator from that catalog
ELASTICSEARCH_OPERATOR_NAMESPACE=openshift-operators-redhat olm_deploy/scripts/operator-install.sh
popd

for test in $( find "${current_dir}/testing-olm" -type f -name 'test-*.sh' | sort); do
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

pushd ../elasticsearch-operator
# uninstall the elasticsearch operator
ELASTICSEARCH_OPERATOR_NAMESPACE=openshift-operators-redhat ../elasticsearch-operator/olm_deploy/scripts/operator-uninstall.sh
# uninstall the catalog containing the elasticsearch operator csv
ELASTICSEARCH_OPERATOR_NAMESPACE=openshift-operators-redhat ../elasticsearch-operator/olm_deploy/scripts/catalog-uninstall.sh
popd

get_logging_pod_logs

if [[ -n "${failed:-}" ]]; then
    exit 1
fi
