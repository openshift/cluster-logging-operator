package telemetry

import (
	"os"
	"path"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("ServiceMonitor", func() {
	var (
		smYamlFile   string
		manifestFile string
	)

	BeforeEach(func() {
		mdir, _ := os.Getwd()
		mdir = path.Dir(path.Dir(mdir))
		smYamlFile = mdir + "/config/prometheus/servicemonitor.yaml"
		manifestFile = mdir + "/bundle/manifests/cluster-logging-operator-metrics-monitor_monitoring.coreos.com_v1_servicemonitor.yaml"
	})

	Describe("Testing ServiceMonitor Spec", func() {
		Context("With config/prometheus/servicemonitor.yaml", func() {
			It("Should generate bundle/manifests/cluster-logging-operator-metrics-monitor_monitoring.coreos.com_v1_servicemonitor.yaml spec correctly", func() {
				sm, _ := os.ReadFile(smYamlFile)
				msm, _ := os.ReadFile(manifestFile)
				Expect(sm).To(MatchYAML(msm))
			})
		})
	})
})
