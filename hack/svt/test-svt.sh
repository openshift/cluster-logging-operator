#!/bin/bash

set -eou pipefail

repo_dir="$( cd "$(dirname "$0")" >/dev/null 2>&1 ; pwd -P )/../.."
KUBECONFIG=${KUBECONFIG:-$HOME/.kube/config}
NAMESPACE=${LOGGING_NS:-openshift-logging}
TOT_ITERATIONS=${TOT_ITERATIONS:-3}

SVT_REPO=https://github.com/RH-ematysek/workloads 
SVT_BRANCH=logtest_v45 
PLAYBOOK=logging.yml

#Get SVT dependencies
svt_workdir=/tmp/svt
mkdir -p $svt_workdir ||:
pushd $svt_workdir
  if [ ! -d $svt_workdir/workloads ] ; then
    git clone $SVT_REPO
    cd workloads
    git checkout $SVT_BRANCH
    echo "[orchestration]" > inventory && \
    echo "127.0.0.1 ansible_connection=local" >> inventory
  fi
popd 

#2500 m/s which is the svt benchmark of
#throughput without message loss
TOT_APPS=${TOT_APPS:-1}
MSG_PER_SEC=${MSG_PER_SEC:-2500}
TEST_LENGTH_MIN=${TEST_LENGTH_MIN:-10}
msg_per_minute=$((MSG_PER_SEC * 60))
tot_lines=$((msg_per_minute * TEST_LENGTH_MIN))

label_all_nodes=""
if [ $TOT_APPS -gt 1 ] ; then
  label_all_nodes="true"
fi

cleanup(){
  local return_code="$?"

  set +e
  if [ "${DO_CLEANUP:-true}" == "true" ] ; then
      ${repo_dir}/olm_deploy/scripts/operator-uninstall.sh
      ${repo_dir}/olm_deploy/scripts/catalog-uninstall.sh

      pushd ${repo_dir}/../elasticsearch-operator
        ELASTICSEARCH_OPERATOR_NAMESPACE=openshift-operators-redhat olm_deploy/scripts/catalog-uninstall.sh
        ELASTICSEARCH_OPERATOR_NAMESPACE=openshift-operators-redhat olm_deploy/scripts/operator-uninstall.sh
      popd
  fi
  
  set -e
  exit ${return_code}
}
trap cleanup exit

if [ "${DO_SETUP:-true}" == "true" ] ; then
  pushd ${repo_dir}/../elasticsearch-operator
    # install the catalog containing the elasticsearch operator csv
    ELASTICSEARCH_OPERATOR_NAMESPACE=openshift-operators-redhat olm_deploy/scripts/catalog-deploy.sh
    # install the elasticsearch operator from that catalog
    ELASTICSEARCH_OPERATOR_NAMESPACE=openshift-operators-redhat olm_deploy/scripts/operator-install.sh
  popd
  
  IMAGE_CLUSTER_LOGGING_OPERATOR=${IMAGE_CLUSTER_LOGGING_OPERATOR:-} \
  ${repo_dir}/olm_deploy/scripts/catalog-deploy.sh
  ${repo_dir}/olm_deploy/scripts/operator-install.sh
 
  echo "Deploying ClusterLogForwarder ..."
echo 'apiVersion: "logging.openshift.io/v1"
kind: "ClusterLogForwarder"
metadata:
  name: "instance"
spec:
  pipelines:
  - name: "default-log-store"
    inputRefs:
    - "application"
    outputRefs:
    - "default"
' | oc -n $NAMESPACE create -f -

  echo "Deploying cluster logging"
echo 'apiVersion: "logging.openshift.io/v1"
kind: "ClusterLogging"
metadata:
  name: "instance"
spec:
  managementState: "Managed"
  logStore:
    type: "elasticsearch"
    elasticsearch:
      nodeCount: 3
      redundancyPolicy: "ZeroRedundancy"
      resources:
        request:
          memory: 2Gi
      retentionPolicy:
        application:
          maxAge: 7d
        infra:
          maxAge: 7d
        audit:
          maxAge: 7d
  visualization:
    type: "kibana"
    kibana:
      replicas: 1
  collection:
    logs:
      type: "fluentd"
' | oc -n $NAMESPACE create -f -

fi

echo "Wait for Elasticsearch pod containers to be Running..."
for i in $(seq 300); do
  if [ "$(oc -n $NAMESPACE get pods --ignore-not-found --no-headers -l component=elasticsearch | grep '2/2.*Running' | wc -l)" == "3" ] ; then
    running=true
    break
  else
    sleep 1.0
  fi
done
if [ "${running:-false}" == "false" ] ; then
  echo "Timed out waiting for Elasticsearch pods to start"
  exit 1
fi

echo "Wait for Collector pod containers to be Running..."
tot_nodes=$(oc get nodes --no-headers | wc -l)
running=""
for i in $(seq 300); do
  if [ "$(oc -n $NAMESPACE get pods --ignore-not-found --no-headers -l component=fluentd | grep '1/1.*Running' | wc -l)" == "$tot_nodes" ] ; then
    running=true
    break
  else
    sleep 1.0
  fi
done
if [ "${running:-false}" == "false" ] ; then
  echo "Timed out waiting for Collector pods to start"
  exit 1
fi

echo "Starting SVT test using $MSG_PER_SEC msg/s"
echo "Running a total of $TOT_ITERATIONS times to verify consistancy..."
for (( i=1; i <= $TOT_ITERATIONS; i++ )) ; do
  echo "Running attempt $i" 
  pushd $svt_workdir/workloads
  if WORKLOAD_DIR=$svt_workdir/workloads \
     NUM_LINES=$tot_lines \
     RATE=$msg_per_minute \
     LABEL_ALL_NODES=$label_all_nodes \
     NUM_PROJECTS=$TOT_APPS \
     ansible-playbook -v -i inventory workloads/$PLAYBOOK \
     ; then
     # place to add verification of message sequence and duplication
     echo "PASS attempt $i"
   else
     echo "FAIL attempt $i"
     exit 1
   fi
done
