#!/bin/sh

set -e
MINIMAL_VERSION=1.13
VERSION=$( go version | grep -o '1\.[0-9|\.]*')

function version { echo "$@" | awk -F. '{ printf("%d%03d%03d\n", $1,$2,$3); }'; }

if [ $(version $VERSION) -lt $(version $MINIMAL_VERSION) ]; then
    echo "$VERSION is older than $MINIMAL_VERSION"
    exit 1
fi
