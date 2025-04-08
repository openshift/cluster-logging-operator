package drop

import (
	. "github.com/onsi/ginkgo/v2"
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
!((match(to_string(._internal.kubernetes.namespace_name) ?? "", r'busybox') && !match(to_string(._internal.level) ?? "", r'd.+')) || (match(to_string(._internal.log_type) ?? "", r'application')) || (match(to_string(._internal.kubernetes.container_name) ?? "", r'error|warning') && !match(to_string(._internal.kubernetes.labels.test) ?? "", r'foo')))
`))
		})
	})

})
