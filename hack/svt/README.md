# OpenShift Logging SVT

This package contains test scripts to run the workloads the SVT team uses to performance
test and verify the OpenShift logging stack.

## Requirements
* `podman`
* ansible v2.9
* `oc` OpenShift cli

## Preparation
1. Ensure `oc`, `podman`, `ansible-playbook` are in your PATH
1. Stand up an OpenShift cluster
1. Login to the cluster with a user that has elevated permissions (e.g. `kube:admin`)

## Running a Test
The default test only verifies that no messages are lost from a single application generating
logs and being collected by a single collector instance. The outcome of this test is only PASS 
or FAIL and is the benchmark for the collector and logging's overall throughput. The test runs multiple
iterations to ensure consistency.
```
./test-svt
```
Variations to the defaults can be made by modifying the following:

| Parameter | Default |Description |
|----|----|---|
| `DO_SETUP` | true | Deploy cluster logging |
| `DO_CLEANUP` | true | Remove all artifacts of cluster logging |
| `MSG_PER_SEC` | 2500 | The message rate of each app producing 1k log messages |
| `TOT_APPS` | 1 | The number of applications generating logs |
| `TOT_ITERATIONS` | 3 |The number of successful iteration for the test to PASS |
| `TEST_LENGTH_MIN` | 10 |The length of a test iteration |
 
## Test Matrix 
Following are some of the general variations tested by SVT and their past expectation of PASS/FAIL
| Message / Sec | No. of Applications | Time(min) | Exp. Result |
|----|----|----|----|
| 2500 | 1 | 10 | PASS |
| 3000 | 1 | 10 | FAIL |
| 2500 | 10 | 20 | PASS |
| 5000 | 10 | 30 | PASS |
| 10000 | 10 | 30 | PASS |

