#!/bin/bash
# Print result of `oc registry info $1`. Activate public route and log in if necessary.

set -e

oc registry info $* && exit 0

if [ "$1" = --public ]; then
    echo "activate public registry route and log in, may take several retries" 1>&2
    oc patch configs.imageregistry.operator.openshift.io/cluster --patch '{"spec":{"defaultRoute":true}}' --type=merge 1>&2
    RETRY=0
    until [ $(( RETRY++ )) -eq 50 ] || oc registry info --public ; do
	echo "retry.." 1>&2
	sleep 6
    done
    [ $(( RETRY++ )) -eq 50 ] && exit 1
    oc registry login --insecure --registry $(oc registry info --public) 1>&2
else
    exit 1
fi
