package drop

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	logging "github.com/openshift/cluster-logging-operator/api/logging/v1"
	"github.com/openshift/cluster-logging-operator/test/matchers"
)

var _ = Describe("drop filter", func() {

	Context("#VRL", func() {
		It("should generate valid VRL for dropping", func() {
			spec := []logging.DropTest{
				{
					DropConditions: []logging.DropCondition{
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
					DropConditions: []logging.DropCondition{
						{
							Field:   ".log_type",
							Matches: "application",
						},
					},
				},
				{
					DropConditions: []logging.DropCondition{
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
			exp, err := MakeDropFilter(&spec)
			Expect(err).To(BeNil())
			Expect(exp).To(matchers.EqualTrimLines(`
!((match(to_string(.kubernetes.namespace_name) ?? "", r'busybox') && !match(to_string(.level) ?? "", r'd.+')) || (match(to_string(.log_type) ?? "", r'application')) || (match(to_string(.kubernetes.container_name) ?? "", r'error|warning') && !match(to_string(.kubernetes.labels.test) ?? "", r'foo')))
`))
		})
	})

})
