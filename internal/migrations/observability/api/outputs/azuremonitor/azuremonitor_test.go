package azuremonitor

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	logging "github.com/openshift/cluster-logging-operator/api/logging/v1"
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("#MapAzureMonitor", func() {
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
	It("should map logging.AzureMonitor to obs.AzureMonitor", func() {
		secret.Data[constants.SharedKey] = []byte("shared-key")
		loggingOutSpec := logging.OutputSpec{
			OutputTypeSpec: logging.OutputTypeSpec{
				AzureMonitor: &logging.AzureMonitor{
					CustomerId:      "cust",
					LogType:         "app",
					AzureResourceId: "my-id",
					Host:            "my-host",
				},
			},
			Tuning: &logging.OutputTuningSpec{
				Delivery: logging.OutputDeliveryModeAtLeastOnce,
			},
		}
		expAzMon := &obs.AzureMonitor{
			CustomerId:      "cust",
			LogType:         "app",
			AzureResourceId: "my-id",
			Host:            "my-host",
			Authentication: &obs.AzureMonitorAuthentication{
				SharedKey: &obs.SecretReference{
					Key:        constants.SharedKey,
					SecretName: secretName,
				},
			},
			Tuning: &obs.BaseOutputTuningSpec{
				Delivery: obs.DeliveryModeAtLeastOnce,
			},
		}
		Expect(MapAzureMonitor(loggingOutSpec, secret)).To(Equal(expAzMon))
	})
})
