package gcl

import (
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	logging "github.com/openshift/cluster-logging-operator/api/logging/v1"
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"

	"github.com/openshift/cluster-logging-operator/internal/constants"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/output/gcl"
	"github.com/openshift/cluster-logging-operator/internal/utils"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("#MapGoogleCloudLogging", func() {
	const secretName = "my-secret"
	var (
		secret *corev1.Secret
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
	It("should map logging.GoogleCloudLogging to obs.GoogleCloudLogging", func() {
		secret.Data[gcl.GoogleApplicationCredentialsKey] = []byte("google.json")

		loggingOutSpec := logging.OutputSpec{
			OutputTypeSpec: logging.OutputTypeSpec{
				GoogleCloudLogging: &logging.GoogleCloudLogging{
					BillingAccountID: "foo",
					LogID:            "baz",
				},
			},
			Tuning: &logging.OutputTuningSpec{
				Delivery:         logging.OutputDeliveryModeAtLeastOnce,
				MaxWrite:         utils.GetPtr(resource.MustParse("100m")),
				MinRetryDuration: utils.GetPtr(time.Duration(1)),
				MaxRetryDuration: utils.GetPtr(time.Duration(5)),
			},
		}

		expObsGCP := &obs.GoogleCloudLogging{
			ID: obs.GoogleCloudLoggingID{
				Type:  obs.GoogleCloudLoggingIDTypeBillingAccount,
				Value: "foo",
			},
			LogID: "baz",
			Authentication: &obs.GoogleCloudLoggingAuthentication{
				Credentials: &obs.SecretReference{
					Key:        gcl.GoogleApplicationCredentialsKey,
					SecretName: secretName,
				},
			},
			Tuning: &obs.GoogleCloudLoggingTuningSpec{
				BaseOutputTuningSpec: obs.BaseOutputTuningSpec{
					Delivery:         obs.DeliveryModeAtLeastOnce,
					MaxWrite:         utils.GetPtr(resource.MustParse("100m")),
					MinRetryDuration: utils.GetPtr(time.Duration(1)),
					MaxRetryDuration: utils.GetPtr(time.Duration(5)),
				},
			},
		}

		Expect(MapGoogleCloudLogging(loggingOutSpec, secret)).To(Equal(expObsGCP))

	})
})
