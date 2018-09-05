get_configmap() {
    kubectl -n $NAMESPACE get configmap ${1} -o jsonpath='{.metadata.name}'
}

get_serviceaccount() {
    kubectl -n $NAMESPACE get serviceaccount ${1} -o jsonpath='{.metadata.name}'
}

wait_pod_completion() {
    while kubectl -n $NAMESPACE get po ${1} > /dev/null;
    do
      sleep 5
    done
}
