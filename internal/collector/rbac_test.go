package collector_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/openshift/cluster-logging-operator/internal/collector"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	"github.com/openshift/cluster-logging-operator/test"
)

var _ = Describe("NewMetaDataReaderClusterRoleBinding", func() {
	It("should stub a well-formed clusterrolebinding", func() {
		Expect(test.YAMLString(collector.NewMetaDataReaderClusterRoleBinding(constants.OpenshiftNS, "cluster-logging-metadata-reader", "logcollector"))).To(MatchYAML(
			`apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  creationTimestamp: null
  name: cluster-logging-metadata-reader
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: metadata-reader
subjects:
- kind: ServiceAccount
  name: logcollector
  namespace: openshift-logging
`))
	})
})
