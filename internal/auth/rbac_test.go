package auth_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	auth "github.com/openshift/cluster-logging-operator/internal/auth"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	"github.com/openshift/cluster-logging-operator/test"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("NewMetaDataReaderClusterRoleBinding", func() {
	It("should stub a well-formed clusterrolebinding", func() {
		Expect(test.YAMLString(auth.NewMetaDataReaderClusterRoleBinding(constants.OpenshiftNS, "logcollector"))).To(MatchYAML(
			`apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  creationTimestamp: null
  name: metadata-reader-openshift-logging-logcollector
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

	It("should not have owner references", func() {
		crb := auth.NewMetaDataReaderClusterRoleBinding(constants.OpenshiftNS, "logcollector")
		Expect(crb.OwnerReferences).To(BeEmpty())
	})
})

var _ = Describe("ServiceAccount SCC Role & RoleBinding", func() {
	It("should stub a well-formed role", func() {
		Expect(test.YAMLString(auth.NewServiceAccountSCCRole(constants.OpenshiftNS, "my-clf", "my-sa", metav1.OwnerReference{}))).To(MatchYAML(
			`apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  creationTimestamp: null
  name: my-clf-my-sa-scc
  namespace: openshift-logging
rules:
- apiGroups:
    - security.openshift.io
  resourceNames:
    - logging-scc
  resources:
    - securitycontextconstraints
  verbs:
    - use
`))
	})

	It("should stub a well-formed roleBinding", func() {
		Expect(test.YAMLString(auth.NewServiceAccountSCCRoleBinding(constants.OpenshiftNS, "my-clf", "my-sa", metav1.OwnerReference{}))).To(MatchYAML(
			`apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  creationTimestamp: null
  name: my-clf-scc
  namespace: openshift-logging
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: Role
  name: my-clf-my-sa-scc
subjects:
  - kind: ServiceAccount
    name: my-sa
    namespace: openshift-logging
`))
	})

})
