package logfilemetricsexporter

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	configv1 "github.com/openshift/api/config/v1"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
)

var _ = Describe("#mapAPIServerToLogFileMetricExporter", func() {
	var reconciler ReconcileLogFileMetricExporter

	It("should return a request for the singleton LogFileMetricExporter when the object is the cluster APIServer", func() {
		apiServer := &configv1.APIServer{
			ObjectMeta: metav1.ObjectMeta{
				Name: "cluster",
			},
		}

		requests := reconciler.mapAPIServerToLogFileMetricExporter(context.TODO(), apiServer)
		Expect(requests).To(Equal([]ctrl.Request{
			{
				NamespacedName: types.NamespacedName{
					Name:      constants.SingletonName,
					Namespace: constants.OpenshiftNS,
				},
			},
		}))
	})

	It("should return nil when the object is not the cluster APIServer", func() {
		apiServer := &configv1.APIServer{
			ObjectMeta: metav1.ObjectMeta{
				Name: "other",
			},
		}

		requests := reconciler.mapAPIServerToLogFileMetricExporter(context.TODO(), apiServer)
		Expect(requests).To(BeNil())
	})

	It("should return nil when the object is not an APIServer", func() {
		obj := &configv1.ClusterVersion{
			ObjectMeta: metav1.ObjectMeta{
				Name: "cluster",
			},
		}

		requests := reconciler.mapAPIServerToLogFileMetricExporter(context.TODO(), obj)
		Expect(requests).To(BeNil())
	})
})
