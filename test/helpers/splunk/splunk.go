package splunk

import (
	"time"

	"github.com/openshift/cluster-logging-operator/test/framework/functional"
	. "github.com/onsi/gomega"
)

// WaitOnSplunk waits for Splunk to be ready by checking HEC health and service status
func WaitOnSplunk(f *functional.CollectorFunctionalFramework) {
	time.Sleep(20 * time.Second)
	Eventually(func() string {
		// Run the Splunk CLI status command to check if splunkd is running
		output, err := f.SplunkHealthCheck()
		if err != nil {
			return output
		}
		return output
	}, 90*time.Second, 3*time.Second).Should(ContainSubstring("HEC is healthy"))
	time.Sleep(1 * time.Second)
	Eventually(func() string {
		// Run the Splunk CLI status command to check if splunkd is running
		output, err := f.ReadSplunkStatus()
		if err != nil {
			return output
		}
		return output
	}, 15*time.Second, 3*time.Second).Should(SatisfyAll(
		ContainSubstring("splunkd is running"),
		ContainSubstring("splunk helpers are running"),
	))
}