#!/usr/bin/bash
# Download latest release of openshift-client, intended for use by ../Makefile

set -e
VERSION=$(curl -s -L  https://openshift-release.svc.ci.openshift.org/api/v1/releasestream/4.7.0-0.ci/latest | jq --raw-output '.name')

DOWNLOAD_URL=$(curl -s -L  https://openshift-release.svc.ci.openshift.org/api/v1/releasestream/4.7.0-0.ci/latest | jq --raw-output '.downloadURL')

NAME="openshift-client-linux-$VERSION.tar.gz"
mkdir -p bin
if [ -f "bin/$NAME" ]; then
    echo "openshift client binary $FILE exists."
    exit 0
fi
pushd bin/
echo "Extracting openshift client binary...."
curl -sSfL "$DOWNLOAD_URL/$NAME" -O > "$NAME"
tar xfz "$NAME" oc
rm "$NAME"
echo "Done"
popd
