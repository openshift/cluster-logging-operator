#!/bin/sh 
set -eou pipefail
OPENSHIFT_VERSION=${OPENSHIFT_VERSION:-4.6}
export IMAGE_CLUSTER_LOGGING_OPERATOR_REGISTRY=${IMAGE_CLUSTER_LOGGING_OPERATOR_REGISTRY:-registry.svc.ci.openshift.org/ocp/${OPENSHIFT_VERSION}:cluster-logging-operator-registry}
export IMAGE_CLUSTER_LOGGING_OPERATOR=${IMAGE_CLUSTER_LOGGING_OPERATOR:-registry.svc.ci.openshift.org/ocp/${OPENSHIFT_VERSION}:cluster-logging-operator}
export IMAGE_OAUTH_PROXY=${IMAGE_OAUTH_PROXY:-registry.svc.ci.openshift.org/ocp/${OPENSHIFT_VERSION}:elasticsearch-proxy}
export IMAGE_LOGGING_CURATOR5=${IMAGE_LOGGING_CURATOR5:-registry.svc.ci.openshift.org/ocp/${OPENSHIFT_VERSION}:logging-curator5}
export IMAGE_LOGGING_FLUENTD=${IMAGE_LOGGING_FLUENTD:-registry.svc.ci.openshift.org/ocp/${OPENSHIFT_VERSION}:logging-fluentd}
export IMAGE_PROMTAIL=${IMAGE_PROMTAIL:-registry.svc.ci.openshift.org/ocp/${OPENSHIFT_VERSION}:promtail}
export IMAGE_ELASTICSEARCH6=${IMAGE_ELASTICSEARCH6:-registry.svc.ci.openshift.org/ocp/${OPENSHIFT_VERSION}:logging-elasticsearch6}
export IMAGE_LOGGING_KIBANA6=${IMAGE_LOGGING_KIBANA6:-registry.svc.ci.openshift.org/ocp/${OPENSHIFT_VERSION}:logging-kibana6}

CLUSTER_LOGGING_OPERATOR_NAMESPACE=${CLUSTER_LOGGING_OPERATOR_NAMESPACE:-openshift-logging}

if [ -n "${IMAGE_FORMAT:-}" ] ; then
  export IMAGE_CLUSTER_LOGGING_OPERATOR_REGISTRY=$(echo $IMAGE_FORMAT | sed -e "s,\${component},cluster-logging-operator-registry,")
  export IMAGE_CLUSTER_LOGGING_OPERATOR=$(echo $IMAGE_FORMAT | sed -e "s,\${component},cluster-logging-operator,")
  export IMAGE_OAUTH_PROXY=$(echo $IMAGE_FORMAT | sed -e "s,\${component},oauth-proxy,")
  export IMAGE_LOGGING_CURATOR5=$(echo $IMAGE_FORMAT | sed -e "s,\${component},logging-curator5,")
  export IMAGE_LOGGING_FLUENTD=$(echo $IMAGE_FORMAT | sed -e "s,\${component},logging-fluentd,")
  export IMAGE_PROMTAIL=$(echo $IMAGE_FORMAT | sed -e "s,\${component},promtail,")
  export IMAGE_ELASTICSEARCH6=$(echo $IMAGE_FORMAT | sed -e "s,\${component},logging-elasticsearch6,")
  export IMAGE_LOGGING_KIBANA6=$(echo $IMAGE_FORMAT | sed -e "s,\${component},logging-kibana6,")
fi

echo "Deploying operator catalog with bundle using images: "
echo "cluster logging operator registry: ${IMAGE_CLUSTER_LOGGING_OPERATOR_REGISTRY}"
echo "cluster logging operator: ${IMAGE_CLUSTER_LOGGING_OPERATOR}"
echo "oauth proxy: ${IMAGE_OAUTH_PROXY}"
echo "curator5: ${IMAGE_LOGGING_CURATOR5}"
echo "fluentd: ${IMAGE_LOGGING_FLUENTD}"
echo "promtail: ${IMAGE_PROMTAIL}"
echo "elastic6: ${IMAGE_ELASTICSEARCH6}"
echo "kibana: ${IMAGE_LOGGING_KIBANA6}"

echo "In namespace: ${CLUSTER_LOGGING_OPERATOR_NAMESPACE}"

if oc get project ${CLUSTER_LOGGING_OPERATOR_NAMESPACE} > /dev/null 2>&1 ; then
  echo using existing project ${CLUSTER_LOGGING_OPERATOR_NAMESPACE} for operator catalog deployment
else
  oc create namespace ${CLUSTER_LOGGING_OPERATOR_NAMESPACE}
fi

# substitute image names into the catalog deployment yaml and deploy it
envsubst < olm_deploy/operatorregistry/registry-deployment.yaml | oc create -n ${CLUSTER_LOGGING_OPERATOR_NAMESPACE} -f -
oc wait -n ${CLUSTER_LOGGING_OPERATOR_NAMESPACE} --timeout=120s --for=condition=available deployment/cluster-logging-operator-registry

# create the catalog service
oc create -n ${CLUSTER_LOGGING_OPERATOR_NAMESPACE} -f olm_deploy/operatorregistry/service.yaml

# find the catalog service ip, substitute it into the catalogsource and create the catalog source
export CLUSTER_IP=$(oc get -n ${CLUSTER_LOGGING_OPERATOR_NAMESPACE} service cluster-logging-operator-registry -o jsonpath='{.spec.clusterIP}')
envsubst < olm_deploy/operatorregistry/catalog-source.yaml | oc create -n ${CLUSTER_LOGGING_OPERATOR_NAMESPACE} -f -
