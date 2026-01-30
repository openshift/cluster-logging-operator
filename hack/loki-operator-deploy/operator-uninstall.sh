#!/bin/sh
set -eou pipefail
current_dir=$(dirname "${BASH_SOURCE[0]}" )
source "${current_dir}/env.sh"

oc delete Subscription loki-operator -n "${LOKI_OPERATOR_NAMESPACE}"
oc delete ClusterServiceVersion --all -n "${LOKI_OPERATOR_NAMESPACE}"

oc delete --wait --ignore-not-found crd alertingrules.loki.grafana.com
oc delete --wait --ignore-not-found crd lokistacks.loki.grafana.com
oc delete --wait --ignore-not-found crd recordingrules.loki.grafana.com
oc delete --wait --ignore-not-found crd rulerconfigs.loki.grafana.com

oc delete --wait --ignore-not-found clusterrole alertingrules.loki.grafana.com-v1-admin
oc delete --wait --ignore-not-found clusterrole alertingrules.loki.grafana.com-v1-crdview
oc delete --wait --ignore-not-found clusterrole alertingrules.loki.grafana.com-v1-edit 
oc delete --wait --ignore-not-found clusterrole alertingrules.loki.grafana.com-v1-view
oc delete --wait --ignore-not-found clusterrole logging-loki-gateway-authorizer
oc delete --wait --ignore-not-found clusterrole logging-loki-ruler-authorizer
oc delete --wait --ignore-not-found clusterrole loki-operator-metrics-reader
oc delete --wait --ignore-not-found clusterrole lokistacks.loki.grafana.com-v1-admin 
oc delete --wait --ignore-not-found clusterrole lokistacks.loki.grafana.com-v1-crdview
oc delete --wait --ignore-not-found clusterrole lokistacks.loki.grafana.com-v1-edit
oc delete --wait --ignore-not-found clusterrole lokistacks.loki.grafana.com-v1-view
oc delete --wait --ignore-not-found clusterrole recordingrules.loki.grafana.com-v1-admin
oc delete --wait --ignore-not-found clusterrole recordingrules.loki.grafana.com-v1-crdview
oc delete --wait --ignore-not-found clusterrole recordingrules.loki.grafana.com-v1-edit
oc delete --wait --ignore-not-found clusterrole recordingrules.loki.grafana.com-v1-view 
oc delete --wait --ignore-not-found clusterrole rulerconfigs.loki.grafana.com-v1-admin
oc delete --wait --ignore-not-found clusterrole rulerconfigs.loki.grafana.com-v1-crdview 
oc delete --wait --ignore-not-found clusterrole rulerconfigs.loki.grafana.com-v1-edit  
oc delete --wait --ignore-not-found clusterrole rulerconfigs.loki.grafana.com-v1-view
