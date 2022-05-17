#!/usr/bin/bash

source .bingo/variables.env

set -euo pipefail

BUNDLE_DIR=${2:-"bundle/manifests"}
CLF_CRD_FILE="logging.openshift.io_clusterlogforwarders_crd.yaml"
CLO_CRD_FILE="logging.openshift.io_clusterloggings_crd.yaml"
CLO_PATCH_FILE="crd-v1-clusterloggings-patches.yaml"
CLF_PATCH_FILE="crd-v1-singleton-patch.yaml"
KUSTOMIZATIONS_FILE="kustomization.yaml"
METRICS_SERVICEMONITOR="cluster-logging-operator-metrics-monitor_monitoring.coreos.com_v1_servicemonitor.yaml"
METADATA_READER_CLUSTERROLEBINDING="cluster-logging-metadata-reader_rbac.authorization.k8s.io_v1_clusterrolebinding.yaml"
METADATA_READER_CLUSTERROLE="metadata-reader_rbac.authorization.k8s.io_v1_clusterrole.yaml"
PRIORITY_CLASS="cluster-logging_scheduling.k8s.io_v1_priorityclass.yaml"

BUNDLE_GEN_FLAGS=$1

echo "--------------------------------------------------------------"
echo "Generate CRDs for apiVersion v1"
echo "--------------------------------------------------------------"
$OPERATOR_SDK generate kustomize manifests -q
	$KUSTOMIZE build config/manifests | $OPERATOR_SDK generate bundle $BUNDLE_GEN_FLAGS
rm ${BUNDLE_DIR}/cluster-logging-operator.clusterserviceversion.yaml
mv ${BUNDLE_DIR}/logging.openshift.io_clusterlogforwarders.yaml ${BUNDLE_DIR}/${CLF_CRD_FILE}
mv ${BUNDLE_DIR}/logging.openshift.io_clusterloggings.yaml ${BUNDLE_DIR}/${CLO_CRD_FILE}

cp manifests/patches/${CLO_PATCH_FILE} ${BUNDLE_DIR}
cp manifests/patches/${CLF_PATCH_FILE} ${BUNDLE_DIR}
cp manifests/patches/${KUSTOMIZATIONS_FILE} ${BUNDLE_DIR}

echo "---------------------------------------------------------------"
echo "Kustomize: Patch CRDs for singeltons"
echo "---------------------------------------------------------------"
oc kustomize "${BUNDLE_DIR}" | \
    awk -v clf="${BUNDLE_DIR}/${CLF_CRD_FILE}" \
        -v clo="${BUNDLE_DIR}/${CLO_CRD_FILE}"\
        'BEGIN{filename = clf} /---/ {getline; filename = clo}{print $0> filename}'

cp ${BUNDLE_DIR}/${CLF_CRD_FILE}  manifests/${LOGGING_VERSION}/${CLF_CRD_FILE}
cp ${BUNDLE_DIR}/${CLO_CRD_FILE}  manifests/${LOGGING_VERSION}/${CLO_CRD_FILE}
cp ${BUNDLE_DIR}/${METRICS_SERVICEMONITOR} manifests/${LOGGING_VERSION}/${METRICS_SERVICEMONITOR}
cp ${BUNDLE_DIR}/${METADATA_READER_CLUSTERROLEBINDING} manifests/${LOGGING_VERSION}/${METADATA_READER_CLUSTERROLEBINDING}
cp ${BUNDLE_DIR}/${METADATA_READER_CLUSTERROLE} manifests/${LOGGING_VERSION}/${METADATA_READER_CLUSTERROLE}
cp ${BUNDLE_DIR}/${PRIORITY_CLASS} manifests/${LOGGING_VERSION}/${PRIORITY_CLASS}

echo "---------------------------------------------------------------"
echo "Cleanup operator-sdk generation folder"
echo "---------------------------------------------------------------"
rm -rf deploy
rm ${BUNDLE_DIR}/${CLO_PATCH_FILE}
rm ${BUNDLE_DIR}/${CLF_PATCH_FILE}
rm ${BUNDLE_DIR}/${KUSTOMIZATIONS_FILE}

