package oc_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

const podSpecTemplate = `
apiVersion: v1
kind: Pod
metadata:
  name: log-generator
  labels:
    component: test
  namespace: %s
spec:
  containers:
    - name: log-generator
      image: quay.io/quay/busybox
      args: ["sh", "-c", "i=0; while true; do echo $i: Test message; i=$((i+1)) ; sleep 1; done"]
  securityContext:
    allowPrivilegeEscalation: false
    capabilities:
      drop:
        - ALL
    runAsNonRoot: true
    seccompProfile:
      type: RuntimeDefault
`

func TestOC(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Test OC Commands")
}
