package loki

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

var _ = Describe("#MapLoki", func() {
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
	It("should map logging.Loki to obs.Loki", func() {
		secret.Data[constants.BearerTokenFileKey] = []byte("token")
		loggingOutSpec := logging.OutputSpec{
			URL: url,
			OutputTypeSpec: logging.OutputTypeSpec{
				Loki: &logging.Loki{
					TenantKey: "app",
					LabelKeys: []string{"foo", "bar"},
				},
			},
			Tuning: &logging.OutputTuningSpec{
				Delivery:         logging.OutputDeliveryModeAtLeastOnce,
				MaxWrite:         utils.GetPtr(resource.MustParse("100m")),
				MinRetryDuration: utils.GetPtr(time.Duration(1)),
				MaxRetryDuration: utils.GetPtr(time.Duration(5)),
				Compression:      "snappy",
			},
		}
		expObsLoki := &obs.Loki{
			URLSpec: obs.URLSpec{
				URL: url,
			},
			TenantKey: `{.app||"none"}`,
			LabelKeys: []string{"foo", "bar"},
			Tuning: &obs.LokiTuningSpec{
				BaseOutputTuningSpec: obs.BaseOutputTuningSpec{
					Delivery:         obs.DeliveryModeAtLeastOnce,
					MaxWrite:         utils.GetPtr(resource.MustParse("100m")),
					MinRetryDuration: utils.GetPtr(time.Duration(1)),
					MaxRetryDuration: utils.GetPtr(time.Duration(5)),
				},
				Compression: "snappy",
			},
			Authentication: &obs.HTTPAuthentication{
				Token: &obs.BearerToken{
					From: obs.BearerTokenFromSecret,
					Secret: &obs.BearerTokenSecretKey{
						Name: secretName,
						Key:  constants.BearerTokenFileKey,
					},
				},
			},
		}

		Expect(MapLoki(loggingOutSpec, secret)).To(Equal(expObsLoki))
	})
})
