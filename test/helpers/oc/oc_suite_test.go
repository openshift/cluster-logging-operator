package oc_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	"github.com/onsi/ginkgo/config"
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
`

func TestOC(t *testing.T) {
	RegisterFailHandler(Fail)
	config.DefaultReporterConfig.SlowSpecThreshold = 120
	RunSpecs(t, "Test OC Commands")
}
