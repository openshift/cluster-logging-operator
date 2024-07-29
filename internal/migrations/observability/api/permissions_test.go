package api

import (
	"context"
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	"github.com/openshift/cluster-logging-operator/test"
	rbacv1 "k8s.io/api/rbac/v1"

	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

var _ = Describe("LogCollector Service Account Permissions", func() {
	var (
		k8sClient = fake.NewFakeClient() //nolint
	)
	It("should stub a well-formed clusterrolebinding", func() {
		Expect(test.YAMLString(NewLogCollectorClusterRoleBinding("foo-crb", string(obs.InputTypeApplication)))).To(MatchYAML(
			`apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  creationTimestamp: null
  name: foo-crb
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: collect-application-logs
subjects:
- kind: ServiceAccount
  name: logcollector
  namespace: openshift-logging
`))
	})

	It("should create new ClusterRoleBindings for all input types", func() {
		err := CreateLogCollectorSAPermissions(k8sClient)
		Expect(err).To(BeNil())

		// Check for CRB Creation
		for _, inputType := range obs.ReservedInputTypes.List() {
			name := fmt.Sprintf("%s-collect-%s-logs", constants.CollectorServiceAccountName, inputType)
			current := &rbacv1.ClusterRoleBinding{}
			key := client.ObjectKey{Name: name}
			err := k8sClient.Get(context.TODO(), key, current)
			Expect(err).To(BeNil())
			Expect(current.RoleRef.Name).To(Equal(fmt.Sprintf("collect-%s-logs", inputType)))
			Expect(current.Subjects[0].Name).To(Equal(constants.CollectorServiceAccountName))
		}
	})

})
