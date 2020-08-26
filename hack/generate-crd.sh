#!/usr/bin/bash

source .bingo/variables.env

set -euo pipefail

MANIFESTS_DIR=${1:-"manifests/${OCP_VERSION}"}
CLF_CRD_FILE="logging.openshift.io_clusterlogforwarders_crd.yaml"
CLO_CRD_FILE="logging.openshift.io_clusterloggings_crd.yaml"

echo "--------------------------------------------------------------"
echo "Generate k8s golang code"
echo "--------------------------------------------------------------"
$OPERATOR_SDK generate k8s

echo "--------------------------------------------------------------"
echo "Generate CRDs for apiVersion v1"
echo "--------------------------------------------------------------"
$OPERATOR_SDK generate crds --crd-version v1
mv "deploy/crds/${CLF_CRD_FILE}" "${MANIFESTS_DIR}"
mv "deploy/crds/${CLO_CRD_FILE}" "${MANIFESTS_DIR}"

echo "---------------------------------------------------------------"
echo "Kustomize: Patch CRDs for singeltons and backward-compatibility"
echo "---------------------------------------------------------------"
oc kustomize "${MANIFESTS_DIR}" | \
    awk -v clf="${MANIFESTS_DIR}/${CLF_CRD_FILE}" \
        -v clo="${MANIFESTS_DIR}/${CLO_CRD_FILE}"\
        'BEGIN{filename = clf} /---/ {getline; filename = clo}{print $0> filename}'

echo "---------------------------------------------------------------"
echo "Cleanup operator-sdk generation folder"
echo "---------------------------------------------------------------"
rm -rf deploy
