#!/bin/bash
set -euo pipefail

if [ -n "${DEBUG:-}" ]; then
    set -x
fi

IMAGE_ELASTICSEARCH_OPERATOR=${IMAGE_ELASTICSEARCH_OPERATOR:-quay.io/openshift/origin-elasticsearch-operator:latest}

if [ -n "${IMAGE_FORMAT:-}" ] ; then
  IMAGE_ELASTICSEARCH_OPERATOR=$(sed -e "s,\${component},elasticsearch-operator," <(echo $IMAGE_FORMAT))
fi

KUBECONFIG=${KUBECONFIG:-$HOME/.kube/config}

TEST_NAMESPACE="${TEST_NAMESPACE:-e2e-test-${RANDOM}}"

if oc get project ${TEST_NAMESPACE} > /dev/null 2>&1 ; then
  echo using existing project ${TEST_NAMESPACE}
else
  oc create namespace ${TEST_NAMESPACE}
fi

oc create -n ${TEST_NAMESPACE} -f \
https://raw.githubusercontent.com/coreos/prometheus-operator/master/example/prometheus-operator-crd/prometheusrule.crd.yaml || :
oc create -n ${TEST_NAMESPACE} -f \
https://raw.githubusercontent.com/coreos/prometheus-operator/master/example/prometheus-operator-crd/servicemonitor.crd.yaml || :

if [ "${REMOTE_CLUSTER:-true}" = false ] ; then
  sudo sysctl -w vm.max_map_count=262144 ||:

  manifest=$(mktemp)
  files="01-service-account.yaml 02-role.yaml 03-role-bindings.yaml 05-deployment.yaml"
  pushd manifests;
    for f in ${files}; do
       cat ${f} >> ${manifest};
    done;
  popd
  # update the manifest with the image built by ci
  sed -i "s,quay.io/openshift/origin-elasticsearch-operator:latest,${IMAGE_ELASTICSEARCH_OPERATOR}," ${manifest}
  sed -i "s/namespace: openshift-logging/namespace: ${TEST_NAMESPACE}/g" ${manifest}

  TEST_NAMESPACE=${TEST_NAMESPACE} go test ./test/e2e/... \
    -root=$(pwd) \
    -kubeconfig=${KUBECONFIG} \
    -globalMan manifests/04-crd.yaml \
    -namespacedMan ${manifest} \
    -v \
    -parallel=1 \
    -singleNamespace \
    -timeout 900s
else

  if [ -n "${OPENSHIFT_BUILD_NAMESPACE:-}" -a -n "${IMAGE_FORMAT:-}" ] ; then
    imageprefix=$( echo "$IMAGE_FORMAT" | sed -e 's,/stable:.*$,/,' )
    testimage=${imageprefix}pipeline:src
    testroot=$( pwd )

    # create test secret with kubeconfig for pod
    oc create secret -n ${TEST_NAMESPACE} generic test-secret --from-file=config=${KUBECONFIG}

    testpod=$(mktemp)
    cat test/files/e2e-test-pod.yaml > ${testpod}
    sed -i "s,\${TEST_NAMESPACE},${TEST_NAMESPACE}," ${testpod}
    sed -i "s,\${IMAGE_ELASTICSEARCH_OPERATOR},${IMAGE_ELASTICSEARCH_OPERATOR}," ${testpod}
    sed -i "s,\${IMAGE_E2E_TEST},${testimage}," ${testpod}

    oc create \
      -n ${TEST_NAMESPACE} \
      -f ${testpod}

    echo $KUBECONFIG
    oc project || :
    oc config current-context || :

    # wait for the pod to be ready first...
    oc wait pod/elasticsearch-operator-e2e-test -n ${TEST_NAMESPACE} --for=condition=PodScheduled -o yaml --timeout=300s

    oc wait pod/elasticsearch-operator-e2e-test -n ${TEST_NAMESPACE} --for=condition=Ready --timeout=300s || oc logs elasticsearch-operator-e2e-test -n ${TEST_NAMESPACE}
    oc get pods -n ${TEST_NAMESPACE}

    #oc logs -f elasticsearch-operator-e2e-test -n ${TEST_NAMESPACE}
  else
    echo "Failed to run e2e test"
  fi

  # make sure that pod completely successfully?
fi

oc delete namespace ${TEST_NAMESPACE}
