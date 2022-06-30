#!/bin/bash
# Print events for the selected namespaces while a command is running.
# Also print event for namespaces created with label test-client.
#
# Usage: oc-watch.sh [NAMESPACE...] -- CMD [ARG...]

echo "$0: $*"

trap 'jobs -p | xargs kill' EXIT		# Kill all background processes on exit

# Watch all events, filter out the interesting ones.
FILTER='(NAMESPACE)|(test-[a-z0-9]+)'
while test -n "$1" && test $1 != "--"; do FILTER="$FILTER|($1)"; shift; done
oc get event -A --watch-only | egrep "^($FILTER) "&

# Execute the watched command. trap will clean up on exit.
env "$@"
