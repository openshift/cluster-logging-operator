#!/bin/bash

# Credit to Brett Jones the original author of create-aws-cluster.sh
# most of this is taken directly from that script
# Purpose: Provide a simple/default set of instructions for
# easily creating an OpenShift cluster in sts mode (manual credentials mode)
# Includes flags for creating the clf resources necessary for role based SA tokens

set -ueo pipefail

# colors
RED='\033[0;31m'
GREEN='\033[0;32m'
TAN='\033[0;33m'
AQUA='\033[0;36m'
NC='\033[0m'

ME=$(basename "$0")
KERBEROS_USERNAME=${KERBEROS_USERNAME:-$(whoami)}
TMP_DIR=${TMP_DIR:-$HOME/tmp}
CLUSTER_DIR=${CLUSTER_DIR:-${TMP_DIR}/installer}
CCO_UTILITY_DIR=${CCO_UTILITY_DIR:-${TMP_DIR}/cco}
CCO_RESOURCE_NAME=${KERBEROS_USERNAME}-$(date +'%m%d')
CLUSTER_NAME=${KERBEROS_USERNAME}-$(date +'%m%d')cluster$(date +'%H%M')
RELEASE_IMAGE=${RELEASE_IMAGE:-"quay.io/openshift-release-dev/ocp-release:4.9.12-x86_64"}
REGION=${REGION:-"us-east-2"}
SSH_KEY="$(cat $HOME/.ssh/id_ed25519.pub)"
PULL_SECRET="$(tr -d '[:space:]' < $HOME/.docker/config.json)"
ROLE_NAME=${ROLE_NAME:-"role-for-sts"}
AWS_ACCOUNT_ID=${AWS_ACCOUNT_ID:-$(aws sts get-caller-identity --query Account)}

DEBUG='info'
CONFIG_ONLY=${CONFIG_ONLY:-0}

usage() {
	cat <<-EOF

	Deploy a cluster to AWS with STS enabled via manual mode and the CCO utility

	Usage:
	  ${ME} [flags]

	Flags:
      --cleanup            Destroy existing cluster if necessary and exit
  -d, --dir string         Install Assets directory (default "${CLUSTER_DIR}")
	  -c, --config             Create install config only
	  -l, --logging-role       Create logging resources for cloudwatch sts
	  -f, --logforwarding      Create an instance of logforwarding only (assuming secrets already setup)
	  -r, --region             Specify AWS region (default "${REGION}")
	  -d, --debug              Use debug log level for install
	  -h, --help               Help for ${ME}
	EOF
	echo
#  for (( i = 30; i < 38; i++ )); do
#    echo -e "\033[0;"$i"m Normal: (0;$i); \033[1;"$i"m Light: (1;$i)";
#  done
#  echo -e "$NC"
}

main() {
  CLEANUP_ONLY=0
  while [[ $# -gt 0 ]]; do
    key="$1"
    case $key in
      --cleanup)
        CLEANUP_ONLY=1
        shift # past argument
        ;;
			-d|--dir)
				CLUSTER_DIR="$2"
				shift # past argument
				shift # past value
				;;
      -r|--region)
        REGION="$2"
        shift # past argument
        shift # past value
        ;;
      -c|--config)
        CONFIG_ONLY=1
        shift # past argument
        ;;
      -d|--debug)
        DEBUG='debug'
        shift # past argument
        ;;
      -l|--logging-role)
        create_logging_role && exit 0
        ;;
      -f|--logforwarding)
        cluster_log_forwarder && exit 0
        ;;
      -h|--help)
        usage && exit 0
        ;;
      *)
        echo -e "${RED}Unknown flag $1${NC}" > /dev/stderr
        echo
        usage
        exit 1
        ;;
    esac
  done

  if [ "${CLEANUP_ONLY}" -eq 1 ]; then
      destroy_cluster
      exit 0
  fi

  setup

  # exit after setup if config only
  [ $CONFIG_ONLY -eq 1 ] && echo && exit 0

  echo "Extracting credential requests"
  oc adm release extract --credentials-requests --cloud=aws ${RELEASE_IMAGE} \
    --to=${CCO_UTILITY_DIR}/credrequests
  echo
  echo "Creating IAM resources"
  cd ${CCO_UTILITY_DIR}
  ccoctl aws create-all --region=${REGION} \
    --name=${CCO_RESOURCE_NAME} \
    --credentials-requests-dir=${CCO_UTILITY_DIR}/credrequests
  echo -e "Creating installer manifests"
  openshift-install create manifests --dir ${CLUSTER_DIR}
  echo "Copying manifest files to install directory"
  cp ${CCO_UTILITY_DIR}/manifests/* ${CLUSTER_DIR}/manifests/
  echo "Copying the private key"
  cp -a ${CCO_UTILITY_DIR}/tls ${CLUSTER_DIR}

  echo -e "\nSetup is complete and ready to create cluster..."

  create_cluster
}

# Remove existing files and create install config
setup() {
  echo -e "\n${GREEN}Reminder: VPN must be connected before we start the installer${NC}"
  confirm

  if [[ -d ${CLUSTER_DIR} || -d ${CCO_UTILITY_DIR} ]]; then
      echo -e "\n${AQUA}Existing install or cco utility files need to removed from ${TMP_DIR}${NC}"
      confirm
      [ -d ${CLUSTER_DIR} ] && rm -r ${CLUSTER_DIR} && echo "Removing ${CLUSTER_DIR}"
      [ -d ${CCO_UTILITY_DIR} ] && rm -r ${CCO_UTILITY_DIR} && echo "Removing ${CCO_UTILITY_DIR}"
  fi

  make_config
}

make_config() {
  echo -e "\nCreating install config"
  mkdir -p ${CLUSTER_DIR}

	cat <<-EOF > "${CLUSTER_DIR}/install-config.yaml"
	---
	apiVersion: v1
	baseDomain: devcluster.openshift.com
	credentialsMode: Manual
	compute:
	- architecture: amd64
	  hyperthreading: Enabled
	  name: worker
	  platform: {}
	  replicas: 3
	controlPlane:
	  architecture: amd64
	  hyperthreading: Enabled
	  name: master
	  platform: {}
	  replicas: 3
	metadata:
	  creationTimestamp: null
	  name: ${CLUSTER_NAME}
	networking:
	  clusterNetwork:
	  - cidr: 10.128.0.0/14
	    hostPrefix: 23
	  machineNetwork:
	  - cidr: 10.0.0.0/16
	  networkType: OpenShiftSDN
	  serviceNetwork:
	  - 172.30.0.0/16
	platform:
	  aws:
	    region: ${REGION}
	publish: External
	pullSecret: |-
	  ${PULL_SECRET}
	sshKey: |-
	  ${SSH_KEY}
	EOF

  echo "Install config file created at ${CLUSTER_DIR}"
  # create a copy
  cp ${CLUSTER_DIR}/install-config.yaml ${TMP_DIR}/install-config$(date +'%m%d').yaml
  echo "Copy created at ${TMP_DIR}/install-config$(date +'%m%d').yaml"
}

create_cluster() {
  #  just in case
  [ $CONFIG_ONLY -eq 1 ] && exit 0

  confirm

  echo "Creating cluster ${CLUSTER_NAME} at ${CLUSTER_DIR}"
  echo

  if openshift-install create cluster --dir ${CLUSTER_DIR} --log-level=${DEBUG} ; then
    post_install
  else
    _notify_send -t 5000 \
      'FAILED to create cluster' \
      'Errors creating cluster see /.openshift_install.log for details'
    return 1
  fi
}

confirm() {
    echo
    read -p "Do you want to continue (y/N)? " CONT
    if [ "$CONT" != "y" ]; then
      echo "Okay, Exiting."
      exit 0
    fi
    echo
}

create_logging_role() {
  request_dir="credrequests"

  echo "Creating logging credrequest at ${request_dir}"
  mkdir -p ${request_dir}

	cat <<-EOF > ${request_dir}/${ROLE_NAME}-credrequest.yaml
---
apiVersion: cloudcredential.openshift.io/v1
kind: CredentialsRequest
metadata:
  name: ${ROLE_NAME}-credrequest
  namespace: openshift-logging
spec:
  providerSpec:
    apiVersion: cloudcredential.openshift.io/v1
    kind: AWSProviderSpec
    statementEntries:
      - action:
          - logs:PutLogEvents
          - logs:CreateLogGroup
          - logs:PutRetentionPolicy
          - logs:CreateLogStream
          - logs:DescribeLogGroups
          - logs:DescribeLogStreams
        effect: Allow
        resource: arn:aws:logs:*:*:*
  secretRef:
    name: ${ROLE_NAME}
    namespace: openshift-logging
  serviceAccountNames:
    - logcollector
EOF

  echo -e "\nCreating role at AWS and output file for applying our secret"
  ccoctl aws create-iam-roles --name=${CCO_RESOURCE_NAME} --region=${REGION} \
    --credentials-requests-dir=${request_dir} \
    --output-dir=output \
    --identity-provider-arn=arn:aws:iam::${AWS_ACCOUNT_ID}:oidc-provider/${CCO_RESOURCE_NAME}-oidc.s3.${REGION}.amazonaws.com

  echo -e "\nLogging in to cluster"
  export KUBECONFIG=${CLUSTER_DIR}/auth/kubeconfig
  oc login -u kubeadmin -p $(cat ${CLUSTER_DIR}/auth/kubeadmin-password)

  echo -e "\nCreating secret based on new role and OIDC bucket"
  oc create ns openshift-logging
  oc apply -f output/manifests/openshift-logging-${ROLE_NAME}-credentials.yaml

  echo
  exit 0
}

cluster_log_forwarder() {
  echo -e "\nReady to install logfowarding resources"
  confirm

  echo -e "\nApplying secret for ${ROLE_NAME}"
  oc apply -f output/manifests/openshift-logging-${ROLE_NAME}-credentials.yaml

  echo -e "\nCreating logforwarder instance resource file"
  cat <<-EOF > hack/cw-logforwarder-$(date +'%m%d').yaml
---
apiVersion: "logging.openshift.io/v1"
kind: ClusterLogForwarder
metadata:
  name: instance
  namespace: openshift-logging
spec:
  outputs:
    - name: cw
      type: cloudwatch
      cloudwatch:
        groupBy: logType
        groupPrefix: ${CCO_RESOURCE_NAME}
        region: ${REGION}
      secret:
        name: ${ROLE_NAME}
  pipelines:
    - name: all-logs
      inputRefs:
        - infrastructure
        - audit
        - application
      outputRefs:
        - cw
EOF

  echo -e "\nApply logforwarder yaml file"
  oc apply -f hack/cw-logforwarder-$(date +'%m%d').yaml

  echo
  exit 0
}


post_install() {
  _notify_send -t 5000 \
    "OCP cluster ${CLUSTER_NAME} " \
    "Created successfully"

  echo -e "\nSuccessfully deployed cluster!"

  echo -e "\n--- Setup ---\n"
  echo -e "Run: ${AQUA}export KUBECONFIG=${CLUSTER_DIR}/auth/kubeconfig${NC}"
  echo -e "Login with: ${AQUA}oc login -u kubeadmin -p \$(cat ${CLUSTER_DIR}/auth/kubeadmin-password)${NC}"

  echo -e "\n--- Create log forwarding resources for cloudwatch roles (IRSA) ---\n"
  echo -e "Run: ${TAN}${ME} -l (--logging-role)${NC} to create AWS resources and OCP secret"
  echo -e "Run: ${TAN}${ME} -f (--logforwarding)${NC} to create logforwarder instance"

  echo -e "\n--- Cleanup Commands ---\n"
  echo -e "Run cleanup flag: ${GREEN}${ME} --cleanup${NC}"

  echo -e "\nAlternatively, to clean up ccoctl resources after destroying cluster..."
  echo -e "Run: ${GREEN}ccoctl aws delete --name=${CCO_RESOURCE_NAME} --region=${REGION}${NC}"

  echo -e "\nCloudwatch log groups..."
  echo -e "Find using: ${GREEN}aws logs describe-log-groups --query 'logGroups[?starts_with(logGroupName,\`${CCO_RESOURCE_NAME}\`)].logGroupName' --region ${REGION} --output text ${NC}"
  echo -e "To delete (append each log name): ${GREEN}aws logs delete-log-group --region ${REGION} --log-group-name ${CCO_RESOURCE_NAME}- ${NC}"

  echo -e "\n----- Done -----\n"
  exit 0
}

destroy_cluster() {

	echo -e "\nDestroying cluster under ${CLUSTER_DIR}..."
  confirm

	if openshift-install --dir "${CLUSTER_DIR}" destroy cluster; then
		_notify_send -t 5000 \
			'OCP cluster deleted' \
			'Successfully deleted OCP cluster' || :
	else
		_notify_send -t 5000 \
			'FAILED to delete OCP cluster' \
			'FAILURE trying to delete OCP cluster. See log for details'
		return 1
	fi

	cleanup_cco_utility_resources

  echo -e "\n--- Done ---\n"
  exit 0
}

cleanup_cco_utility_resources() {
  echo -e "\nCleaning up cco resources at AWS"

  if ccoctl aws delete --name=${CCO_RESOURCE_NAME} --region=${REGION}; then
    _notify_send -t 5000 \
      'AWS resources deleted' \
      'Successfully deleted OCP AWS resources' || :
  else
    _notify_send -t 5000 \
      'FAILED to delete AWS resources' \
      'FAILURE trying to delete OCP cluster resources'
    return 1
  fi
}

_notify_send() {
	notify-send "$@"
}

# abort with an error message (not currently used)
abort() {
	read -r line func file <<< "$(caller 0)"
	echo -e "${RED}ERROR in $file:$func:$line: $1${NC}" > /dev/stderr
	echo "Bye"
	exit 1
}

# ---
# Never put anything below this line. This is to prevent any partial execution
# if curl ever interrupts the download prematurely. In that case, this script
# will not execute since this is the last line in the script.
err_report() { echo "Error on line $1"; }
trap 'err_report $LINENO' ERR

main "$@"
