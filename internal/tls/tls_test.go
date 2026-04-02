package tls_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	configv1 "github.com/openshift/api/config/v1"
	. "github.com/openshift/cluster-logging-operator/internal/tls"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/event"
)

var _ = Describe("#TLSCiphers", func() {
	It("should return the default ciphers when none are defined", func() {
		Expect(TLSCiphers(configv1.TLSProfileSpec{})).To(BeEquivalentTo(DefaultTLSCiphers))
	})
	It("should return the profile ciphers when they are defined", func() {
		Expect(TLSCiphers(configv1.TLSProfileSpec{Ciphers: []string{"a", "b"}})).To(Equal([]string{"a", "b"}))
	})
})

var _ = Describe("#MinTLSVersion", func() {
	It("should return the default min TLS version when not defined", func() {
		Expect(string(DefaultMinTLSVersion)).To(Equal(MinTLSVersion(configv1.TLSProfileSpec{})))
	})
	It("should return the profile min TLS version when defined", func() {
		Expect(string(configv1.VersionTLS13)).To(Equal(MinTLSVersion(configv1.TLSProfileSpec{MinTLSVersion: configv1.VersionTLS13})))
	})
})

var _ = Describe("#IsClusterAPIServer", func() {
	It("should return true for cluster APIServer", func() {
		apiServer := &configv1.APIServer{
			ObjectMeta: metav1.ObjectMeta{
				Name: APIServerName,
			},
		}
		Expect(IsClusterAPIServer(apiServer)).To(BeTrue())
	})

	It("should return false for non-cluster APIServer", func() {
		apiServer := &configv1.APIServer{
			ObjectMeta: metav1.ObjectMeta{
				Name: "not-cluster",
			},
		}
		Expect(IsClusterAPIServer(apiServer)).To(BeFalse())
	})
})

var _ = Describe("#APIServerTLSProfileChangedPredicate", func() {
	Context("with reconcileOnCreate=true", func() {
		It("should allow create events for cluster APIServer", func() {
			pred := APIServerTLSProfileChangedPredicate(true)
			e := event.CreateEvent{
				Object: &configv1.APIServer{
					ObjectMeta: metav1.ObjectMeta{Name: APIServerName},
				},
			}
			Expect(pred.Create(e)).To(BeTrue())
		})

		It("should reject create events for non-cluster APIServer", func() {
			pred := APIServerTLSProfileChangedPredicate(true)
			e := event.CreateEvent{
				Object: &configv1.APIServer{
					ObjectMeta: metav1.ObjectMeta{Name: "other"},
				},
			}
			Expect(pred.Create(e)).To(BeFalse())
		})
	})

	Context("with reconcileOnCreate=false", func() {
		It("should reject create events", func() {
			pred := APIServerTLSProfileChangedPredicate(false)
			e := event.CreateEvent{
				Object: &configv1.APIServer{
					ObjectMeta: metav1.ObjectMeta{Name: APIServerName},
				},
			}
			Expect(pred.Create(e)).To(BeFalse())
		})
	})

	It("should allow update events when TLS profile changed", func() {
		pred := APIServerTLSProfileChangedPredicate(true)
		e := event.UpdateEvent{
			ObjectOld: &configv1.APIServer{
				ObjectMeta: metav1.ObjectMeta{Name: APIServerName},
				Spec: configv1.APIServerSpec{
					TLSSecurityProfile: &configv1.TLSSecurityProfile{
						Type: configv1.TLSProfileIntermediateType,
					},
				},
			},
			ObjectNew: &configv1.APIServer{
				ObjectMeta: metav1.ObjectMeta{Name: APIServerName},
				Spec: configv1.APIServerSpec{
					TLSSecurityProfile: &configv1.TLSSecurityProfile{
						Type: configv1.TLSProfileOldType,
					},
				},
			},
		}
		Expect(pred.Update(e)).To(BeTrue())
	})

	It("should reject update events when TLS profile has not changed", func() {
		pred := APIServerTLSProfileChangedPredicate(true)
		profile := &configv1.TLSSecurityProfile{
			Type: configv1.TLSProfileIntermediateType,
		}
		e := event.UpdateEvent{
			ObjectOld: &configv1.APIServer{
				ObjectMeta: metav1.ObjectMeta{Name: APIServerName},
				Spec:       configv1.APIServerSpec{TLSSecurityProfile: profile},
			},
			ObjectNew: &configv1.APIServer{
				ObjectMeta: metav1.ObjectMeta{Name: APIServerName},
				Spec:       configv1.APIServerSpec{TLSSecurityProfile: profile},
			},
		}
		Expect(pred.Update(e)).To(BeFalse())
	})

	It("should reject delete events", func() {
		pred := APIServerTLSProfileChangedPredicate(true)
		e := event.DeleteEvent{
			Object: &configv1.APIServer{
				ObjectMeta: metav1.ObjectMeta{Name: APIServerName},
			},
		}
		Expect(pred.Delete(e)).To(BeFalse())
	})
})
