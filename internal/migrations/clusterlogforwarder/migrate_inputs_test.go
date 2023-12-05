package clusterlogforwarder

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	logging "github.com/openshift/cluster-logging-operator/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/internal/constants"
)

var _ = Describe("migrateInputs", func() {
	Context("for reserved input names", func() {

		It("should stub 'application' as an input when referenced", func() {
			spec := logging.ClusterLogForwarderSpec{
				Pipelines: []logging.PipelineSpec{
					{InputRefs: []string{logging.InputNameApplication}},
				},
			}
			extras := map[string]bool{}
			result, _, _ := MigrateInputs("", "", spec, nil, extras, "", "")
			Expect(result.Inputs).To(HaveLen(1))
			Expect(result.Inputs[0]).To(Equal(logging.InputSpec{Name: logging.InputNameApplication, Application: &logging.Application{}}))
			Expect(extras).To(Equal(map[string]bool{constants.MigrateInputApplication: true}))
		})
		It("should stub 'infrastructure' as an input when referenced", func() {
			spec := logging.ClusterLogForwarderSpec{
				Pipelines: []logging.PipelineSpec{
					{InputRefs: []string{logging.InputNameInfrastructure}},
				},
			}
			extras := map[string]bool{}
			result, _, _ := MigrateInputs("", "", spec, nil, extras, "", "")
			Expect(result.Inputs).To(HaveLen(1))
			Expect(result.Inputs[0]).To(Equal(logging.InputSpec{Name: logging.InputNameInfrastructure, Infrastructure: &logging.Infrastructure{}}))
			Expect(extras).To(Equal(map[string]bool{constants.MigrateInputInfrastructure: true}))
		})
		It("should stub 'audit' as an input when referenced", func() {
			spec := logging.ClusterLogForwarderSpec{
				Pipelines: []logging.PipelineSpec{
					{InputRefs: []string{logging.InputNameAudit}},
				},
			}
			extras := map[string]bool{}
			result, _, _ := MigrateInputs("", "", spec, nil, extras, "", "")
			Expect(result.Inputs).To(HaveLen(1))
			Expect(result.Inputs[0]).To(Equal(logging.InputSpec{Name: logging.InputNameAudit, Audit: &logging.Audit{}}))
			Expect(extras).To(Equal(map[string]bool{constants.MigrateInputAudit: true}))
		})
	})

})
