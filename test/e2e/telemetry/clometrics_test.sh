#!/usr/bin/bash

#ensure you are logged into OCP cluster beforehand
#ensure clo operator is installed and CL, CLF instances are running beforehand

SECRET=`oc get secret -n openshift-monitoring | grep  prometheus-k8s-token* | head -n 1 | awk '{print $1 }'`

TOKEN=`echo $(oc get secret $SECRET -n openshift-monitoring -o json | jq -r '.data.token') | base64 -d`

THANOS_QUERIER_HOST=`oc get route thanos-querier -n openshift-monitoring -o json | jq -r '.spec.host'`
echo $THANOS_QUERIER_HOST

export LOGGING_NS=${LOGGING_NS:-openshift-logging}





declare -a clometrics=("log_logging_info" "log_collector_error_count_total" "log_forwarder_input_info" "log_forwarder_output_info" "log_forwarder_pipeline_info" "log_file_metric_exporter_info")


## now loop through the above array
for clometricname in "${clometrics[@]}"

do
   echo "$clometricname"
   curl -X GET -kG "https://$THANOS_QUERIER_HOST/api/v1/query?query=${clometricname}" --data-urlencode "query=${clometricname}{namespace='$LOGGING_NS'}" -H "Authorization: Bearer $TOKEN"
   result=`curl -X GET -kG "https://$THANOS_QUERIER_HOST/api/v1/query?query=${clometricname}" --data-urlencode "query=${clometricname}{namespace='$LOGGING_NS'}" -H "Authorization: Bearer $TOKEN" | xargs |  grep -s -c status:success - `
   #check if metric is successfully published in prometheus end point 
   if [[ $result -eq 1 ]]
   then
     echo "CLO metrics target is up in prometheus "
   else 
     echo "CLO metrics target is NOT up in prometheus "
   exit 0
   fi

done

