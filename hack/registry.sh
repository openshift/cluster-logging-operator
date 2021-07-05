#!/bin/bash
# Print result of `oc registry info $1`. Activate public route and log in if necessary.

set -e

# This flag is set during CI runs.
if [[ "${OPENSHIFT_CI:-}" == "false" ]]; then
    oc registry info $* && exit 0	# Already works

    if [ "$1" = --public ]; then
        echo "activate public registry route and log in, may take several retries" 1>&2
        oc patch configs.imageregistry.operator.openshift.io/cluster --patch '{"spec":{"defaultRoute":true}}' --type=merge 1>&2
        RETRY=0
        until [ $(( RETRY++ )) -eq 50 ]; do
        REGISTRY=$(oc registry info --public) &&
            test -n "$REGISTRY" &&
            oc registry login --insecure --registry $REGISTRY 1>&2 &&
            echo $REGISTRY &&
            exit 0
        echo "retry registry login.." 1>&2
        sleep 6
        done
    fi
fi
exit 1
