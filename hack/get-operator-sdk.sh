#!/bin/bash
# Download latest release of operator-sdk, intended for use by ../Makefile

set -e
BASE=https://github.com/operator-framework/operator-sdk
VERSION="$(curl -fLsS -o /dev/null -w '%{url_effective}' $BASE/releases/latest | sed 's-.*/--')"
URL=$BASE/releases/download/$VERSION/operator-sdk-$VERSION-$(uname -i)-linux-gnu
echo downloading $URL
curl -fL -o bin/operator-sdk $URL && chmod +x bin/operator-sdk
bin/operator-sdk version > /dev/null # Verify executable works
