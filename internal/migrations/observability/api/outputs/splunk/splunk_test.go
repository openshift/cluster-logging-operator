package splunk

import (
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	logging "github.com/openshift/cluster-logging-operator/api/logging/v1"
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	"github.com/openshift/cluster-logging-operator/internal/utils"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("#MapSplunk", func() {
	const secretName = "my-secret"
	var (
		secret *corev1.Secret
		url    = "0.0.0.0:9200"
	)
	BeforeEach(func() {
		secret = &corev1.Secret{
			ObjectMeta: v1.ObjectMeta{
				Name:      secretName,
				Namespace: "foo-space",
			},
			Data: map[string][]byte{
				constants.ClientCertKey:      []byte("cert"),
				constants.ClientPrivateKey:   []byte("privatekey"),
				constants.TrustedCABundleKey: []byte("cabundle"),
				constants.Passphrase:         []byte("pass"),
			},
		}
	})
	It("should map logging.Splunk to obs.Splunk", func() {
		secret.Data[constants.SplunkHECTokenKey] = []byte("hec-token")
		loggingOutSpec := logging.OutputSpec{
			URL: url,
			OutputTypeSpec: logging.OutputTypeSpec{
				Splunk: &logging.Splunk{
					IndexKey: "bar",
				},
			},
			Tuning: &logging.OutputTuningSpec{
				Delivery:         logging.OutputDeliveryModeAtLeastOnce,
				MaxWrite:         utils.GetPtr(resource.MustParse("100m")),
				MinRetryDuration: utils.GetPtr(time.Duration(1)),
				MaxRetryDuration: utils.GetPtr(time.Duration(5)),
			},
		}

		expObsSplunk := &obs.Splunk{
			URLSpec: obs.URLSpec{
				URL: url,
			},
			Index: `{.bar||""}`,
			Tuning: &obs.SplunkTuningSpec{
				BaseOutputTuningSpec: obs.BaseOutputTuningSpec{
					Delivery:         obs.DeliveryModeAtLeastOnce,
					MaxWrite:         utils.GetPtr(resource.MustParse("100m")),
					MinRetryDuration: utils.GetPtr(time.Duration(1)),
					MaxRetryDuration: utils.GetPtr(time.Duration(5)),
				},
			},
			Authentication: &obs.SplunkAuthentication{
				Token: &obs.SecretReference{
					Key:        constants.SplunkHECTokenKey,
					SecretName: secretName,
				},
			},
		}

		Expect(MapSplunk(loggingOutSpec, secret)).To(Equal(expObsSplunk))
	})
})
