package telemetry

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/openshift/cluster-logging-operator/version"
)

func TestCart(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "clo telemetry test")
}

//Test if ServiceMonitor spec is correct or not, also Prometheus Metrics get Registered, Updated, Retrieved properly or not
var _ = Describe("telemetry", func() {

	var (
		data          = NewTD()
		mdir          string
		smYamlFile    string
		CLinfo        = data.CLInfo.M
		CLFinputType  = data.CLFInputType.M
		CLFoutputType = data.CLFOutputType.M
		manifestFile  string
	)

	BeforeEach(func() {

		mdir, _ = os.Getwd()
		mdir = path.Dir(path.Dir(mdir))
		smYamlFile = mdir + "/config/prometheus/monitor.yaml"
		manifestFile = mdir + "/bundle/manifests/cluster-logging-operator-metrics-monitor_monitoring.coreos.com_v1_servicemonitor.yaml"

	})

	JustAfterEach(func() {
		fmt.Printf("monitor.yaml and manifest .yaml %v %v\n", smYamlFile, manifestFile)
	})

	Describe("Testing ServiceMonitor Spec", func() {
		Context("With config/prometheus/monitor.yaml", func() {
			It("Should generate bundle/manifests/cluster-logging-operator-metrics-monitor_monitoring.coreos.com_v1_servicemonitor.yaml spec correctly", func() {
				sm, _ := ioutil.ReadFile(smYamlFile)
				msm, _ := ioutil.ReadFile(manifestFile)
				Expect(sm).To(Equal(msm))
			})
			It("Info metric must have version, managedStatus, healthStatus as default values", func() {
				Expect(CLinfo["version"]).To(Equal(version.Version))
				Expect(CLFinputType["application"]).To(Equal("1"))
				Expect(CLFoutputType["default"]).To(Equal("1"))
			})
		})
	})

	Describe("Testing Registering metrics for prometheus", func() {
		Context("Registering metrics for prometheus", func() {
			It("Show register metrics without throwing any errors", func() {
				err := RegisterMetrics()
				fmt.Printf("RegisterMetrics call returns %v", err)
				Expect(err).Should(BeNil())
			})
		})
	})

	Describe("Testing updating metrics for prometheus", func() {
		Context("Updating metrics for prometheus", func() {
			It("Show metrics updated without throwing any errors", func() {
				CLinfo["version"] = "0.0.2"
				err := UpdateMetrics()
				fmt.Printf("UpdateMetrics call returns %v", err)
				Expect(err).Should(BeNil())
			})

		})

	})
})
