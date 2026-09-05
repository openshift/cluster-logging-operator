package utils_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/openshift/cluster-logging-operator/internal/runtime"
	"github.com/openshift/cluster-logging-operator/internal/utils"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("EnsureCanUpdateOwnedResource", func() {
	var (
		clfOwner = metav1.OwnerReference{
			APIVersion: "observability.openshift.io/v1",
			Kind:       "ClusterLogForwarder",
			Name:       "lokistack",
			UID:        "clf-uid",
			Controller: utils.GetPtr(true),
		}
		lokiOwner = metav1.OwnerReference{
			APIVersion: "loki.grafana.com/v1",
			Kind:       "LokiStack",
			Name:       "lokistack",
			UID:        "loki-uid",
			Controller: utils.GetPtr(true),
		}
	)

	It("should allow create when the object is not persisted", func() {
		cm := runtime.NewConfigMap("openshift-logging", "lokistack-config", nil)
		Expect(utils.EnsureCanUpdateOwnedResource(cm, clfOwner)).To(Succeed())
	})

	It("should allow update when owned by the desired owner", func() {
		cm := runtime.NewConfigMap("openshift-logging", "lokistack-config", nil)
		cm.SetResourceVersion("1")
		utils.AddOwnerRefToObject(cm, clfOwner)
		Expect(utils.EnsureCanUpdateOwnedResource(cm, clfOwner)).To(Succeed())
	})

	It("should allow update when neither object has an owner", func() {
		cm := runtime.NewConfigMap("openshift-config-managed", "grafana-dashboard-cluster-logging", nil)
		cm.SetResourceVersion("1")
		Expect(utils.EnsureCanUpdateOwnedResource(cm)).To(Succeed())
	})

	It("should refuse update when owned by another resource", func() {
		cm := runtime.NewConfigMap("openshift-logging", "lokistack-config", nil)
		cm.SetResourceVersion("1")
		utils.AddOwnerRefToObject(cm, lokiOwner)
		err := utils.EnsureCanUpdateOwnedResource(cm, clfOwner)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("refusing to overwrite"))
	})

	It("should refuse update when the object has no owner", func() {
		cm := runtime.NewConfigMap("openshift-logging", "lokistack-config", nil)
		cm.SetResourceVersion("1")
		err := utils.EnsureCanUpdateOwnedResource(cm, clfOwner)
		Expect(err).To(HaveOccurred())
	})
})
