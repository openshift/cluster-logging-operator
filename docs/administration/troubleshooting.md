# Cluster Logging Operator Troubleshooting

## Known issues

### After Upgrading CLO and then EO, CLO doesn't know about `kibana` type
The responsibility of managing Kibana was moved to the EO from CLO as part of the 4.5 release feature work. This included creating a new CRD for Kibana that the EO watches for.

If CLO is upgraded before EO, it will try to create a Kibana CR but the Kibana CRD has not yet been created. This will cause CLO to error out with messages indicating it does not know about the type "Kibana".

If this happens, ensure that EO has been updated to at least 4.5 as well so that the Kibana CRD will be created and then delete your running CLO pod. It will restart without the prior error messages and you will see a new Kibana instance roll out (it will be managed by EO instead).

### Orphaned Resources
Related Jira: https://issues.redhat.com/browse/LOG-3316

When deleting the `openshift-logging` namespace directly, there will be orphaned resources namely `ClusterRoles(CR)`  and `ClusterRoleBindings(CRB)` related to logging.

Originally they were related to `Loki` RBAC CR and CRBs

List of `Loki` CRs and CRBS:
- logging-application-logs-reader
- logging-collector-logs-writer
- logging-all-authenticated-application-logs-reader

The above can be manually removed with
`oc delete`

Upon further inspection other CRs and CRBs are:
- cluster-logging-metadata-reader
- clusterloggings.logging.openshift.io-v1-view
- clusterloggings.logging.openshift.io-v1-edit
- clusterloggings.logging.openshift.io-v1-crdview
- clusterloggings.logging.openshift.io-v1-admin
- clusterlogforwarders.logging.openshift.io-v1-view
- clusterlogforwarders.logging.openshift.io-v1-edit
- clusterlogforwarders.logging.openshift.io-v1-crdview
- clusterlogforwarders.logging.openshift.io-v1-admin

These can be removed with the below commands.
```
oc delete --wait --ignore-not-found crd clusterloggings.logging.openshift.io
oc delete --wait --ignore-not-found crd collectors.logging.openshift.io
oc delete --wait --ignore-not-found crd clusterlogforwarders.logging.openshift.io
```

## Frequently Asked Questions (FAQs)

1. I've made changes to my `ClusterLogForwarder` (CLF) instance but the collectors are not redeployed/updated. Why is that?
    - There could be an issue with one of the inputs, outputs, or pipelines of the CLF. Check the CLF status in one of 2 ways:
        1. Streamed events of the CLF.
            
            ```$ oc describe clf --show-events=true```

        2. Check the `status` section of the CLF instance `YAML` output.
        
            ```$ oc get clf -oyaml```