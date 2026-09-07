#!/usr/bin/env bash

git diff --no-ext-diff --quiet -I'^    createdAt: ' bundle/manifests/cluster-logging.clusterserviceversion.yaml
if ((! $?)) ; then
    git checkout bundle/manifests/cluster-logging.clusterserviceversion.yaml
fi
