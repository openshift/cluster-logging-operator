# OLM Test Script

## Description
This script can be used to test install your operator's manifest file in an OCP cluster

### Variables
| *DEBUG* | default false | Set this to true to enable debug output of the e2e-olm script |

| *TEST_NAMESPACE* | default olm-test | This is the namespace where the test subscription, operator group, and csv will be installed to. Note: this will not be created by this script. |

| *MANIFEST_DIR* | default ./deploy/manifests | This is the path to your operator's manifest directory. |

| *VERSION* | default 4.1 | The version directory under $MANIFEST_DIR where your operator's csv file exists. |

### Execution
MANIFEST_DIR=/data/src/github.com/openshift/ansible-service-broker ./e2e-olm.sh


### Debugging

#### If the test fails while checking the status of subscriptions/olm-testing
Remove any previously created objects for this test (if you are using the olm-test namespace you can simply run 'oc delete namespace olm-test && oc create namespace olm-test').

If that does not resolve this, verify that your \*package.yaml file is pointing to the correct csv in your manifest dir.

#### If the test fails while installing your CSV
Your operator may not have been able to start as part of the installation; check the events of your operator pod and adjust your manifest accordingly.
oc get pods -n olm-test
oc describe pod -n olm-test {pod_name}
