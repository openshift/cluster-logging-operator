package logsingestion

import (
	_ "embed"
	"fmt"
	"strings"

	"github.com/openshift/cluster-logging-operator/internal/constants"
	"github.com/openshift/cluster-logging-operator/internal/runtime"
	"github.com/openshift/cluster-logging-operator/internal/utils"
	"github.com/openshift/cluster-logging-operator/test/framework/functional"
	"github.com/openshift/cluster-logging-operator/test/helpers/certificate"
	"github.com/openshift/cluster-logging-operator/test/helpers/mockoon"
	"github.com/openshift/cluster-logging-operator/test/helpers/oc"
	"github.com/openshift/cluster-logging-operator/test/helpers/types"
	v1 "k8s.io/api/core/v1"
)

//go:embed azure-log-ingestion-api.json
var apiConfig string

const (
	apiJsonFile         = "azure-log-ingestion-api.json"
	SecretName          = "azure-client-secret"
	ClientSecretKeyName = "client_secret"
	acme                = "acme.com"
)

// NewMockoonVisitor sets up a Mockoon mock server for the Azure Log Ingestion API.
// It generates TLS certs, deploys the Mockoon container, and configures the collector container
// with env vars to redirect OAuth2 token requests to the mock server.
func NewMockoonVisitor(pb *runtime.PodBuilder, framework *functional.CollectorFunctionalFramework) error {
	// Generate CA and server certificate for acme.com.
	// The CA and server cert must have different Organization names so that OpenSSL
	// can distinguish the issuer from the subject when building the certificate chain.
	ca := certificate.NewCA(nil, "Test CA")
	serverCert := certificate.NewCert(ca, "test", acme)

	configMap := runtime.NewConfigMap(framework.Namespace, mockoon.ContainerName, map[string]string{})
	runtime.NewConfigMapBuilder(configMap).
		Add(apiJsonFile, apiConfig).
		Add("tls.crt", string(serverCert.CertificatePEM())).
		Add("tls.key", string(serverCert.PrivateKeyPEM())).
		Add("ca.crt", string(ca.CertificatePEM()))
	if err := framework.Test.Create(configMap); err != nil {
		return err
	}

	hostAlias := v1.HostAlias{
		IP:        "127.0.0.1",
		Hostnames: []string{acme},
	}

	mountPath := "/data"
	mockoonDataVolume := "mockoon-data"

	pb.AddConfigMapVolume(mockoonDataVolume, mockoon.ContainerName).
		AddHostAlias(hostAlias).
		AddContainer(mockoon.ContainerName, mockoon.Image).
		AddContainerPort(mockoon.ContainerName, mockoon.Port).
		WithCmdArgs([]string{
			fmt.Sprintf("--data=%s/%s", mountPath, apiJsonFile),
			"--log-transaction",
		}).AddVolumeMount(mockoonDataVolume, mountPath, "", true).End()

	// Build a combined CA bundle (system CAs + test CA) at a writable path, and set
	// SSL_CERT_FILE so both OpenSSL (native-tls) and rustls (via rustls-native-certs)
	// trust the Mockoon TLS certificate.
	combinedBundle := "/tmp/ca-bundle.crt"
	pb.GetContainer(constants.CollectorName).
		AddEnvVar("AZURE_AUTHORITY_HOST", fmt.Sprintf("https://%s:%d", acme, mockoon.Port)).
		AddEnvVar("SSL_CERT_FILE", combinedBundle).
		WithCmd([]string{"sh", "-c", fmt.Sprintf(
			"cat /etc/pki/tls/certs/ca-bundle.crt %s/ca.crt > %s && exec /opt/app-root/src/run.sh",
			mountPath, combinedBundle)}).
		AddVolumeMount(mockoonDataVolume, mountPath, "", true).
		Update()

	return nil
}

// ReadApplicationLog reads and parses application logs from the
// Mockoon container used for the Azure Log Ingestion API mock.
func ReadApplicationLog(framework *functional.CollectorFunctionalFramework) ([]types.ApplicationLog, error) {
	output, err := oc.Literal().From("oc logs -n %s pod/%s -c %s", framework.Test.NS.Name, framework.Name, mockoon.ContainerName).Run()
	if err != nil {
		return nil, err
	}
	return extractLogs(output)
}

func extractLogs(output string) ([]types.ApplicationLog, error) {
	logs, err := mockoon.DecodeLogs(output)
	if err != nil {
		return nil, err
	}

	var appLogs []types.ApplicationLog
	for _, log := range logs {
		if log.RequestMethod == "POST" && log.ResponseStatus == 204 &&
			strings.Contains(log.RequestPath, "dataCollectionRules") {
			appLog := log.Transaction.Request.Body
			var tmp []types.ApplicationLog
			if err := types.ParseLogsFrom(utils.ToJsonLogs([]string{appLog}), &tmp, false); err != nil {
				fmt.Printf("error parsing log: %v\n", err)
				continue
			}
			appLogs = append(appLogs, tmp...)
		}
	}

	return appLogs, nil
}
