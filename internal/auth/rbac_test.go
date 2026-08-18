package auth_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	auth "github.com/openshift/cluster-logging-operator/internal/auth"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	"github.com/openshift/cluster-logging-operator/test"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("NewMetaDataReaderClusterRoleBinding", func() {
	It("should stub a well-formed clusterrolebinding", func() {
		Expect(test.YAMLString(auth.NewMetaDataReaderClusterRoleBinding(constants.OpenshiftNS, "logcollector", metav1.OwnerReference{}))).To(MatchYAML(
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
})

var _ = Describe("ServiceAccount SCC Role & RoleBinding", func() {
	It("should stub a well-formed role", func() {
		Expect(test.YAMLString(auth.NewServiceAccountSCCRole(constants.OpenshiftNS, "my-sa", metav1.OwnerReference{}))).To(MatchYAML(
			`apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  creationTimestamp: null
  name: my-sa-scc
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
		Expect(test.YAMLString(auth.NewServiceAccountSCCRoleBinding(constants.OpenshiftNS, "my-sa", "customSA-scc", "customSA", metav1.OwnerReference{}))).To(MatchYAML(
			`apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  creationTimestamp: null
  name: my-sa-scc
  namespace: openshift-logging
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: Role
  name: customSA-scc
subjects:
  - kind: ServiceAccount
    name: customSA
    namespace: openshift-logging
`))
	})

})

var _ = Describe("NewMetricsAuthClusterRoleBinding", func() {
	It("should stub a well-formed system:auth-delegator binding", func() {
		crb := auth.NewMetricsAuthClusterRoleBinding("cluster-logging-logfilesmetricexporter-metrics-auth", "openshift-logging", "logfilesmetricexporter")

		Expect(crb.Name).To(Equal("cluster-logging-logfilesmetricexporter-metrics-auth"))
		Expect(crb.RoleRef.Kind).To(Equal("ClusterRole"))
		Expect(crb.RoleRef.Name).To(Equal("system:auth-delegator"))
		Expect(crb.RoleRef.APIGroup).To(Equal("rbac.authorization.k8s.io"))
		Expect(crb.Subjects).To(HaveLen(1))
		Expect(crb.Subjects[0].Kind).To(Equal("ServiceAccount"))
		Expect(crb.Subjects[0].Name).To(Equal("logfilesmetricexporter"))
		Expect(crb.Subjects[0].Namespace).To(Equal("openshift-logging"))
	})
})
