# Cluster Logging Operator Troubleshooting

## Known issues

## Frequently Asked Questions (FAQs)

1. I've made changes to my `ClusterLogForwarder` (CLF) instance but the collectors are not redeployed/updated. Why is that?
    - There could be an issue with one of the inputs, outputs, or pipelines of the CLF. Check the CLF status in one of 2 ways:
        1. Streamed events of the CLF.
            
            ```$ oc describe clf --show-events=true```

        2. Check the `status` section of the CLF instance `YAML` output.
        
            ```$ oc get clf -oyaml```