
### Steps to send logs to Google Cloud Logging

 1. Create a secret which contains google application credentials (the credentials which will be used to send logs to Google Cloud Logging)
    ```
    oc -n openshift-logging create secret generic gcp-secret --from-file=google-application-credentials.json
    ```

    Where `google-application-credentials.json` has following format
    ```json
    {
      "type": "service_account",
      "project_id": "<gcp-project-id>",
      "private_key_id": "9faaaca18b246b499e4f0654324a57651b82cf65",
      "private_key": "-----BEGIN PRIVATE KEY-----\n ... \n-----END PRIVATE KEY-----\n",
      "client_email": "aos-serviceaccount@openshift-gce-devel.iam.gserviceaccount.com",
      "client_id": "<client-id>",
      "auth_uri": "https://accounts.google.com/o/oauth2/auth",
      "token_uri": "https://oauth2.googleapis.com/token",
      "auth_provider_x509_cert_url": "https://www.googleapis.com/oauth2/v1/certs",
      "client_x509_cert_url": "https://www.googleapis.com/robot/v1/metadata/x509/aos-serviceaccount%40openshift-gce-devel.iam.gserviceaccount.com"
    }

    ```
    Replace `project_id` and `private_key` with real values.

2. Create a `Cluster Logging` instance with vector collector with following yaml.
  ```
  oc apply -f cluster-logging.yaml
  ```
cluster-logging.yaml:
  ```yaml
apiVersion: "logging.openshift.io/v1"
kind: "ClusterLogging"
metadata:
  name: "instance"
  namespace: "openshift-logging"
  annotations:
spec:
  collection:
    type: vector
  managementState: Managed

  ```


3. Create a Cluster Log Forwarder instance with following yaml.

  ```
  oc apply -f cluster-log-forwarder.yaml
  ```
cluster-log-forwarder.yaml
  ```yaml
apiVersion: "logging.openshift.io/v1"
kind: "ClusterLogForwarder"
metadata:
  name: "instance"
  namespace: "openshift-logging"
spec:
  outputs:
    - name: gcp-1
      type: googleCloudLogging
      secret:
        name: gcp-secret
      googleCloudLogging:
        projectId : "openshift-gce-devel"
        logId : "app-gcp"
  pipelines:
    - name: demo-logs
      inputRefs:
        - application
        - infrastructure
        - audit
      outputRefs:
        - gcp-1
  ```

4. Login to google console and check logs

set query as:
```
logName="projects/openshift-gce-devel/logs/app-gcp"
```
the log name should match `logId` set in Cluster Log Forwarder.

![Logs in Google Cloud Logging](./logs-in-gcp.png "logs in google cloud logging")
