#!/usr/bin/bash

#ensure you are logged into OCP cluster beforehand
#ensure clo operator is installed and CL, CLF instances are running beforehand

export LOGGING_NS=${LOGGING_NS:-openshift-logging}

EXPORTERNAME="logfilesmetricexporter"

declare -a metrics=("log_logged_bytes_total")

lfmepod=`oc get pods -n $LOGGING_NS | grep $EXPORTERNAME | awk 'NR==1{print $1}' `
token=`oc create token prometheus-k8s -n openshift-monitoring`

# ## now loop through the above array
for metricname in "${metrics[@]}"

do 
  echo "$metricname"
  count=`oc exec -n $LOGGING_NS ${lfmepod} -c $EXPORTERNAME -- curl -k -H "Authorization: Bearer ${token}" -s -H "Content-type: application/json" https://$EXPORTERNAME.$LOGGING_NS.svc:2112/metrics | grep -s -c ${metricname}`

  if [[ $count -ge 1 ]]
  then 
    echo "metric found $metricname"
  else 
    echo "metric not found $metricname"
    exit 1
  fi

done