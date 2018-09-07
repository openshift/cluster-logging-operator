
sleep 10s

while : ; do
  checkIfEmpty=$(kubectl get po | awk '/elastic1-clientdatamaster-.*/{print $3}')
  if [ -z "$checkIfEmpty" ]; then
    echo "No pod found... Actual pod status: $checkIfEmpty"
		kubectl get all
		echo "==== Operator logs ===="
	  pod=$(kubectl get po | awk '/elasticsearch-operator-.*/{print $1}')
	  kubectl logs -f $pod
	  exit 1
  fi
  if [ $checkIfEmpty = "Running" ]; then
  	echo "Elasticsearch started"
  	break
  elif [ $checkIfEmpty = "ImagePullBackOff" -o $checkIfEmpty = "CrashLoopBackOff" ]; then
	  echo "Failed to deploy Elasticsearch"
	  pod=$(kubectl get po | awk '/elastic1-.*/{print $1}')
	  kubectl logs -f $pod
	  exit 1
  elif [ $checkIfEmpty = "Error" ]; then
	  echo "Failed to deploy Elasticsearch"
	  pod=$(kubectl get po | awk '/elastic1-.*/{print $1}')
	  kubectl logs -f $pod
	  exit 1
  else
  	echo "Waiting for Elasticsearch pod to spin up: ${checkIfEmpty}"
  	sleep 20s
  fi
done

while : ; do
  checkIfEmpty=$(kubectl get po | awk '/elastic1-clientdatamaster-.*/{print $2}')
  if [ -z "$checkIfEmpty" ]; then
    echo "No pod found... Actual pod status: $checkIfEmpty"
		kubectl get all
		echo "==== Operator logs ===="
	  pod=$(kubectl get po | awk '/elasticsearch-operator-.*/{print $1}')
	  kubectl logs -f $pod
  	exit 1
  fi
  if [ $checkIfEmpty = "1/1" ]; then
  	echo "Elasticsearch Deployed"
  	break
  elif [  $checkIfEmpty = "ImagePullBackOff" -o $checkIfEmpty = "CrashLoopBackOff" ]; then
  	echo "Failed to deploy Elasticsearch"
	  pod=$(kubectl get po | awk '/elastic1-.*/{print $1}')
	  kubectl logs -f $pod
    exit 1
  else
  	echo "Waiting for Elasticsearch pod to become ready: ${checkIfEmpty}"
  	sleep 20s
  fi
done
