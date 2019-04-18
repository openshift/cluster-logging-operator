#!/bin/sh

kubectl create -f deploy/operator.yaml

while : ; do
  checkIfEmpty=$(kubectl get po | awk '/elasticsearch-operator-.*/{print $3}')
  if [ -z "$checkIfEmpty" ]; then
    echo "No pod found... Actual pod status: $checkIfEmpty"
		kubectl get all
	  exit 1
  fi
  if [ $checkIfEmpty = "Running" ]; then
  	echo "Operator started successfully"
  	break
  elif [ $checkIfEmpty = "ImagePullBackOff" -o $checkIfEmpty = "CrashLoopBackOff" ]; then
	  echo "Failed to deploy Elasticsearch operator"
	  pod=$(kubectl get po | awk '/elasticsearch-operator-.*/{print $1}')
	  kubectl logs -f $pod
	  exit 1
  elif [ $checkIfEmpty = "Error" ]; then
	  echo "Failed to deploy Elasticsearch operator"
	  pod=$(kubectl get po | awk '/elasticsearch-operator-.*/{print $1}')
	  kubectl logs -f $pod
	  exit 1
  else
  	echo "Waiting for Elasticsearch operator pod to spin up: ${checkIfEmpty}"
  	sleep 20s
  fi
done

