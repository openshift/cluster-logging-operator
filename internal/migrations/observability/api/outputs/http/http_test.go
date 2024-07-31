package http

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

var _ = Describe("#MapHTTP", func() {
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
	It("should map logging.HTTP to obs.HTTP", func() {
		secret.Data[constants.ClientUsername] = []byte("user")
		secret.Data[constants.ClientPassword] = []byte("pass")

		headers := map[string]string{"k1": "v1", "k2": "v2"}

		loggingOutSpec := logging.OutputSpec{
			URL: url,
			OutputTypeSpec: logging.OutputTypeSpec{
				Http: &logging.Http{
					Headers: headers,
					Method:  "POST",
					Timeout: 100,
				},
			},
			Tuning: &logging.OutputTuningSpec{
				Delivery:         logging.OutputDeliveryModeAtLeastOnce,
				MaxWrite:         utils.GetPtr(resource.MustParse("100m")),
				MinRetryDuration: utils.GetPtr(time.Duration(1)),
				MaxRetryDuration: utils.GetPtr(time.Duration(5)),
				Compression:      "gzip",
			},
		}

		expObsHTTP := &obs.HTTP{
			URLSpec: obs.URLSpec{
				URL: url,
			},
			Headers: headers,
			Method:  "POST",
			Timeout: 100,
			Authentication: &obs.HTTPAuthentication{
				Username: &obs.SecretReference{
					Key:        constants.ClientUsername,
					SecretName: secretName,
				},
				Password: &obs.SecretReference{
					Key:        constants.ClientPassword,
					SecretName: secretName,
				},
			},
			Tuning: &obs.HTTPTuningSpec{
				BaseOutputTuningSpec: obs.BaseOutputTuningSpec{
					Delivery:         obs.DeliveryModeAtLeastOnce,
					MaxWrite:         utils.GetPtr(resource.MustParse("100m")),
					MinRetryDuration: utils.GetPtr(time.Duration(1)),
					MaxRetryDuration: utils.GetPtr(time.Duration(5)),
				},
				Compression: "gzip",
			},
		}

		Expect(MapHTTP(loggingOutSpec, secret)).To(Equal(expObsHTTP))
	})
})
