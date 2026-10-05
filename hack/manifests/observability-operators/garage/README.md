# Garage S3 Storage for Loki Operator

Deploys [Garage](https://garagehq.deuxfleurs.fr/) as an S3-compatible object storage backend for testing the Loki Operator on OpenShift.

Uses Garage's `--single-node` and `--default-bucket` flags to automatically configure the cluster layout, create a `loki` bucket, and provision access keys on first startup — no manual setup required.

Furthermore, a secret named `test` will be deployed to allow easy integration with the LokiStack manifests we have upstream

## Deploy

```bash
oc apply -k manifests/garage/
oc wait pod --for=condition=Ready -n openshift-logging -l app.kubernetes.io/name=garage --timeout=120s
```

## Verify

```bash
oc exec -n openshift-logging garage-0 -- /garage bucket list
oc exec -n openshift-logging garage-0 -- /garage key list
```
