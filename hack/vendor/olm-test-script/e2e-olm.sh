#!/usr/bin/env bash

set -euo pipefail

if [ "${DEBUG:-}" = "true" ]; then
	set -x
fi

TEST_NAMESPACE=${TEST_NAMESPACE:-olm-test}
TARGET_NAMESPACE=${TARGET_NAMESPACE:-olm-test}
CREATE_OPERATORGROUP=${CREATE_OPERATORGROUP:-"true"}
OPERATOR_IMAGE=${OPERATOR_IMAGE:-""}

# Get the manifests dir where the package and version dir are.
MANIFEST_DIR=${MANIFEST_DIR:-$(realpath deploy/manifests)}

# Version that your testing. folder where CRDs and CSV exist must match.
VERSION=${VERSION:-4.1}

# indent is used to indent the yaml created from manifest directory correctly
indent() {
  INDENT="      "
  sed "s/^/$INDENT/" | sed "s/^${INDENT}\($1\)/${INDENT:0:-2}- \1/"
}

# TODO: RIPPED from the upstream kube test shell library. Everyone will need
# this. What do we do? -- Thanks pmorie
if [ -t 1 ] ; then
  readonly reset=$(tput sgr0)
  readonly  bold=$(tput bold)
  readonly black=$(tput setaf 0)
  readonly   red=$(tput setaf 1)
  readonly green=$(tput setaf 2)
else
  readonly reset=""
  readonly  bold=""
  readonly black=""
  readonly   red="ERROR "
  readonly green="SUCCESS "
fi

test::object_assert() {
  local tries=$1
  local object=$2
  local request=$3
  local expected=$4
  local args=${5:-}

  for j in $(seq 1 ${tries}); do
    res=$(eval oc get ${args} ${object} -o jsonpath=\"${request}\") || :
    echo $res
    if [[ "${res}" =~ ^$expected$ ]]; then
      echo -n "${green}"
      echo "Successful get ${object} ${request}: ${res}"
      echo -n "${reset}"
      return 0
    fi
    echo "Waiting for Get ${object} ${request} ${args}: expected: ${expected}, got: ${res}"
    sleep $((${j}-1)) || :
  done
  echo "${bold}${red}"
  echo "FAIL!"
  echo "Get ${object} ${request}"
  echo "  Expected: ${expected}"
  echo "  Got:      ${res}"
  echo "${reset}${red}"
  caller
  echo "${reset}"
  return 1
}

# Name of the configmap that we will create
CONFIGMAP_NAME=${CONFIGMAP_NAME:-openshift-olm-test${SUFFIX:-}}

CRD=$(cat $MANIFEST_DIR/$VERSION/*crd.yaml | sed '/^#!.*$/d' | grep -v -- "---" | indent apiVersion)
PKG=$(sed '/^#!.*$/d' $MANIFEST_DIR/*package.yaml | indent packageName)
CSV=$(sed '/^#!.*$/d' $MANIFEST_DIR/$VERSION/*version.yaml | sed 's/namespace: placeholder/namespace: '$TEST_NAMESPACE'/' | sed 's/imagePullPolicy: IfNotPresent/imagePullPolicy: Always/' | grep -v -- "---" |  indent apiVersion)

if [ -n "${OPERATOR_IMAGE:-}" ] ; then
  CSV=$(echo "$CSV" | sed -e "s~containerImage:.*~containerImage: ${OPERATOR_IMAGE}~" | indent apiVersion)
  CSV=$(echo "$CSV" | sed -e "s~image:.*~image: ${OPERATOR_IMAGE}\n~" | indent ApiVersion)
fi

cat > /tmp/configmap.yaml <<EOF | sed 's/^  *$//'
kind: ConfigMap
apiVersion: v1
metadata:
  name: $CONFIGMAP_NAME
data:
  customResourceDefinitions: |-
$CRD
  clusterServiceVersions: |-
$CSV
  packages: |-
$PKG
EOF

CSV_CHANNEL=$(sed -nr 's,.*name: \"?([^"][^"]*)\"?,\1,p' $MANIFEST_DIR/*package.yaml)
CURRENT_CSV=$(sed -nr 's,.*currentCSV: (.*),\1,p' $MANIFEST_DIR/*package.yaml)
PACKAGE_NAME=$(sed -nr 's,.*packageName: (.*),\1,p' $MANIFEST_DIR/*package.yaml)

rc=0
if ! oc create -n $TEST_NAMESPACE -f /tmp/configmap.yaml ; then
  echo Error creating configmap
  exit 1
fi
if [ "${CREATE_OPERATORGROUP}" == "true" ] ; then
  if [ "${GLOBAL:-false}" == "true" ] ; then
    if ! oc -n ${TARGET_NAMESPACE} process -f "$(dirname $0)/operatorgroup-template-global.yaml" -p SUFFIX="${SUFFIX:-}" | oc create -n $TEST_NAMESPACE -f - ; then
      rc=1
    fi
  elif ! oc -n ${TARGET_NAMESPACE} process -f "$(dirname $0)/operatorgroup-template.yaml" -p SUFFIX="${SUFFIX:-}" -p TARGET_NAMESPACE=${TARGET_NAMESPACE} | oc create -n $TEST_NAMESPACE -f - ; then
    rc=1
  fi
  if [ $rc = 1 ] ; then
    echo Error processing operatorgroup template
    exit 1
  fi
fi

if ! oc -n ${TEST_NAMESPACE} process -f "$(dirname $0)/subscription.yaml" -p SUFFIX="${SUFFIX:-}" -p CONFIGMAP_NAME=${CONFIGMAP_NAME:-} -p TEST_NAMESPACE=${TEST_NAMESPACE} -p PACKAGE_NAME=${PACKAGE_NAME} -p STARTING_CSV=${CURRENT_CSV} -p CHANNEL=${CSV_CHANNEL} | oc create -n $TEST_NAMESPACE -f - ; then
  echo "Error processing subscription template"
  exit 1
fi

test::object_assert 100 subscriptions.operators.coreos.com/olm-testing${SUFFIX:-} "{.status.catalogHealth[?(@.catalogSourceRef.name=='openshift-olm-test${SUFFIX:-}')].healthy}" "true" "-n $TEST_NAMESPACE"
# Need to change to match the name of the CSV with version.
test::object_assert 50 clusterserviceversions.operators.coreos.com/${CURRENT_CSV} "{.status.phase}" Succeeded "-n $TEST_NAMESPACE"
