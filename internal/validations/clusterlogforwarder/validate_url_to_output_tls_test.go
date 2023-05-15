package clusterlogforwarder

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	configv1 "github.com/openshift/api/config/v1"
	"github.com/openshift/cluster-logging-operator/apis/logging/v1"
)

var _ = Describe("[internal][validations] ClusterLogForwarder: Output URL vs Output TLS", func() {
	var clf = &v1.ClusterLogForwarder{
		Spec: v1.ClusterLogForwarderSpec{
			Outputs: []v1.OutputSpec{
				{
					Name: "myOutput",
				},
			},
		},
	}

	Context("#validateUrlAccordingToTls", func() {
		It("should fail validation when not secure URL and tls.InsecureSkipVerify=true", func() {
			clf.Spec.Outputs[0].URL = "http://local.svc:514"
			clf.Spec.Outputs[0].TLS = &v1.OutputTLSSpec{
				InsecureSkipVerify: true,
			}
			Expect(validateUrlAccordingToTls(*clf, nil, nil)).To(Not(Succeed()))
		})
		It("should pass validation when not secure URL and no TLS config", func() {
			clf.Spec.Outputs[0].URL = "http://local.svc:514"
			clf.Spec.Outputs[0].TLS = nil
			Expect(validateUrlAccordingToTls(*clf, nil, nil)).To(Succeed())
		})
		It("should pass validation when when not secure URL and tls.InsecureSkipVerify=false", func() {
			clf.Spec.Outputs[0].URL = "http://local.svc:514"
			clf.Spec.Outputs[0].TLS = &v1.OutputTLSSpec{
				InsecureSkipVerify: false,
			}
			Expect(validateUrlAccordingToTls(*clf, nil, nil)).To(Succeed())
		})
		It("should fail validation when not secure URL and exist TLS config: tls.TLSSecurityProfile", func() {
			clf.Spec.Outputs[0].URL = "http://local.svc:514"
			clf.Spec.Outputs[0].TLS = &v1.OutputTLSSpec{
				TLSSecurityProfile: &configv1.TLSSecurityProfile{
					Type: configv1.TLSProfileOldType,
				},
			}
			Expect(validateUrlAccordingToTls(*clf, nil, nil)).To(Not(Succeed()))
		})
		It("should pass validation when secure URL and exist TLS config: tls.InsecureSkipVerify=true", func() {
			clf.Spec.Outputs[0].URL = "https://local.svc:514"
			clf.Spec.Outputs[0].TLS = &v1.OutputTLSSpec{
				InsecureSkipVerify: true,
			}
			Expect(validateUrlAccordingToTls(*clf, nil, nil)).To(Succeed())
		})
		It("should pass pass validation when secure URL and exist TLS config: tls.InsecureSkipVerify=false", func() {
			clf.Spec.Outputs[0].URL = "https://local.svc:514"
			clf.Spec.Outputs[0].TLS = &v1.OutputTLSSpec{
				InsecureSkipVerify: false,
			}
			Expect(validateUrlAccordingToTls(*clf, nil, nil)).To(Succeed())
		})
		It("should pass pass validation when secure URL and exist TLS config: tls.TLSSecurityProfile", func() {
			clf.Spec.Outputs[0].URL = "https://local.svc:514"
			clf.Spec.Outputs[0].TLS = &v1.OutputTLSSpec{
				TLSSecurityProfile: &configv1.TLSSecurityProfile{
					Type: configv1.TLSProfileOldType,
				},
			}
			Expect(validateUrlAccordingToTls(*clf, nil, nil)).To(Succeed())
		})
	})
})
