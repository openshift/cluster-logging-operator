package util

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"strings"

	g "github.com/onsi/ginkgo/v2"
	o "github.com/onsi/gomega"
	exutil "github.com/openshift/origin/test/extended/util"
	"github.com/openshift/origin/test/extended/util/compat_otp"
	"github.com/tidwall/gjson"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	e2e "k8s.io/kubernetes/test/e2e/framework"
)

// CheckPlatform check the cluster's platform
func CheckPlatform(oc *exutil.CLI) string {
	output, _ := oc.AsAdmin().WithoutNamespace().Run("get").Args("infrastructure", "cluster", "-o=jsonpath={.status.platformStatus.type}").Output()
	return strings.ToLower(output)
}

func GetProxyFromEnv() string {
	var proxy string
	if os.Getenv("http_proxy") != "" {
		proxy = os.Getenv("http_proxy")
	} else if os.Getenv("http_proxy") != "" {
		proxy = os.Getenv("https_proxy")
	}
	return proxy
}

// Get Proxy CaBundle from the cluster, return "" if no bundle or the cluster isn't proxy enabled cluster
func getProxyCaBundle(oc *exutil.CLI) (caBundle string) {
	e2e.Logf("Check if proxxy caBundle is enabled")
	configMapName, err := oc.WithoutNamespace().AsAdmin().Run("get").Args("proxy", "cluster", "-o=jsonpath={.spec.trustedCA.name}").Output()
	if configMapName == "" {
		e2e.Logf("GetProxyCaBundle can not find proxy trustedCA, %v", err)
		return ""
	}
	caBundle, err = oc.AsAdmin().WithoutNamespace().Run("get").Args("configmap", configMapName, "-n", "openshift-config", `-ojsonpath={.data.ca-bundle\.crt}`).Output()
	o.Expect(err).NotTo(o.HaveOccurred())
	return caBundle
}

func GetClusterID(oc *exutil.CLI) (string, error) {
	return oc.AsAdmin().WithoutNamespace().Run("get").Args("clusterversion/version", "-ojsonpath={.spec.clusterID}").Output()
}

func IsFipsEnabled(oc *exutil.CLI) bool {
	nodes, err := compat_otp.GetSchedulableLinuxWorkerNodes(oc)
	o.Expect(err).NotTo(o.HaveOccurred())
	fips, err := compat_otp.DebugNodeWithChroot(oc, nodes[0].Name, "bash", "-c", "fips-mode-setup --check")
	o.Expect(err).NotTo(o.HaveOccurred())
	return strings.Contains(fips, "FIPS mode is enabled.")
}

// GetInfrastructureName returns the infrastructureName. For example:  anli922-jglp4
func GetInfrastructureName(oc *exutil.CLI) string {
	infrastructureName, err := oc.AsAdmin().WithoutNamespace().Run("get").Args("infrastructure/cluster", "-o=jsonpath={.status.infrastructureName}").Output()
	o.Expect(err).NotTo(o.HaveOccurred())
	return infrastructureName
}

func GetStorageClassName(oc *exutil.CLI) (string, error) {
	scs, err := oc.AdminKubeClient().StorageV1().StorageClasses().List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return "", err
	}
	if len(scs.Items) == 0 {
		return "", fmt.Errorf("there is no storageclass in the cluster")
	}
	for _, sc := range scs.Items {
		if sc.ObjectMeta.Annotations["storageclass.kubernetes.io/is-default-class"] == "true" {
			return sc.Name, nil
		}
	}
	return scs.Items[0].Name, nil
}

// Get OIDC provider for the cluster
func GetOIDC(oc *exutil.CLI) (string, error) {
	oidc, err := oc.AsAdmin().WithoutNamespace().Run("get").Args("authentication.config", "cluster", "-o=jsonpath={.spec.serviceAccountIssuer}").Output()
	if err != nil {
		return "", err
	}
	return strings.TrimPrefix(oidc, "https://"), nil
}

func GetPoolID(oc *exutil.CLI) (string, error) {
	// pool_id="$(oc get authentication cluster -o json | jq -r .spec.serviceAccountIssuer | sed 's/.*\/\([^\/]*\)-oidc/\1/')"
	issuer, err := GetOIDC(oc)
	if err != nil {
		return "", err
	}

	return strings.Split(strings.Split(issuer, "/")[1], "-oidc")[0], nil
}

func HasMaster(oc *exutil.CLI) bool {
	masterNodes, err := oc.AdminKubeClient().CoreV1().Nodes().List(context.Background(), metav1.ListOptions{LabelSelector: "node-role.kubernetes.io/master="})
	if err != nil {
		e2e.Logf("hit error when listing master nodes: %v", err)
	}
	return len(masterNodes.Items) > 0
}

func GetNetworkType(oc *exutil.CLI) string {
	output, err := oc.WithoutNamespace().AsAdmin().Run("get").Args("network.operator", "cluster", "-o=jsonpath={.spec.defaultNetwork.type}").Output()
	o.Expect(err).NotTo(o.HaveOccurred())
	return strings.ToLower(output)
}

func GetAppDomain(oc *exutil.CLI) (string, error) {
	subDomain, err := oc.AsAdmin().WithoutNamespace().Run("get").Args("ingresses.config/cluster", "-ojsonpath={.spec.domain}").Output()
	if err != nil {
		return "", err
	}
	return subDomain, nil
}

// GetIPVersionStackType gets IP-version Stack type of the cluster
func GetIPVersionStackType(oc *exutil.CLI) (ipvStackType string) {
	svcNetwork, err := oc.WithoutNamespace().AsAdmin().Run("get").Args("network.operator", "cluster", "-o=jsonpath={.spec.serviceNetwork}").Output()
	o.Expect(err).NotTo(o.HaveOccurred())
	if strings.Count(svcNetwork, ":") >= 2 && strings.Count(svcNetwork, ".") >= 2 {
		ipvStackType = "dualstack"
	} else if strings.Count(svcNetwork, ":") >= 2 {
		ipvStackType = "ipv6single"
	} else if strings.Count(svcNetwork, ".") >= 2 {
		ipvStackType = "ipv4single"
	}
	e2e.Logf("The test cluster IP-version Stack type is :\"%s\".", ipvStackType)
	return ipvStackType
}

// GetAwsCredentialFromCluster get aws credential from cluster
func GetAwsCredentialFromCluster(oc *exutil.CLI) {
	credential, err := oc.AsAdmin().WithoutNamespace().Run("get").Args("secret/aws-creds", "-n", "kube-system", "-o", "json").Output()
	// Skip for sts and c2s clusters.
	if err != nil {
		g.Skip("Did not get credential to access aws, skip the testing.")

	}
	o.Expect(err).NotTo(o.HaveOccurred())
	accessKeyIDBase64, secureKeyBase64 := gjson.Get(credential, `data.aws_access_key_id`).String(), gjson.Get(credential, `data.aws_secret_access_key`).String()
	accessKeyID, err1 := base64.StdEncoding.DecodeString(accessKeyIDBase64)
	o.Expect(err1).NotTo(o.HaveOccurred())
	secureKey, err2 := base64.StdEncoding.DecodeString(secureKeyBase64)
	o.Expect(err2).NotTo(o.HaveOccurred())
	clusterRegion, err3 := oc.AsAdmin().WithoutNamespace().Run("get").Args("infrastructure", "cluster", "-o=jsonpath={.status.platformStatus.aws.region}").Output()
	o.Expect(err3).NotTo(o.HaveOccurred())
	os.Setenv("AWS_ACCESS_KEY_ID", string(accessKeyID))
	os.Setenv("AWS_SECRET_ACCESS_KEY", string(secureKey))
	os.Setenv("AWS_REGION", clusterRegion)
}

// checkout the cloudType of this cluster's platform
func GetAzureCloudName(oc *exutil.CLI) string {
	cloudName, err := oc.AsAdmin().WithoutNamespace().Run("get").Args("infrastructure", "cluster", "-o=jsonpath={.status.platformStatus.azure.cloudName}").Output()
	if err == nil && len(cloudName) > 0 {
		return strings.ToLower(cloudName)
	}
	return ""
}
