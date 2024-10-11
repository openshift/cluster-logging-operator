package reconcile_test

import (
	"context"

	security "github.com/openshift/api/security/v1"
	"github.com/openshift/cluster-logging-operator/internal/reconcile"
	"github.com/openshift/cluster-logging-operator/internal/runtime"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	"github.com/google/go-cmp/cmp"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
)

var _ = Describe("reconciling ", func() {

	var (
		anSCC = &security.SecurityContextConstraints{
			TypeMeta: metav1.TypeMeta{
				Kind:       "SecurityContextConstraints",
				APIVersion: security.SchemeGroupVersion.String(),
			},
			ObjectMeta: metav1.ObjectMeta{
				Name: "bar",
			},
			ReadOnlyRootFilesystem: true,
		}
	)

	var _ = DescribeTable("SCC", func(initial *security.SecurityContextConstraints, desired security.SecurityContextConstraints) {

		globalScheme := scheme.Scheme
		Expect(security.Install(globalScheme)).To(Succeed())

		k8sClient := fake.NewFakeClient()
		if initial != nil {
			k8sClient = fake.NewFakeClient(initial)
		}

		Expect(reconcile.SecurityContextConstraints(k8sClient, &desired)).To(Succeed(), "Expect no error reconciling secrets")

		key := client.ObjectKey{Name: desired.Name}
		act := &security.SecurityContextConstraints{}
		Expect(k8sClient.Get(context.TODO(), key, act)).To(Succeed(), "Exp. no error after reconciliation to try and verify")

		act.ResourceVersion = "" //dont care here
		desired.ResourceVersion = ""

		Expect(cmp.Diff(act, &desired)).To(BeEmpty(), "Exp. the spec to be the same")
		Expect(cmp.Diff(act, initial)).To(Not(BeEmpty()), "Exp. the spec to have been updated")
	},
		Entry("when it does not exist should create it", nil, *anSCC),
		Entry("when spec is modified it should revert it", runtime.NewSCC("bar"), *anSCC),
		Entry("when spec is not modified it should do nothing", anSCC, *anSCC),
	)
})
