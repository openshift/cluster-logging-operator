package api_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/openshift/cluster-logging-operator/must-gather/internal/api"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var _ = Describe("Path", func() {

	It("#NewPath should return a path of the args", func() {
		Expect(api.NewArtifactPath("foo", "bar").String()).To(Equal("foo/bar"))
	})

	It("#Add should return the part to the path", func() {
		Expect(api.NewArtifactPath("foo", "bar").Add("xyz").String()).To(Equal("foo/bar/xyz"))
	})

	Context("#ForResource", func() {

		var (
			path = api.NewArtifactPath("/root")
		)

		It("should use the group and resource", func() {
			Expect(path.ForResource(schema.GroupVersionResource{Group: "apps", Version: "v1", Resource: "deployments"}).String()).To(Equal("/root/apps/deployments"))
		})

		It("should use only the resource when there is no group", func() {
			Expect(path.ForResource(schema.GroupVersionResource{Group: "", Version: "v1", Resource: "deployments"}).String()).To(Equal("/root/deployments"))
		})
	})
})
