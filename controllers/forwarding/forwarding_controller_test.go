package forwarding

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	logging "github.com/openshift/cluster-logging-operator/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/internal/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	runtimeclient "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

var _ = Describe("ReconcileForwarder", func() {

	Context("#fetchOrStubClusterLogging", func() {

		var (
			client runtimeclient.WithWatch
		)

		Context("when ClusterLogging is not found", func() {

			DescribeTable("should stub a virtual ClusterLogging resource for reconcilation",
				func(namespace, name string) {
					client = fake.NewClientBuilder().Build()
					controller := ReconcileForwarder{
						Client: client,
					}
					request := ctrl.Request{
						NamespacedName: types.NamespacedName{
							Namespace: namespace,
							Name:      name,
						},
					}

					exp := runtime.NewClusterLogging(namespace, name)
					exp.Spec = logging.ClusterLoggingSpec{
						Collection: &logging.CollectionSpec{
							Type: logging.LogCollectionTypeVector,
						},
					}

					Expect(controller.fetchOrStubClusterLogging(request)).To(BeEquivalentTo(exp))
				},
				Entry("in the openshift-logging namespace and ClusterLogForwarder is not named 'instance'", "openshift-logging", "mine"),
				Entry("in a namespace other then openshift-logging", "trymehere", "mine"),
			)

			It("should fail when in the openshift-logging namespace and ClusterLogForwarder is named 'instance' ", func() {
				client = fake.NewClientBuilder().Build()
				controller := ReconcileForwarder{
					Client: client,
				}
				request := ctrl.Request{
					NamespacedName: types.NamespacedName{
						Namespace: "openshift-logging",
						Name:      "instance",
					},
				}
				_, err := controller.fetchOrStubClusterLogging(request)
				Expect(err).To(Not(BeNil()))
			})

		})

		It("should fetch the ClusterLogging resource with the same namespace and name as the ClusterLogForwarder", func() {
			exp := runtime.NewClusterLogging("somenamespace", "somename")
			exp.Labels = map[string]string{
				"foo": "bar",
			}
			exp.Spec = logging.ClusterLoggingSpec{
				Collection: &logging.CollectionSpec{
					Type: logging.LogCollectionTypeVector,
				},
			}

			clf := runtime.NewClusterLogForwarder(exp.Namespace, exp.Name)

			client = fake.NewClientBuilder().WithRuntimeObjects(exp, clf).Build()
			controller := ReconcileForwarder{
				Client: client,
			}
			request := ctrl.Request{
				NamespacedName: types.NamespacedName{
					Namespace: "somenamespace",
					Name:      "somename",
				},
			}
			Expect(controller.fetchOrStubClusterLogging(request)).To(BeEquivalentTo(exp))
		})

	})
})
