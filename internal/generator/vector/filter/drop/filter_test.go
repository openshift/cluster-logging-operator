package drop

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	"github.com/openshift/cluster-logging-operator/test/matchers"
)

var _ = Describe("drop filter", func() {

	Context("#VRL", func() {
		It("should generate valid VRL for dropping", func() {
			spec := []obs.DropTest{
				{
					DropConditions: []obs.DropCondition{
						{
							Field:   ".kubernetes.namespace_name",
							Matches: "busybox",
						},
						{
							Field:      ".level",
							NotMatches: "d.+",
						},
					},
				},
				{
					DropConditions: []obs.DropCondition{
						{
							Field:   ".log_type",
							Matches: "application",
						},
					},
				},
				{
					DropConditions: []obs.DropCondition{
						{
							Field:   ".kubernetes.container_name",
							Matches: "error|warning",
						},
						{
							Field:      ".kubernetes.labels.test",
							NotMatches: "foo",
						},
					},
				},
			}
			Expect(NewFilter(spec).VRL()).To(matchers.EqualTrimLines(`
!((match(.kubernetes.namespace_name, r'busybox') && !match(.level, r'd.+') ?? false) || (match(.log_type, r'application') ?? false) || (match(.kubernetes.container_name, r'error|warning') && !match(.kubernetes.labels.test, r'foo') ?? false))
`))
		})
	})

})
