package gcl

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	_ "embed"
	"encoding/json"
	"encoding/pem"
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

//go:embed gcl-logging-api.json
var apiConfig string

const (
	apiJsonFile = "gcl-logging-api.json"
	SecretName  = "gcl-secret"
	GCLDomain   = "logging.googleapis.com"
	GCLPort     = int32(443)
)

type GCLWriteRequest struct {
	Entries []GCLEntry `json:"entries"`
}

type GCLEntry struct {
	LogName     string                 `json:"logName"`
	Resource    map[string]interface{} `json:"resource"`
	JsonPayload map[string]interface{} `json:"jsonPayload"`
	Severity    interface{}            `json:"severity"`
}

type serviceAccountJSON struct {
	Type                    string `json:"type"`
	ProjectID               string `json:"project_id"`
	PrivateKeyID            string `json:"private_key_id"`
	PrivateKey              string `json:"private_key"`
	ClientEmail             string `json:"client_email"`
	ClientID                string `json:"client_id"`
	AuthURI                 string `json:"auth_uri"`
	TokenURI                string `json:"token_uri"`
	AuthProviderX509CertURL string `json:"auth_provider_x509_cert_url"`
	ClientX509CertURL       string `json:"client_x509_cert_url"`
}

func GenerateFakeServiceAccountJSON(tokenURI string) ([]byte, error) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, fmt.Errorf("failed to generate RSA key: %w", err)
	}
	keyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(key),
	})

	sa := serviceAccountJSON{
		Type:                    "service_account",
		ProjectID:               "test-project",
		PrivateKeyID:            "test-key-id",
		PrivateKey:              string(keyPEM),
		ClientEmail:             "test@test-project.iam.gserviceaccount.com",
		ClientID:                "123456789",
		AuthURI:                 fmt.Sprintf("https://%s/auth", GCLDomain),
		TokenURI:                tokenURI,
		AuthProviderX509CertURL: fmt.Sprintf("https://%s/certs", GCLDomain),
		ClientX509CertURL:       fmt.Sprintf("https://%s/certs/test", GCLDomain),
	}
	return json.Marshal(sa)
}

// NewMockoonVisitor prepares the framework for GCP Cloud Logging testing and returns
// a PodBuilder visitor. Vector's gcp_stackdriver_logs sink hardcodes
// logging.googleapis.com:443 as the endpoint with no override, so we redirect
// at the network level: a host alias resolves the domain to 127.0.0.1, Mockoon
// runs as root (UID 0) to bind port 443, and a TLS certificate is generated
// for the real domain name.
func NewMockoonVisitor(framework *functional.CollectorFunctionalFramework) runtime.PodBuilderVisitor {
	return func(pb *runtime.PodBuilder) error {
		ca := certificate.NewCA(nil, "Test CA")
		serverCert := certificate.NewCert(ca, "test", GCLDomain)

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
			Hostnames: []string{GCLDomain},
		}

		mountPath := "/data"
		mockoonDataVolume := "mockoon-data"

		pb.AddConfigMapVolume(mockoonDataVolume, mockoon.ContainerName).
			AddHostAlias(hostAlias).
			AddContainer(mockoon.ContainerName, mockoon.Image).
			AddRunAsUser(0).
			AddContainerPort(mockoon.ContainerName, GCLPort).
			WithCmdArgs([]string{
				fmt.Sprintf("--data=%s/%s", mountPath, apiJsonFile),
				"--log-transaction",
			}).AddVolumeMount(mockoonDataVolume, mountPath, "", true).End()

		combinedBundle := "/tmp/ca-bundle.crt"
		pb.GetContainer(constants.CollectorName).
			AddEnvVar("SSL_CERT_FILE", combinedBundle).
			WithCmd([]string{"sh", "-c", fmt.Sprintf(
				"cat /etc/pki/tls/certs/ca-bundle.crt %s/ca.crt > %s && exec /opt/app-root/src/run.sh",
				mountPath, combinedBundle)}).
			AddVolumeMount(mockoonDataVolume, mountPath, "", true).
			Update()

		return nil
	}
}

func readMockoonLogs(ns, podName string) (string, error) {
	return oc.Logs().WithNamespace(ns).WithPod(podName).WithContainer(mockoon.ContainerName).Run()
}

func extractRawEntries(output string) ([]string, error) {
	logs, err := mockoon.DecodeLogs(output)
	if err != nil {
		return nil, err
	}

	var entries []string
	for _, log := range logs {
		if log.RequestMethod != "POST" || log.ResponseStatus != 200 ||
			!strings.Contains(log.RequestPath, "entries:write") {
			continue
		}

		var writeReq GCLWriteRequest
		if err := json.Unmarshal([]byte(log.Transaction.Request.Body), &writeReq); err != nil {
			fmt.Printf("error parsing GCL write request: %v\n", err)
			continue
		}

		for _, entry := range writeReq.Entries {
			payloadJSON, err := json.Marshal(entry.JsonPayload)
			if err != nil {
				fmt.Printf("error marshaling jsonPayload: %v\n", err)
				continue
			}
			entries = append(entries, string(payloadJSON))
		}
	}

	return entries, nil
}

// ReadRawLogEntries reads raw JSON payloads from the Mockoon transaction logs.
func ReadRawLogEntries(ns, podName string) ([]string, error) {
	output, err := readMockoonLogs(ns, podName)
	if err != nil {
		return nil, err
	}
	return extractRawEntries(output)
}

// ReadApplicationLog reads and parses application logs from the Mockoon container.
func ReadApplicationLog(ns, podName string) ([]types.ApplicationLog, error) {
	rawEntries, err := ReadRawLogEntries(ns, podName)
	if err != nil {
		return nil, err
	}

	var appLogs []types.ApplicationLog
	for _, raw := range rawEntries {
		var tmp []types.ApplicationLog
		if err := types.ParseLogsFrom(utils.ToJsonLogs([]string{raw}), &tmp, false); err != nil {
			fmt.Printf("error parsing log: %v\n", err)
			continue
		}
		appLogs = append(appLogs, tmp...)
	}
	return appLogs, nil
}
