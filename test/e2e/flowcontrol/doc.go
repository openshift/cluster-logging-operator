/*
End-to-End Tests for Log Flow Control
LOG-1043: Log Flow Control API Enhancement

In all the tests, we create logstressor pods which generate logs at a certain rate.
We define ClusterLogging and ClusterLoggingForwarder Spec with defined policies. We expect
the rate of logs flowing through the pipeline to be equal to the threshold set in the
CLF Spec policies. To get stable metrics, we take average rate of metrics over a time
window of 2 minutes. So, we wait for 30s after vector metrics first starts showing up on Prometheus
Dashboard.

The following cases have been covered for the respective components in E2E Tests:
- Rate limit application logs (selected by namespace) by applying policy at the group level
- Rate limit application logs (selected by namespace) by applying policy at the container level
- Rate limit logs sent to loki
- Rate limits applied to all application logs (group level) and Loki
*/
package flowcontrol
