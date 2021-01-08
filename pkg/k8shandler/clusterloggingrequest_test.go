package k8shandler

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	logging "github.com/openshift/cluster-logging-operator/pkg/apis/logging/v1"
)

var _ = Describe("ClusterLoggingRequest", func() {

	var (
		clr *ClusterLoggingRequest
	)
	BeforeEach(func() {
		clr = &ClusterLoggingRequest{
			Cluster: &logging.ClusterLogging{
				Spec: logging.ClusterLoggingSpec{},
			},
		}

	})
	Describe("#IncludesManagedStorage", func() {
		Context("when logstore", func() {
			Context("is defined", func() {

				It("should return true because we are writing to the managed store", func() {
					clr.Cluster.Spec.LogStore = &logging.LogStoreSpec{}
					Expect(clr.IncludesManagedStorage()).To(BeTrue())
				})

			})
			Context("is not defined", func() {
				It("should return false because there is nowhere to write logs", func() {
					Expect(clr.IncludesManagedStorage()).To(BeFalse())
				})
			})

		})
	})
})
