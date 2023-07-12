/*
Functional Tests for Log Flow Control

In all the tests, application logs are written in the functional test namespace which are then
collected by vector. This setup takes some time to ship logs to loki. To account for this
time lag, we wait till loki starts receiving logs (using `QueryUntil` method).
Then we try to get logs beyond the set Policy threshold to check if any more logs were received.
Ideally, no logs more than the set threshold (as set in the CLF spec) should be sent to loki server.
We expect loki to receive only a certain number of logs as defined in the CLF Spec Policy, hence
controlling the log flow in logging pipeline.

The following test cases have been covered for the respective components in functional tests:
- (Input) Application:
  - Rate limit all application logs by applying policy at the container level
  - Rate limiting logs by label selector by applying policy at the group policy
  - Rate limit logs by namespace selector

- (Output) Loki:
  - Drop Policy
*/
package flowcontrol
