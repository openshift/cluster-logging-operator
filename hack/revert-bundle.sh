#!/usr/bin/bash

git diff --quiet -I'^    createdAt: ' bundle/manifests/clusterlogging.clusterserviceversion.yaml
if ((! $?)) ; then
    git checkout bundle/manifests/clusterlogging.clusterserviceversion.yaml
fi
