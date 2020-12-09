#!/usr/bin/env bash
#
# Author: Brett Jones <blockloop>
# Purpose: Provide a simple/default set of instructions for easily creating an
# OpenShift cluster for testing purposes

set -ueo pipefail

# colors
RED='\033[0;31m'
NC='\033[0m'

ME=$(basename "$0")
KERBEROS_USERNAME=${KERBEROS_USERNAME:-$(whoami)}
CLUSTER_DIR=${CLUSTER_DIR:-${HOME}/.openshift/cluster}
QUIET=${QUIET:-0}
FORCE=${FORCE:-0}
REGION=us-west-2
SSH_KEY=

usage() {
	cat <<-EOF
	Creates OpenShift clusters with openshift-install using recommended settings

	Usage:
	  ${ME} [flags]

	Flags:
	      --cleanup            Destroy existing cluster if necessary and exit
	  -d, --dir string         Assets directory (default "${CLUSTER_DIR}")
	  -f, --force              Force deletion of any existing clusters if necessary
	  -h, --help               Help for ${ME}
	  -r, --region             Specify AWS region (default "${REGION}")
	  -q, --quiet              Don't send notifications
	      --ssh-key-file       SSH key file to deploy to nodes (default NONE)
	  -u, --user               Your Kerberos username (default from env KERBEROS_USERNAME or \$(whoami))
	EOF
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
			-f|--force)
				FORCE=1
				shift # past argument
				;;
			-h|--help)
				usage && exit 0
				;;
			-q|--quiet)
				QUIET=1
				shift # past argument
				;;
			-r|--region)
				REGION="$2"
				shift # past argument
				shift # past value
				;;
			--ssh-key-file)
				SSH_KEY=$(cat "$2")
				shift # past argument
				shift # past value
				;;
			-u|--user)
				KERBEROS_USERNAME="$2"
				shift # past argument
				shift # past value
				;;
			*)
				echo -e "${RED}Unknown flag $1${NC}" >/dev/stderr
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

	if [ -f "${CLUSTER_DIR}/terraform.tfstate" ]; then
		if [ "${FORCE}" -eq 1 ]; then
			destroy_cluster
		else
			abort "$(cat <<-EOF 
			It appears you already have a running cluster.
			It's possible you forgot to cleanup your previous cluster.
			To force creation of a new cluster and destroy the existing 
			cluster use the -f flag.
			EOF
			)"
		fi
	fi

	make_config
	create_cluster
}

make_config() {
	cat <<-EOF > "${CLUSTER_DIR}/install-config.yaml"
	---
	apiVersion: v1
	baseDomain: devcluster.openshift.com
	compute:
	- architecture: amd64
	  hyperthreading: Enabled
	  name: worker
	  platform:
	    aws:
	      type: m4.2xlarge
	  replicas: 3
	controlPlane:
	  architecture: amd64
	  hyperthreading: Enabled
	  name: master
	  platform: {}
	  replicas: 3
	metadata:
	  creationTimestamp: null
	  name: ${KERBEROS_USERNAME}-$(uuidgen | cut -d- -f1)
	networking:
	  clusterNetwork:
	  - cidr: 10.128.0.0/14
	    hostPrefix: 23
	  machineNetwork:
	  - cidr: 10.0.0.0/16
	  networkType: OpenShiftSDN
	  serviceNetwork:
	  - 172.30.0.0/16
	fips: false
	platform:
	  aws:
	    region: ${REGION}
	publish: External
	pullSecret: |-
	  $(tr -d '[:space:]' < ~/.docker/config.json )
	sshKey: |-
	  ${SSH_KEY}
	EOF
}

destroy_cluster() {
	echo "Destroying cluster under ${CLUSTER_DIR}..."
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
}

create_cluster() {
	echo "Creating cluster ${KERBEROS_USERNAME} cluster under ${CLUSTER_DIR}..."
	if openshift-install --dir "${CLUSTER_DIR}" create cluster ; then 
		_notify_send -t 5000 \
			'OCP cluster created' \
			'Successfully created OCP cluster' || :
	else
		_notify_send -t 5000 \
			'FAILED to create OCP cluster' \
			'FAILURE trying to create OCP cluster. See log for details'
		return 1
	fi
}

_notify_send() {
	[ "${QUIET}" -eq 0 ] && notify-send "$@"
}

# abort with an error message
abort() {
	read -r line func file <<< "$(caller 0)"
	echo -e "${RED}ERROR in $file:$func:$line: $1{NC}" > /dev/stderr
	exit 1
}

# never put anything below this line. This is to prevent any partial execution
# if curl ever interrupts the download prematurely. In that case, this script
# will not execute since this is the last line in the script.
err_report() { echo "Error on line $1"; }
trap 'err_report $LINENO' ERR
main "$@"
