#!/usr/bin/bash

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
AWS_ACCESS_KEY_ID=${AWS_ACCESS_KEY_ID:-$(aws configure get aws_access_key_id)}
AWS_BUCKET_NAME=${AWS_BUCKET_NAME:-$(whoami)-$(date +'%m%d')}
AWS_PROFILE=${AWS_PROFILE:-openshift-dev}
AWS_SECRET_ACCESS_KEY=${AWS_SECRET_ACCESS_KEY:-$(aws configure get aws_secret_access_key)}
REGION=${REGION:-us-east-1}
LOKI_SECRET_NAME=${LOKI_SECRET_NAME:-loki-s3}
DEPLOY_NAMESPACE=openshift-logging-storage


# create the s3 storage
function prep-s3-storage() {
  if aws s3api get-bucket-location --bucket ${AWS_BUCKET_NAME} > /dev/null 2>&1 ; then
    echo -e "\n s3 bucket \"${AWS_BUCKET_NAME}\" found"
  else
    echo -e "\n creating s3 bucket: ${AWS_BUCKET_NAME}"
    if [[ "${REGION}" = "us-east-1" ]] ; then
      aws s3api create-bucket --acl private --region ${REGION} --bucket ${AWS_BUCKET_NAME}
    else
      aws s3api create-bucket --acl private --region ${REGION} --bucket ${AWS_BUCKET_NAME} \
        --create-bucket-configuration LocationConstraint=${REGION}
    fi
  fi
}

# create storage secret
function create-storage-secret() {
  oc -n ${DEPLOY_NAMESPACE} delete secret --force ${LOKI_SECRET_NAME}||:
  oc -n ${DEPLOY_NAMESPACE} create secret generic ${LOKI_SECRET_NAME} \
    --from-literal=region=${REGION} \
    --from-literal=bucketnames=${AWS_BUCKET_NAME} \
    --from-literal=access_key_id=${AWS_ACCESS_KEY_ID} \
    --from-literal=access_key_secret=${AWS_SECRET_ACCESS_KEY} \
    --from-literal=endpoint=https://s3.${REGION}.amazonaws.com
}

oc apply -k ${SCRIPT_DIR}/operator
oc wait --for=create crd/lokistacks.loki.grafana.com --timeout=30s
prep-s3-storage
oc -n ${DEPLOY_NAMESPACE} wait --for=jsonpath='{.status.phase}'=Running -lapp.kubernetes.io/name=loki-operator pod
oc apply -f ${SCRIPT_DIR}/lokistack/namespace.yaml
create-storage-secret
oc apply -k ${SCRIPT_DIR}/lokistack
