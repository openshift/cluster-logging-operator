#!/bin/bash
# Wait for an object to exist, then call 'oc wait' to wait for a condition.
set -e
FOR=$1
NAMESPACE=$2
NAME=$3

WAIT_CMD="oc wait $FOR --timeout=2m -n $NAMESPACE $NAME"
echo $WAIT_CMD

# Show events while we are waiting.
oc get events -n $NAMESPACE --watch-only& trap "kill $!" EXIT
for _ in $(seq 10); do
    if test -n "$(oc get -n $NAMESPACE $NAME --ignore-not-found -o name)"; then
	$WAIT_CMD
	exit $?
    else
	sleep 5
    fi
done
exit 1
