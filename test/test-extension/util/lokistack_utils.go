package util

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	g "github.com/onsi/ginkgo/v2"
	o "github.com/onsi/gomega"
	exutil "github.com/openshift/origin/test/extended/util"
	"github.com/openshift/origin/test/extended/util/compat_otp"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	k8sresource "k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	e2e "k8s.io/kubernetes/test/e2e/framework"
)

// S3Credential defines the s3 credentials
type S3Credential struct {
	Region          string
	AccessKeyID     string
	SecretAccessKey string
	Endpoint        string //the endpoint of s3 service
}

func GetAWSCredentialFromCluster(oc *exutil.CLI) S3Credential {
	region, err := compat_otp.GetAWSClusterRegion(oc)
	o.Expect(err).NotTo(o.HaveOccurred())

	dirname := "/tmp/" + oc.Namespace() + "-creds"
	defer os.RemoveAll(dirname)
	err = os.MkdirAll(dirname, 0777)
	o.Expect(err).NotTo(o.HaveOccurred())

	_, err = oc.AsAdmin().WithoutNamespace().Run("extract").Args("secret/aws-creds", "-n", "kube-system", "--confirm", "--to="+dirname).Output()
	o.Expect(err).NotTo(o.HaveOccurred())

	accessKeyID, err := os.ReadFile(dirname + "/aws_access_key_id")
	o.Expect(err).NotTo(o.HaveOccurred())
	secretAccessKey, err := os.ReadFile(dirname + "/aws_secret_access_key")
	o.Expect(err).NotTo(o.HaveOccurred())

	cred := S3Credential{Region: region, AccessKeyID: string(accessKeyID), SecretAccessKey: string(secretAccessKey)}
	return cred
}

func getMinIOCreds(oc *exutil.CLI, ns string) AWSClientConfig {
	dirname := "/tmp/" + oc.Namespace() + "-creds"
	defer os.RemoveAll(dirname)
	err := os.MkdirAll(dirname, 0777)
	o.Expect(err).NotTo(o.HaveOccurred())

	_, err = oc.AsAdmin().WithoutNamespace().Run("extract").Args("secret/"+minioSecret, "-n", ns, "--confirm", "--to="+dirname).Output()
	o.Expect(err).NotTo(o.HaveOccurred())

	accessKeyID, err := os.ReadFile(dirname + "/access_key_id")
	o.Expect(err).NotTo(o.HaveOccurred())
	secretAccessKey, err := os.ReadFile(dirname + "/secret_access_key")
	o.Expect(err).NotTo(o.HaveOccurred())

	endpoint := "http://" + GetRouteAddress(oc, ns, "minio")
	return AWSClientConfig{Endpoint: endpoint, AccessKey: string(accessKeyID), SecretKey: string(secretAccessKey), InsecureSkipTLS: true, Region: "auto"}
}

// createSecretForAWSS3Bucket creates a secret for Loki to connect to s3 bucket
func createSecretForAWSS3Bucket(oc *exutil.CLI, bucketName, secretName, ns string, cred AWSClientConfig) error {
	if len(secretName) == 0 {
		return fmt.Errorf("secret name shouldn't be empty")
	}

	err := LoadAccessKeySecretKeyFromFile(context.TODO(), &cred)
	if err != nil {
		return fmt.Errorf("error loading aws credentials")
	}

	endpoint := "https://s3." + cred.Region + ".amazonaws.com"
	return oc.NotShowInfo().AsAdmin().WithoutNamespace().Run("create").Args("secret", "generic", secretName, "--from-literal=access_key_id="+cred.AccessKey, "--from-literal=access_key_secret="+cred.SecretKey, "--from-literal=region="+cred.Region, "--from-literal=bucketnames="+bucketName, "--from-literal=endpoint="+endpoint, "-n", ns).Execute()
}

func CreateSecretForODFBucket(oc *exutil.CLI, bucketName, secretName, ns string) error {
	if len(secretName) == 0 {
		return fmt.Errorf("secret name shouldn't be empty")
	}
	dirname := "/tmp/" + oc.Namespace() + "-creds"
	err := os.MkdirAll(dirname, 0777)
	o.Expect(err).NotTo(o.HaveOccurred())
	defer os.RemoveAll(dirname)
	_, err = oc.AsAdmin().WithoutNamespace().Run("extract").Args("secret/noobaa-admin", "-n", "openshift-storage", "--confirm", "--to="+dirname).Output()
	o.Expect(err).NotTo(o.HaveOccurred())

	endpoint := "http://s3.openshift-storage.svc:80"
	return oc.AsAdmin().WithoutNamespace().Run("create").Args("secret", "generic", secretName, "--from-file=access_key_id="+dirname+"/AWS_ACCESS_KEY_ID", "--from-file=access_key_secret="+dirname+"/AWS_SECRET_ACCESS_KEY", "--from-literal=bucketnames="+bucketName, "--from-literal=endpoint="+endpoint, "-n", ns).Execute()
}

func CreateSecretForMinIOBucket(oc *exutil.CLI, bucketName, secretName, ns string, cred AWSClientConfig) error {
	if len(secretName) == 0 {
		return fmt.Errorf("secret name shouldn't be empty")
	}
	return oc.NotShowInfo().AsAdmin().WithoutNamespace().Run("create").Args("secret", "generic", secretName, "--from-literal=access_key_id="+cred.AccessKey, "--from-literal=access_key_secret="+cred.SecretKey, "--from-literal=bucketnames="+bucketName, "--from-literal=endpoint="+cred.Endpoint, "-n", ns).Execute()
}

func CreateSecretForGCSBucketWithSTS(oc *exutil.CLI, namespace, secretName, bucketName string) error {
	return oc.NotShowInfo().AsAdmin().WithoutNamespace().Run("create").Args("secret", "generic", "-n", namespace, secretName, "--from-literal=bucketname="+bucketName).Execute()
}

// CreateSecretForGCSBucket creates a secret for Loki to connect to gcs bucket
func CreateSecretForGCSBucket(oc *exutil.CLI, bucketName, secretName, ns string) error {
	if len(secretName) == 0 {
		return fmt.Errorf("secret name shouldn't be empty")
	}

	//get gcp-credentials from env var GOOGLE_APPLICATION_CREDENTIALS
	gcsCred := os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")
	return oc.AsAdmin().WithoutNamespace().Run("create").Args("secret", "generic", secretName, "-n", ns, "--from-literal=bucketname="+bucketName, "--from-file=key.json="+gcsCred).Execute()
}

// Creates a secret for Loki to connect to azure container
func CreateSecretForAzureContainer(oc *exutil.CLI, bucketName, secretName, ns string) error {
	environment := "AzureGlobal"
	cloudName, err := oc.AsAdmin().WithoutNamespace().Run("get").Args("infrastructure", "cluster", "-o=jsonpath={.status.platformStatus.azure.cloudName}").Output()
	if err != nil {
		return fmt.Errorf("can't get azure cluster type  %v", err)
	}
	if strings.ToLower(cloudName) == "azureusgovernmentcloud" {
		environment = "AzureUSGovernment"
	}
	if strings.ToLower(cloudName) == "azurechinacloud" {
		environment = "AzureChinaCloud"
	}
	if strings.ToLower(cloudName) == "azuregermancloud" {
		environment = "AzureGermanCloud"
	}

	accountName, accountKey, err1 := compat_otp.GetAzureStorageAccountFromCluster(oc)
	if err1 != nil {
		return fmt.Errorf("can't get azure storage account from cluster: %v", err1)
	}
	return oc.NotShowInfo().AsAdmin().WithoutNamespace().Run("create").Args("secret", "generic", "-n", ns, secretName, "--from-literal=environment="+environment, "--from-literal=container="+bucketName, "--from-literal=account_name="+accountName, "--from-literal=account_key="+accountKey).Execute()
}

func CreateSecretForSwiftContainer(oc *exutil.CLI, containerName, secretName, ns string, cred *compat_otp.OpenstackCredentials) error {
	userID, domainID := compat_otp.GetOpenStackUserIDAndDomainID(cred)
	err := oc.NotShowInfo().AsAdmin().WithoutNamespace().Run("create").Args("secret", "generic", "-n", ns, secretName,
		"--from-literal=auth_url="+cred.Clouds.Openstack.Auth.AuthURL,
		"--from-literal=username="+cred.Clouds.Openstack.Auth.Username,
		"--from-literal=user_domain_name="+cred.Clouds.Openstack.Auth.UserDomainName,
		"--from-literal=user_domain_id="+domainID,
		"--from-literal=user_id="+userID,
		"--from-literal=password="+cred.Clouds.Openstack.Auth.Password,
		"--from-literal=domain_id="+domainID,
		"--from-literal=domain_name="+cred.Clouds.Openstack.Auth.UserDomainName,
		"--from-literal=container_name="+containerName,
		"--from-literal=project_id="+cred.Clouds.Openstack.Auth.ProjectID,
		"--from-literal=project_name="+cred.Clouds.Openstack.Auth.ProjectName,
		"--from-literal=project_domain_id="+domainID,
		"--from-literal=project_domain_name="+cred.Clouds.Openstack.Auth.UserDomainName).Execute()
	return err
}

// CheckODF check if the ODF is installed in the cluster or not
// here only checks the sc/ocs-storagecluster-ceph-rbd and svc/s3
func CheckODF(oc *exutil.CLI) bool {
	svcFound := false
	expectedSC := []string{"openshift-storage.noobaa.io"}
	var scInCluster []string
	scs, err := oc.AdminKubeClient().StorageV1().StorageClasses().List(context.Background(), metav1.ListOptions{})
	o.Expect(err).NotTo(o.HaveOccurred())

	for _, sc := range scs.Items {
		scInCluster = append(scInCluster, sc.Name)
	}

	for _, s := range expectedSC {
		if !Contain(scInCluster, s) {
			return false
		}
	}

	_, err = oc.AdminKubeClient().CoreV1().Services("openshift-storage").Get(context.Background(), "s3", metav1.GetOptions{})
	if err == nil {
		svcFound = true
	}
	return svcFound
}

func CreateObjectBucketClaim(oc *exutil.CLI, ns, name string) error {
	template := compat_otp.FixturePath("testdata", "logging", "odf", "objectBucketClaim.yaml")
	err := ApplyResourceFromTemplate(oc, ns, "-p", "NAME="+name, "NAMESPACE="+ns, "-f", template)
	if err != nil {
		return err
	}
	err = WaitForResourceToAppear(oc, "objectbucketclaims", name, ns)
	if err != nil {
		return err
	}
	err = WaitForResourceToAppear(oc, "objectbuckets", "obc-"+ns+"-"+name, ns)
	if err != nil {
		return err
	}
	AssertResourceStatus(oc, "objectbucketclaims", name, ns, "{.status.phase}", "Bound")
	return nil
}

func DeleteObjectBucketClaim(oc *exutil.CLI, ns, name string) error {
	err := DeleteResourceFromCluster(oc, "objectbucketclaims", name, ns)
	if err != nil {
		return err
	}
	return WaitUntilResourceIsGone(oc, "objectbucketclaims", name, ns)
}

// checkMinIO
func CheckMinIO(oc *exutil.CLI, ns string) (bool, error) {
	podReady, svcFound := false, false
	pod, err := oc.AdminKubeClient().CoreV1().Pods(ns).List(context.Background(), metav1.ListOptions{LabelSelector: "app=minio"})
	if err != nil {
		return false, err
	}
	if len(pod.Items) > 0 && pod.Items[0].Status.Phase == "Running" {
		podReady = true
	}
	_, err = oc.AdminKubeClient().CoreV1().Services(ns).Get(context.Background(), "minio", metav1.GetOptions{})
	if err == nil {
		svcFound = true
	}
	return podReady && svcFound, err
}

func UseExtraObjectStorage(oc *exutil.CLI) string {
	if CheckODF(oc) {
		e2e.Logf("use the existing ODF storage service")
		return "odf"
	}
	ready, err := CheckMinIO(oc, minioNS)
	if ready {
		e2e.Logf("use existing MinIO storage service")
		return "minio"
	}
	if strings.Contains(err.Error(), "No resources found") || strings.Contains(err.Error(), "not found") {
		e2e.Logf("deploy MinIO and use this MinIO as storage service")
		DeployMinIO(oc)
		return "minio"
	}
	return ""
}

func PatchLokiOperatorWithAWSRoleArn(oc *exutil.CLI, subNamespace, roleArn string) {
	roleArnPatchConfig := `{
		"spec": {
		  "config": {
			"env": [
			  {
				"name": "ROLEARN",
				"value": "%s"
			  }
			]
		  }
		}
	  }`

	subName, err := oc.AsAdmin().WithoutNamespace().Run("get").Args("sub", "-n", subNamespace, `-ojsonpath={.items[?(@.spec.name=="loki-operator")].metadata.name}`).Output()
	o.Expect(err).NotTo(o.HaveOccurred())
	o.Expect(subName).ShouldNot(o.BeEmpty())
	err = oc.NotShowInfo().AsAdmin().WithoutNamespace().Run("patch").Args("sub", subName, "-n", subNamespace, "-p", fmt.Sprintf(roleArnPatchConfig, roleArn), "--type=merge").Execute()
	o.Expect(err).NotTo(o.HaveOccurred())
	WaitForPodReadyByLabel(oc, LoNS, "name=loki-operator-controller-manager")
}

// return the storage type per different platform
func GetStorageType(oc *exutil.CLI) string {
	platform := compat_otp.CheckPlatform(oc)
	switch platform {
	case "aws":
		{
			return "s3"
		}
	case "gcp":
		{
			return "gcs"
		}
	case "azure":
		{
			return "azure"
		}
	case "openstack":
		{
			return "swift"
		}
	default:
		{
			return UseExtraObjectStorage(oc)
		}
	}
}

// LokiStack contains the configurations of loki stack
type LokiStack struct {
	Name          string // lokiStack name
	Namespace     string // lokiStack namespace
	TSize         string // size
	StorageType   string // the backend storage type, currently support s3, gcs, azure, swift, ODF and minIO
	StorageSecret string // the secret name for loki to use to connect to backend storage
	StorageClass  string // storage class name
	BucketName    string // the butcket or the container name where loki stores it's data in
	Template      string // the file used to create the loki stack
}

func (l LokiStack) setTSize(size string) LokiStack {
	l.TSize = size
	return l
}

// PrepareResourcesForLokiStack creates buckets/containers in backend storage provider, and creates the secret for Loki to use
func (l LokiStack) PrepareResourcesForLokiStack(oc *exutil.CLI) error {
	var err error
	if len(l.BucketName) == 0 {
		return fmt.Errorf("the bucketName should not be empty")
	}
	switch l.StorageType {
	case "s3":
		{
			cred, err1 := GetAWSCredentials(oc)
			if err1 != nil {
				g.Skip("Skip since no AWS credetial! No Env AWS_SHARED_CREDENTIALS_FILE, Env CLUSTER_PROFILE_DIR  or $HOME/.aws/credentials file")
			}
			factory, err2 := NewAWSClientFactory(context.TODO(), cred)
			if err2 != nil {
				e2e.Failf("error loading aws config: %v", err2)
			}
			if compat_otp.IsWorkloadIdentityCluster(oc) {
				partition := "aws"
				if strings.HasPrefix(cred.Region, "us-gov") {
					partition = "aws-us-gov"
				}
				iamClient := factory.IAM()
				stsClient := factory.STS()
				awsAccountID, _ := GetAwsAccount(stsClient)
				oidcName, err := GetOIDC(oc)
				o.Expect(err).NotTo(o.HaveOccurred())
				lokiIAMRoleName := l.Name + "-" + compat_otp.GetRandomString()
				roleArn := CreateIAMRoleForS3Bucket(iamClient, oidcName, awsAccountID, partition, l.Namespace, l.Name, lokiIAMRoleName)
				os.Setenv("LOKI_ROLE_NAME_ON_STS", lokiIAMRoleName)
				PatchLokiOperatorWithAWSRoleArn(oc, LoNS, roleArn)
				CreateObjectStorageSecretOnAWSSTSCluster(oc, cred.Region, l.StorageSecret, l.BucketName, l.Namespace)
			} else {
				err = createSecretForAWSS3Bucket(oc, l.BucketName, l.StorageSecret, l.Namespace, cred)
				if err != nil {
					return fmt.Errorf("failed to create secret for aws s3 buctke: %v", err)
				}
			}
			s3Client := factory.S3()
			return CreateS3Bucket(s3Client, l.BucketName, cred.Region)
		}
	case "azure":
		{
			if compat_otp.IsWorkloadIdentityCluster(oc) {
				if !ReadAzureCredentials() {
					g.Skip("Azure Credentials not found. Skip case!")
				} else {
					PerformManagedIdentityAndSecretSetupForAzureWIF(oc, l.Name, l.Namespace, l.BucketName, l.StorageSecret)
				}
			} else {
				accountName, accountKey, err1 := compat_otp.GetAzureStorageAccountFromCluster(oc)
				if err1 != nil {
					return fmt.Errorf("can't get azure storage account from cluster: %v", err1)
				}
				client, err2 := compat_otp.NewAzureContainerClient(oc, accountName, accountKey, l.BucketName)
				if err2 != nil {
					return err2
				}
				err = compat_otp.CreateAzureStorageBlobContainer(client)
				if err != nil {
					return err
				}
				err = CreateSecretForAzureContainer(oc, l.BucketName, l.StorageSecret, l.Namespace)
			}
		}
	case "gcs":
		{
			projectID, errGetID := GetGcpProjectID(oc)
			o.Expect(errGetID).NotTo(o.HaveOccurred())
			err = CreateGCSBucket(projectID, l.BucketName)
			if err != nil {
				return err
			}
			if compat_otp.IsWorkloadIdentityCluster(oc) {
				clusterName := GetInfrastructureName(oc)
				gcsSAName := GenerateServiceAccountNameForGCS(clusterName)
				os.Setenv("LOGGING_GCS_SERVICE_ACCOUNT_NAME", gcsSAName)
				projectNumber, err1 := GetGCPProjectNumber(projectID)
				if err1 != nil {
					return fmt.Errorf("can't get GCP project number: %v", err1)
				}
				poolID, err2 := GetPoolID(oc)
				if err2 != nil {
					return fmt.Errorf("can't get pool ID: %v", err2)
				}
				sa, err3 := CreateServiceAccountOnGCP(projectID, gcsSAName)
				if err3 != nil {
					return fmt.Errorf("can't create service account: %v", err3)
				}
				os.Setenv("LOGGING_GCS_SERVICE_ACCOUNT_EMAIL", sa.Email)
				err4 := GrantPermissionsToGCPServiceAccount(poolID, projectID, projectNumber, l.Namespace, l.Name, sa.Email)
				if err4 != nil {
					return fmt.Errorf("can't add roles to the serviceaccount: %v", err4)
				}

				PatchLokiOperatorOnGCPSTSforCCO(oc, LoNS, projectNumber, poolID, sa.Email)

				err = CreateSecretForGCSBucketWithSTS(oc, l.Namespace, l.StorageSecret, l.BucketName)
			} else {
				err = CreateSecretForGCSBucket(oc, l.BucketName, l.StorageSecret, l.Namespace)
			}
		}
	case "swift":
		{
			cred, err1 := compat_otp.GetOpenStackCredentials(oc)
			o.Expect(err1).NotTo(o.HaveOccurred())
			client := compat_otp.NewOpenStackClient(cred, "object-store")
			err = compat_otp.CreateOpenStackContainer(client, l.BucketName)
			if err != nil {
				return err
			}
			err = CreateSecretForSwiftContainer(oc, l.BucketName, l.StorageSecret, l.Namespace, cred)
		}
	case "odf":
		{
			err = CreateObjectBucketClaim(oc, l.Namespace, l.BucketName)
			if err != nil {
				return err
			}
			err = CreateSecretForODFBucket(oc, l.BucketName, l.StorageSecret, l.Namespace)
		}
	case "minio":
		{
			cred := getMinIOCreds(oc, minioNS)
			factory, err := NewAWSClientFactory(context.TODO(), cred)
			if err != nil {
				e2e.Failf("error loading aws config: %v", err)
			}
			client := factory.S3()
			if err != nil {
				return err
			}
			err = CreateS3Bucket(client, l.BucketName, "")
			if err != nil {
				return err
			}
			err = CreateSecretForMinIOBucket(oc, l.BucketName, l.StorageSecret, l.Namespace, cred)
		}
	}
	return err
}

// DeployLokiStack creates the lokiStack CR with basic settings: name, namespace, size, storage.secret.name, storage.secret.type, storageClassName
// optionalParameters is designed for adding parameters to deploy lokiStack with different tenants or some other settings
func (l LokiStack) DeployLokiStack(oc *exutil.CLI, optionalParameters ...string) error {
	e2e.Logf("Running deployLokiStack")

	var storage string
	if l.StorageType == "odf" || l.StorageType == "minio" {
		storage = "s3"
	} else {
		storage = l.StorageType
	}

	lokistackTemplate := l.Template
	if GetIPVersionStackType(oc) == "ipv6single" {
		lokistackTemplate = strings.Replace(l.Template, ".yaml", "-ipv6.yaml", -1)
	}
	//Add Proxy CA for object storage
	caBundle := getProxyCaBundle(oc)
	if caBundle != "" {
		e2e.Logf("Enable caBundle for lokistack storage")
		lokistackTemplate = strings.Replace(lokistackTemplate, ".yaml", "-tls.yaml", -1)
		err := oc.AsAdmin().WithoutNamespace().Run("create").Args("-n", l.Namespace, "configmap", l.Name+"-proxy-ca", "--from-literal=ca-bundle.crt="+caBundle).Execute()
		o.Expect(err).NotTo(o.HaveOccurred())
		optionalParameters = append(optionalParameters, "CA_NAME="+l.Name+"-proxy-ca")
		optionalParameters = append(optionalParameters, "CA_KEY_NAME=ca-bundle.crt")
	}
	parameters := []string{"-f", lokistackTemplate, "-n", l.Namespace, "-p", "NAME=" + l.Name, "NAMESPACE=" + l.Namespace, "SIZE=" + l.TSize, "SECRET_NAME=" + l.StorageSecret, "STORAGE_TYPE=" + storage, "STORAGE_CLASS=" + l.StorageClass}

	if len(optionalParameters) != 0 {
		parameters = append(parameters, optionalParameters...)
	}

	file, err := ProcessTemplate(oc, parameters...)
	defer os.Remove(file)
	compat_otp.AssertWaitPollNoErr(err, fmt.Sprintf("Can not process %v", parameters))
	err = oc.AsAdmin().WithoutNamespace().Run("apply").Args("-f", file, "-n", l.Namespace).Execute()
	WaitForResourceToAppear(oc, "lokistack", l.Name, l.Namespace)
	return err
}

func (l LokiStack) WaitForLokiStackToBeReady(oc *exutil.CLI) {
	for _, deploy := range []string{l.Name + "-gateway", l.Name + "-distributor", l.Name + "-querier", l.Name + "-query-frontend"} {
		WaitForDeploymentPodsToBeReady(oc, l.Namespace, deploy)
	}
	for _, ss := range []string{l.Name + "-index-gateway", l.Name + "-compactor", l.Name + "-ruler", l.Name + "-ingester"} {
		WaitForStatefulsetReady(oc, l.Namespace, ss)
	}
	if compat_otp.IsWorkloadIdentityCluster(oc) {
		currentPlatform := CheckPlatform(oc)
		switch currentPlatform {
		case "aws", "azure", "gcp":
			ValidateCredentialsRequestGenerationOnSTS(oc, l.Name, l.Namespace)
		}
	}
}

/*
// update existing lokistack CR
// if template is specified, then run command `oc process -f template -p patches | oc apply -f -`
// if template is not specified, then run command `oc patch lokistack/${l.name} -p patches`
// if use patch, should add `--type=` in the end of patches
func (l lokiStack) update(oc *exutil.CLI, template string, patches ...string) {
	var err error
	if template != "" {
		parameters := []string{"-f", template, "-p", "NAME=" + l.name, "NAMESPACE=" + l.namespace}
		if len(patches) > 0 {
			parameters = append(parameters, patches...)
		}
		file, processErr := processTemplate(oc, parameters...)
		defer os.Remove(file)
		if processErr != nil {
			e2e.Failf("error processing file: %v", processErr)
		}
		err = oc.AsAdmin().WithoutNamespace().Run("apply").Args("-f", file, "-n", l.namespace).Execute()
	} else {
		parameters := []string{"lokistack/" + l.name, "-n", l.namespace, "-p"}
		parameters = append(parameters, patches...)
		err = oc.AsAdmin().WithoutNamespace().Run("patch").Args(parameters...).Execute()
	}
	if err != nil {
		e2e.Failf("error updating lokistack: %v", err)
	}
}
*/

func (l LokiStack) RemoveLokiStack(oc *exutil.CLI) {
	_ = DeleteResourceFromCluster(oc, "lokistack", l.Name, l.Namespace)
	_ = oc.AsAdmin().WithoutNamespace().Run("delete").Args("pvc", "-n", l.Namespace, "-l", "app.kubernetes.io/instance="+l.Name).Execute()
	_ = oc.AsAdmin().WithoutNamespace().Run("delete").Args("configmap", "-n", l.Namespace, l.Name+"-proxy-ca").Execute()
}

func (l LokiStack) RemoveObjectStorage(oc *exutil.CLI) {
	e2e.Logf("Remove Object Storage")
	_ = DeleteResourceFromCluster(oc, "secret", l.StorageSecret, l.Namespace)
	var err error
	switch l.StorageType {
	case "s3":
		{
			cred, _ := GetAWSCredentials(oc)
			factory, err1 := NewAWSClientFactory(context.TODO(), cred)
			if err1 != nil {
				e2e.Failf("error loading aws config: %v", err1)
			}
			if compat_otp.IsWorkloadIdentityCluster(oc) {
				iamClient := factory.IAM()
				DeleteIAMroleonAWS(iamClient, os.Getenv("LOKI_ROLE_NAME_ON_STS"))
				os.Unsetenv("LOKI_ROLE_NAME_ON_STS")
				subName, err := oc.AsAdmin().WithoutNamespace().Run("get").Args("sub", "-n", LoNS, `-ojsonpath={.items[?(@.spec.name=="loki-operator")].metadata.name}`).Output()
				o.Expect(err).NotTo(o.HaveOccurred())
				o.Expect(subName).ShouldNot(o.BeEmpty())
				err = oc.AsAdmin().WithoutNamespace().Run("patch").Args("sub", subName, "-n", LoNS, "-p", `[{"op": "remove", "path": "/spec/config"}]`, "--type=json").Execute()
				o.Expect(err).NotTo(o.HaveOccurred())
				WaitForPodReadyByLabel(oc, LoNS, "name=loki-operator-controller-manager")
			}
			client := factory.S3()
			err = DeleteS3Bucket(client, l.BucketName)
		}
	case "azure":
		{
			if compat_otp.IsWorkloadIdentityCluster(oc) {
				resourceGroup, err := GetAzureResourceGroupFromCluster(oc)
				o.Expect(err).NotTo(o.HaveOccurred())
				azureSubscriptionID := os.Getenv("AZURE_SUBSCRIPTION_ID")
				cred := CreateNewDefaultAzureCredential()
				DeleteManagedIdentityOnAzure(cred, azureSubscriptionID, resourceGroup, l.Name)
				DeleteAzureStorageAccount(cred, azureSubscriptionID, resourceGroup, os.Getenv("LOKI_OBJECT_STORAGE_STORAGE_ACCOUNT"))
				os.Unsetenv("LOKI_OBJECT_STORAGE_STORAGE_ACCOUNT")
			} else {
				accountName, accountKey, err1 := compat_otp.GetAzureStorageAccountFromCluster(oc)
				o.Expect(err1).NotTo(o.HaveOccurred())
				client, err2 := compat_otp.NewAzureContainerClient(oc, accountName, accountKey, l.BucketName)
				o.Expect(err2).NotTo(o.HaveOccurred())
				err = compat_otp.DeleteAzureStorageBlobContainer(client)
			}
		}
	case "gcs":
		{
			if compat_otp.IsWorkloadIdentityCluster(oc) {
				sa := os.Getenv("LOGGING_GCS_SERVICE_ACCOUNT_NAME")
				if sa == "" {
					e2e.Logf("LOGGING_GCS_SERVICE_ACCOUNT_NAME is not set, no need to delete the serviceaccount")
				} else {
					os.Unsetenv("LOGGING_GCS_SERVICE_ACCOUNT_NAME")
					email := os.Getenv("LOGGING_GCS_SERVICE_ACCOUNT_EMAIL")
					if email == "" {
						e2e.Logf("LOGGING_GCS_SERVICE_ACCOUNT_EMAIL is not set, no need to delete the policies")
					} else {
						os.Unsetenv("LOGGING_GCS_SERVICE_ACCOUNT_EMAIL")
						projectID, errGetID := GetGcpProjectID(oc)
						o.Expect(errGetID).NotTo(o.HaveOccurred())
						projectNumber, _ := GetGCPProjectNumber(projectID)
						poolID, _ := GetPoolID(oc)
						err = RemovePermissionsFromGCPServiceAccount(poolID, projectID, projectNumber, l.Namespace, l.Name, email)
						o.Expect(err).NotTo(o.HaveOccurred())
						err = RemoveServiceAccountFromGCP("projects/" + projectID + "/serviceAccounts/" + email)
						o.Expect(err).NotTo(o.HaveOccurred())
					}
				}
			}
			err = DeleteGCSBucket(l.BucketName)
		}
	case "swift":
		{
			cred, err1 := compat_otp.GetOpenStackCredentials(oc)
			o.Expect(err1).NotTo(o.HaveOccurred())
			client := compat_otp.NewOpenStackClient(cred, "object-store")
			err = compat_otp.DeleteOpenStackContainer(client, l.BucketName)
		}
	case "odf":
		{
			err = DeleteObjectBucketClaim(oc, l.Namespace, l.BucketName)
		}
	case "minio":
		{
			cred := getMinIOCreds(oc, minioNS)
			factory, err1 := NewAWSClientFactory(context.TODO(), cred)
			if err1 != nil {
				e2e.Failf("error loading aws config: %v", err1)
			}
			client := factory.S3()
			err = DeleteS3Bucket(client, l.BucketName)
		}
	}
	o.Expect(err).NotTo(o.HaveOccurred())
}

func (l LokiStack) CreateSecretFromGateway(oc *exutil.CLI, name, namespace, token string) {
	dirname := "/tmp/" + oc.Namespace() + GetRandomString()
	defer os.RemoveAll(dirname)
	err := os.MkdirAll(dirname, 0777)
	o.Expect(err).NotTo(o.HaveOccurred())

	err = oc.AsAdmin().WithoutNamespace().Run("extract").Args("cm/"+l.Name+"-gateway-ca-bundle", "-n", l.Namespace, "--keys=service-ca.crt", "--confirm", "--to="+dirname).Execute()
	o.Expect(err).NotTo(o.HaveOccurred())

	if token != "" {
		err = oc.NotShowInfo().AsAdmin().WithoutNamespace().Run("create").Args("secret", "generic", name, "-n", namespace, "--from-file=ca-bundle.crt="+dirname+"/service-ca.crt", "--from-literal=token="+token).Execute()
	} else {
		err = oc.AsAdmin().WithoutNamespace().Run("create").Args("secret", "generic", name, "-n", namespace, "--from-file=ca-bundle.crt="+dirname+"/service-ca.crt").Execute()
	}
	o.Expect(err).NotTo(o.HaveOccurred())

}

// Global function to check if logs are pushed to external storage.
// Currently supports Amazon S3, Azure Blob Storage and Google Cloud Storage bucket.
func (l LokiStack) ValidateExternalObjectStorageForLogs(oc *exutil.CLI, tenants []string) {
	switch l.StorageType {
	case "s3":
		{
			cred, err := GetAWSCredentials(oc)
			o.Expect(err).NotTo(o.HaveOccurred())
			factory, err := NewAWSClientFactory(context.TODO(), cred)
			if err != nil {
				e2e.Failf("error loading aws config: %v", err)
			}
			s3Client := factory.S3()
			o.Expect(err).NotTo(o.HaveOccurred())
			ValidatesIfLogsArePushedToS3Bucket(s3Client, l.BucketName, tenants)
		}
	case "azure":
		{
			// For Azure Container Storage
			var accountName string
			var err error
			_, storageAccountURISuffix := GetStorageAccountURISuffixAndEnvForAzure(oc)
			if compat_otp.IsSTSCluster(oc) {
				accountName = os.Getenv("LOKI_OBJECT_STORAGE_STORAGE_ACCOUNT")
			} else {
				_, err = compat_otp.GetAzureCredentialFromCluster(oc)
				o.Expect(err).NotTo(o.HaveOccurred())
				accountName, _, err = compat_otp.GetAzureStorageAccountFromCluster(oc)
				o.Expect(err).NotTo(o.HaveOccurred())
			}
			ValidatesIfLogsArePushedToAzureContainer(storageAccountURISuffix, accountName, l.BucketName, tenants)
		}
	case "gcs":
		{
			// For Google Cloud Storage Bucket
			ValidatesIfLogsArePushedToGCSBucket(l.BucketName, tenants)
		}

	case "swift":
		{
			e2e.Logf("Currently swift is not supported")
			// TODO swift code here
		}
	case "minio":
		{
			ValidatesIfLogsArePushedToMinIOBucket(oc, l.BucketName, tenants)
		}
	default:
		{
			e2e.Logf("Currently %s is not supported", l.StorageType)
		}
	}
}

// TODO: add an option to provide TLS config
type LokiClient struct {
	Username        string //Username for HTTP basic auth.
	Password        string //Password for HTTP basic auth
	Address         string //Server address.
	OrgID           string //adds X-Scope-OrgID to API requests for representing tenant ID. Useful for requesting tenant data when bypassing an auth gateway.
	BearerToken     string //adds the Authorization header to API requests for authentication purposes.
	BearerTokenFile string //adds the Authorization header to API requests for authentication purposes.
	Retries         int    //How many times to retry each query when getting an error response from Loki.
	QueryTags       string //adds X-Query-Tags header to API requests.
	Quiet           bool   //Suppress query metadata.
}

// NewLokiClient initializes a lokiClient with server address
func NewLokiClient(routeAddress string) *LokiClient {
	client := &LokiClient{}
	client.Address = routeAddress
	client.Retries = 5
	client.Quiet = true
	return client
}

// retry sets how many times to retry each query
func (c *LokiClient) Retry(retry int) *LokiClient {
	nc := *c
	nc.Retries = retry
	return &nc
}

// withToken sets the token used to do query
func (c *LokiClient) WithToken(bearerToken string) *LokiClient {
	nc := *c
	nc.BearerToken = bearerToken
	return &nc
}

func (c *LokiClient) WithBasicAuth(username, password string) *LokiClient {
	nc := *c
	nc.Username = username
	nc.Password = password
	return &nc
}

/*
func (c *lokiClient) withTokenFile(bearerTokenFile string) *lokiClient {
	nc := *c
	nc.bearerTokenFile = bearerTokenFile
	return &nc
}
*/

func (c *LokiClient) getHTTPRequestHeader() (http.Header, error) {
	h := make(http.Header)
	if c.Username != "" && c.Password != "" {
		h.Set(
			"Authorization",
			"Basic "+base64.StdEncoding.EncodeToString([]byte(c.Username+":"+c.Password)),
		)
	}
	h.Set("User-Agent", "loki-logcli")

	if c.OrgID != "" {
		h.Set("X-Scope-OrgID", c.OrgID)
	}

	if c.QueryTags != "" {
		h.Set("X-Query-Tags", c.QueryTags)
	}

	if (c.Username != "" || c.Password != "") && (len(c.BearerToken) > 0 || len(c.BearerTokenFile) > 0) {
		return nil, fmt.Errorf("at most one of HTTP basic auth (username/password), bearer-token & bearer-token-file is allowed to be configured")
	}

	if len(c.BearerToken) > 0 && len(c.BearerTokenFile) > 0 {
		return nil, fmt.Errorf("at most one of the options bearer-token & bearer-token-file is allowed to be configured")
	}

	if c.BearerToken != "" {
		h.Set("Authorization", "Bearer "+c.BearerToken)
	}

	if c.BearerTokenFile != "" {
		b, err := os.ReadFile(c.BearerTokenFile)
		if err != nil {
			return nil, fmt.Errorf("unable to read authorization credentials file %s: %s", c.BearerTokenFile, err)
		}
		bearerToken := strings.TrimSpace(string(b))
		h.Set("Authorization", "Bearer "+bearerToken)
	}
	return h, nil
}

func (c *LokiClient) doRequest(path, query string, out interface{}) error {
	h, err := c.getHTTPRequestHeader()
	if err != nil {
		return err
	}

	resp, err := DoHTTPRequest(h, c.Address, path, query, "GET", c.Quiet, c.Retries, nil, 200)
	if err != nil {
		return err
	}
	return json.Unmarshal(resp, out)
}

func (c *LokiClient) doQuery(path string, query string) (*lokiQueryResponse, error) {
	var err error
	var r lokiQueryResponse

	if err = c.doRequest(path, query, &r); err != nil {
		return nil, err
	}

	return &r, nil
}

// query uses the /api/v1/query endpoint to execute an instant query
// lc.query("application", "sum by(kubernetes_namespace_name)(count_over_time({kubernetes_namespace_name=\"multiple-containers\"}[5m]))", 30, false, time.Now())
func (c *LokiClient) Query(tenant string, queryStr string, limit int, forward bool, time time.Time) (*lokiQueryResponse, error) {
	direction := func() string {
		if forward {
			return "FORWARD"
		}
		return "BACKWARD"
	}
	qsb := newQueryStringBuilder()
	qsb.setString("query", queryStr)
	qsb.setInt("limit", int64(limit))
	qsb.setInt("time", time.UnixNano())
	qsb.setString("direction", direction())
	var logPath string
	if len(tenant) > 0 {
		logPath = apiPath + tenant + queryRangePath
	} else {
		logPath = queryRangePath
	}
	return c.doQuery(logPath, qsb.encode())
}

// queryRange uses the /api/v1/query_range endpoint to execute a range query
// tenant: application, infrastructure, audit
// queryStr: string to filter logs, for example: "{kubernetes_namespace_name="test"}"
// limit: max log count
// start: Start looking for logs at this absolute time(inclusive), e.g.: time.Now().Add(time.Duration(-1)*time.Hour) means 1 hour ago
// end: Stop looking for logs at this absolute time (exclusive)
// forward: true means scan forwards through logs, false means scan backwards through logs
func (c *LokiClient) QueryRange(tenant string, queryStr string, limit int, start, end time.Time, forward bool) (*lokiQueryResponse, error) {
	direction := func() string {
		if forward {
			return "FORWARD"
		}
		return "BACKWARD"
	}
	params := newQueryStringBuilder()
	params.setString("query", queryStr)
	params.setInt32("limit", limit)
	params.setInt("start", start.UnixNano())
	params.setInt("end", end.UnixNano())
	params.setString("direction", direction())
	var logPath string
	if len(tenant) > 0 {
		logPath = apiPath + tenant + queryRangePath
	} else {
		logPath = queryRangePath
	}

	return c.doQuery(logPath, params.encode())
}

func (c *LokiClient) SearchLogsInLoki(tenant, query string) (*lokiQueryResponse, error) {
	res, err := c.QueryRange(tenant, query, 5, time.Now().Add(time.Duration(-1)*time.Hour), time.Now(), false)
	return res, err
}

func (c *LokiClient) WaitForLogsAppearByQuery(tenant, query string) error {
	return wait.PollUntilContextTimeout(context.Background(), 10*time.Second, 300*time.Second, true, func(context.Context) (done bool, err error) {
		logs, err := c.SearchLogsInLoki(tenant, query)
		if err != nil {
			e2e.Logf("\ngot err when searching logs: %v, retrying...\n", err)
			return false, nil
		}
		if len(logs.Data.Result) > 0 {
			e2e.Logf(`find logs by %s`, query)
			return true, nil
		}
		return false, nil
	})
}

func (c *LokiClient) SearchByKey(tenant, key, value string) (*lokiQueryResponse, error) {
	res, err := c.SearchLogsInLoki(tenant, "{"+key+"=\""+value+"\"}")
	return res, err
}

func (c *LokiClient) WaitForLogsAppearByKey(tenant, key, value string) {
	err := wait.PollUntilContextTimeout(context.Background(), 10*time.Second, 300*time.Second, true, func(context.Context) (done bool, err error) {
		logs, err := c.SearchByKey(tenant, key, value)
		if err != nil {
			e2e.Logf("\ngot err when searching logs: %v, retrying...\n", err)
			return false, nil
		}
		if len(logs.Data.Result) > 0 {
			e2e.Logf(`find logs by {%s="%s"}`, key, value)
			return true, nil
		}
		return false, nil
	})
	compat_otp.AssertWaitPollNoErr(err, fmt.Sprintf(`can't find logs by {%s="%s"} in last 5 minutes`, key, value))
}

func (c *LokiClient) SearchByNamespace(tenant, projectName string) (*lokiQueryResponse, error) {
	res, err := c.SearchLogsInLoki(tenant, "{kubernetes_namespace_name=\""+projectName+"\"}")
	return res, err
}

func (c *LokiClient) WaitForLogsAppearByProject(tenant, projectName string) {
	err := wait.PollUntilContextTimeout(context.Background(), 10*time.Second, 300*time.Second, true, func(context.Context) (done bool, err error) {
		logs, err := c.SearchByNamespace(tenant, projectName)
		if err != nil {
			e2e.Logf("\ngot err when searching logs: %v, retrying...\n", err)
			return false, nil
		}
		if len(logs.Data.Result) > 0 {
			e2e.Logf("find logs from %s project", projectName)
			return true, nil
		}
		return false, nil
	})
	compat_otp.AssertWaitPollNoErr(err, fmt.Sprintf("can't find logs from %s project in last 5 minutes", projectName))
}

// ExtractLogEntities extract the log entities from loki query response, designed for checking the content of log data in Loki
func ExtractLogEntities(lokiQueryResult *lokiQueryResponse) []LogEntity {
	var lokiLogs []LogEntity
	for _, res := range lokiQueryResult.Data.Result {
		for _, value := range res.Values {
			lokiLog := LogEntity{}
			// only process log data, drop timestamp
			json.Unmarshal([]byte(ConvertInterfaceToArray(value)[1]), &lokiLog)
			lokiLogs = append(lokiLogs, lokiLog)
		}
	}
	return lokiLogs
}

// listLabelValues uses the /api/v1/label endpoint to list label values
func (c *LokiClient) ListLabelValues(tenant, name string, start, end time.Time) (*labelResponse, error) {
	lpath := fmt.Sprintf(labelValuesPath, url.PathEscape(name))
	var labelResponse labelResponse
	params := newQueryStringBuilder()
	params.setInt("start", start.UnixNano())
	params.setInt("end", end.UnixNano())

	path := ""
	if len(tenant) > 0 {
		path = apiPath + tenant + lpath
	} else {
		path = lpath
	}

	if err := c.doRequest(path, params.encode(), &labelResponse); err != nil {
		return nil, err
	}
	return &labelResponse, nil
}

// ListLabelNames uses the /api/v1/label endpoint to list label names
func (c *LokiClient) ListLabelNames(tenant string, start, end time.Time) (*labelResponse, error) {
	var labelResponse labelResponse
	params := newQueryStringBuilder()
	params.setInt("start", start.UnixNano())
	params.setInt("end", end.UnixNano())
	path := ""
	if len(tenant) > 0 {
		path = apiPath + tenant + labelsPath
	} else {
		path = labelsPath
	}

	if err := c.doRequest(path, params.encode(), &labelResponse); err != nil {
		return nil, err
	}
	return &labelResponse, nil
}

// listLabels gets the label names or values
func (c *LokiClient) ListLabels(tenant, labelName string) ([]string, error) {
	var labelResponse *labelResponse
	var err error
	start := time.Now().Add(time.Duration(-2) * time.Hour)
	end := time.Now()
	if len(labelName) > 0 {
		labelResponse, err = c.ListLabelValues(tenant, labelName, start, end)
	} else {
		labelResponse, err = c.ListLabelNames(tenant, start, end)
	}
	return labelResponse.Data, err
}

func (c *LokiClient) QueryRules(tenant, ns string) ([]byte, error) {
	path := apiPath + tenant + rulesPath

	params := url.Values{}
	if ns != "" {
		params.Add("kubernetes_namespace_name", ns)
	}

	h, err := c.getHTTPRequestHeader()
	if err != nil {
		return nil, err
	}

	resp, err := DoHTTPRequest(h, c.Address, path, params.Encode(), "GET", c.Quiet, c.Retries, nil, 200)
	if err != nil {
		/*
			Ignore error "unexpected EOF", adding `h.Add("Accept-Encoding", "identity")` doesn't resolve the error.
			This seems to be an issue in lokistack when tenant=application, recording rules are not in the response.
			No error when tenant=infrastructure
		*/
		if strings.Contains(err.Error(), "unexpected EOF") && len(resp) > 0 {
			e2e.Logf("got error %s when reading the response, but ignore it", err.Error())
			return resp, nil
		}
		return nil, err
	}
	return resp, nil

}

type queryStringBuilder struct {
	values url.Values
}

func newQueryStringBuilder() *queryStringBuilder {
	return &queryStringBuilder{
		values: url.Values{},
	}
}

func (b *queryStringBuilder) setString(name, value string) {
	b.values.Set(name, value)
}

func (b *queryStringBuilder) setInt(name string, value int64) {
	b.setString(name, strconv.FormatInt(value, 10))
}

func (b *queryStringBuilder) setInt32(name string, value int) {
	b.setString(name, strconv.Itoa(value))
}

/*
func (b *queryStringBuilder) setStringArray(name string, values []string) {
	for _, v := range values {
		b.values.Add(name, v)
	}
}
func (b *queryStringBuilder) setFloat32(name string, value float32) {
	b.setString(name, strconv.FormatFloat(float64(value), 'f', -1, 32))
}
func (b *queryStringBuilder) setFloat(name string, value float64) {
	b.setString(name, strconv.FormatFloat(value, 'f', -1, 64))
}
*/

// encode returns the URL-encoded query string based on key-value
// parameters added to the builder calling Set functions.
func (b *queryStringBuilder) encode() string {
	return b.values.Encode()
}

// CompareClusterResources compares the remaning resource with the requested resource provide by user
func CompareClusterResources(oc *exutil.CLI, cpu, memory string) bool {
	nodes, err := compat_otp.GetSchedulableLinuxWorkerNodes(oc)
	o.Expect(err).NotTo(o.HaveOccurred())
	var remainingCPU, remainingMemory int64
	re := compat_otp.GetRemainingResourcesNodesMap(oc, nodes)
	for _, node := range nodes {
		remainingCPU += re[node.Name].CPU
		remainingMemory += re[node.Name].Memory
	}

	requiredCPU, _ := k8sresource.ParseQuantity(cpu)
	requiredMemory, _ := k8sresource.ParseQuantity(memory)
	e2e.Logf("the required cpu is: %d, and the required memory is: %d", requiredCPU.MilliValue(), requiredMemory.MilliValue())
	e2e.Logf("the remaining cpu is: %d, and the remaning memory is: %d", remainingCPU, remainingMemory)
	return remainingCPU > requiredCPU.MilliValue() && remainingMemory > requiredMemory.MilliValue()
}

// ValidateInfraForLoki checks platform type
// supportedPlatforms the platform types which the case can be executed on, if it's empty, then skip this check
func ValidateInfraForLoki(oc *exutil.CLI, supportedPlatforms ...string) bool {
	currentPlatform := compat_otp.CheckPlatform(oc)
	if len(supportedPlatforms) > 0 {
		return Contain(supportedPlatforms, currentPlatform)
	}
	return true
}

// ValidateInfraAndResourcesForLoki checks cluster remaning resources and platform type
// supportedPlatforms the platform types which the case can be executed on, if it's empty, then skip this check
func ValidateInfraAndResourcesForLoki(oc *exutil.CLI, reqMemory, reqCPU string, supportedPlatforms ...string) bool {
	return ValidateInfraForLoki(oc, supportedPlatforms...) && CompareClusterResources(oc, reqCPU, reqMemory)
}

func DeployMinIO(oc *exutil.CLI) {
	// create namespace
	_, err := oc.AdminKubeClient().CoreV1().Namespaces().Get(context.Background(), minioNS, metav1.GetOptions{})
	if apierrors.IsNotFound(err) {
		err = oc.AsAdmin().WithoutNamespace().Run("create").Args("namespace", minioNS).Execute()
		o.Expect(err).NotTo(o.HaveOccurred())
	}
	// create secret
	_, err = oc.AdminKubeClient().CoreV1().Secrets(minioNS).Get(context.Background(), minioSecret, metav1.GetOptions{})
	if apierrors.IsNotFound(err) {
		err = oc.AsAdmin().WithoutNamespace().Run("create").Args("secret", "generic", minioSecret, "-n", minioNS, "--from-literal=access_key_id="+GetRandomString(), "--from-literal=secret_access_key=passwOOrd"+GetRandomString()).Execute()
		o.Expect(err).NotTo(o.HaveOccurred())
	}
	// deploy minIO
	clusterDomain, err := oc.AsAdmin().WithoutNamespace().Run("get").Args("ingress.config/cluster", "-o=jsonpath={.spec.domain}").Output()
	o.Expect(err).NotTo(o.HaveOccurred())
	minioDomain := "logging-s3-minio." + clusterDomain
	deployTemplate := compat_otp.FixturePath("testdata", "logging", "minIO", "deploy.yaml")
	deployFile, err := ProcessTemplate(oc, "-n", minioNS, "-f", deployTemplate, "-p", "NAMESPACE="+minioNS, "NAME=minio", "SECRET_NAME="+minioSecret, "MINIO_DOMAIN="+minioDomain)
	defer os.Remove(deployFile)
	o.Expect(err).NotTo(o.HaveOccurred())
	err = oc.AsAdmin().Run("apply").Args("-f", deployFile, "-n", minioNS).Execute()
	o.Expect(err).NotTo(o.HaveOccurred())
	// wait for minio to be ready
	for _, rs := range []string{"deployment", "svc", "route"} {
		o.Expect(WaitForResourceToAppear(oc, rs, "minio", minioNS)).NotTo(o.HaveOccurred())
	}
	WaitForDeploymentPodsToBeReady(oc, minioNS, "minio")
}

/*
func removeMinIO(oc *exutil.CLI) {
	deleteNamespace(oc, minioNS)
}
*/

// QueryAlertManagerForActiveAlerts queries user-workload alert-manager if isUserWorkloadAM parameter is true.
// All active alerts should be returned when querying Alert Managers
func QueryAlertManagerForActiveAlerts(oc *exutil.CLI, token string, isUserWorkloadAM bool, alertName string, timeInMinutes int) {
	var err error
	if !isUserWorkloadAM {
		alertManagerRoute := GetRouteAddress(oc, "openshift-monitoring", "alertmanager-main")
		h := make(http.Header)
		h.Add("Content-Type", "application/json")
		h.Add("Authorization", "Bearer "+token)
		params := url.Values{}
		err = wait.PollUntilContextTimeout(context.Background(), 30*time.Second, time.Duration(timeInMinutes)*time.Minute, true, func(context.Context) (done bool, err error) {
			resp, err := DoHTTPRequest(h, "https://"+alertManagerRoute, "/api/v2/alerts", params.Encode(), "GET", true, 5, nil, 200)
			if err != nil {
				return false, err
			}
			if strings.Contains(string(resp), alertName) {
				return true, nil
			}
			e2e.Logf("Waiting for alert %s to be in Firing state", alertName)
			return false, nil
		})

	} else {
		userWorkloadAlertManagerURL := "https://alertmanager-user-workload.openshift-user-workload-monitoring.svc:9095/api/v2/alerts"
		authBearer := " \"Authorization: Bearer " + token + "\""
		cmd := "curl -k -H" + authBearer + " " + userWorkloadAlertManagerURL
		err = wait.PollUntilContextTimeout(context.Background(), 30*time.Second, time.Duration(timeInMinutes)*time.Minute, true, func(context.Context) (done bool, err error) {
			alerts, err := compat_otp.RemoteShPod(oc, "openshift-monitoring", "prometheus-k8s-0", "/bin/sh", "-x", "-c", cmd)
			if err != nil {
				return false, err
			}
			if strings.Contains(string(alerts), alertName) {
				return true, nil
			}
			e2e.Logf("Waiting for alert %s to be in Firing state", alertName)
			return false, nil
		})
	}

	compat_otp.AssertWaitPollNoErr(err, fmt.Sprintf("Alert %s is not firing after %d minutes", alertName, timeInMinutes))
}

// EnableUserWorkloadMonitoringForLogging deletes cluster-monitoring-config and user-workload-monitoring-config if exists and recreates configmaps.
// DeleteUserWorkloadManifests() should be called once resources are created by EnableUserWorkloadMonitoringForLogging()
func EnableUserWorkloadMonitoringForLogging(oc *exutil.CLI) {
	oc.AsAdmin().WithoutNamespace().Run("delete").Args("ConfigMap", "cluster-monitoring-config", "-n", "openshift-monitoring", "--ignore-not-found").Execute()
	clusterMonitoringConfigPath := compat_otp.FixturePath("testdata", "logging", "loki-log-alerts", "cluster-monitoring-config.yaml")
	o.Expect(ApplyResourceFromTemplate(oc, "openshift-monitoring", "-f", clusterMonitoringConfigPath)).NotTo(o.HaveOccurred())
	o.Expect(WaitForResourceToAppear(oc, "configmap", "cluster-monitoring-config", "openshift-monitoring")).NotTo(o.HaveOccurred())

	oc.AsAdmin().WithoutNamespace().Run("delete").Args("ConfigMap", "user-workload-monitoring-config", "-n", "openshift-user-workload-monitoring", "--ignore-not-found").Execute()
	userWorkloadMConfigPath := compat_otp.FixturePath("testdata", "logging", "loki-log-alerts", "user-workload-monitoring-config.yaml")
	o.Expect(ApplyResourceFromTemplate(oc, "openshift-user-workload-monitoring", "-f", userWorkloadMConfigPath)).NotTo(o.HaveOccurred())
	o.Expect(WaitForResourceToAppear(oc, "configmap", "user-workload-monitoring-config", "openshift-user-workload-monitoring")).NotTo(o.HaveOccurred())
}

func DeleteUserWorkloadManifests(oc *exutil.CLI) {
	o.Expect(DeleteResourceFromCluster(oc, "configmap", "cluster-monitoring-config", "openshift-monitoring")).NotTo(o.HaveOccurred())
	o.Expect(DeleteResourceFromCluster(oc, "configmap", "user-workload-monitoring-config", "openshift-user-workload-monitoring")).NotTo(o.HaveOccurred())
}

// ValidateCredentialsRequestGenerationOnSTS to check CredentialsRequest is generated by Loki Operator on STS clusters for CCO flow
func ValidateCredentialsRequestGenerationOnSTS(oc *exutil.CLI, lokiStackName, lokiNamespace string) {
	compat_otp.By("Validate that Loki Operator creates a CredentialsRequest object")
	err := oc.AsAdmin().WithoutNamespace().Run("get").Args("CredentialsRequest", lokiStackName, "-n", lokiNamespace).Execute()
	o.Expect(err).NotTo(o.HaveOccurred())
	cloudTokenPath, err := oc.AsAdmin().WithoutNamespace().Run("get").Args("CredentialsRequest", lokiStackName, "-n", lokiNamespace, `-o=jsonpath={.spec.cloudTokenPath}`).Output()
	o.Expect(err).NotTo(o.HaveOccurred())
	o.Expect(cloudTokenPath).Should(o.Equal("/var/run/secrets/storage/serviceaccount/token"))
	serviceAccountNames, err := oc.AsAdmin().WithoutNamespace().Run("get").Args("CredentialsRequest", lokiStackName, "-n", lokiNamespace, `-o=jsonpath={.spec.serviceAccountNames}`).Output()
	o.Expect(err).NotTo(o.HaveOccurred())
	o.Expect(serviceAccountNames).Should(o.Equal(fmt.Sprintf(`["%s","%s-ruler"]`, lokiStackName, lokiStackName)))
}

// CreateLokiClusterRolesForReadAccess creates the cluster roles 'cluster-logging-application-view', 'cluster-logging-infrastructure-view' and 'cluster-logging-audit-view' introduced
// for fine grained read access to LokiStack logs. The ownership of these roles is moved to Cluster Observability Operator (COO) from Cluster Logging Operator (CLO) in Logging 6.0+
func CreateLokiClusterRolesForReadAccess(oc *exutil.CLI) {
	rbacFile := compat_otp.FixturePath("testdata", "logging", "lokistack", "fine-grained-access-roles.yaml")
	msg, err := oc.AsAdmin().WithoutNamespace().Run("apply").Args("-f", rbacFile).Output()
	o.Expect(err).NotTo(o.HaveOccurred(), msg)
}

func DeleteLokiClusterRolesForReadAccess(oc *exutil.CLI) {
	roles := []string{"cluster-logging-application-view", "cluster-logging-infrastructure-view", "cluster-logging-audit-view"}
	for _, role := range roles {
		msg, err := oc.AsAdmin().WithoutNamespace().Run("delete").Args("clusterrole", role).Output()
		if err != nil {
			e2e.Logf("Failed to delete Loki RBAC role '%s': %s", role, msg)
		}
	}
}

// Patches Loki Operator running on a GCP WIF cluster. Operator is deployed with CCO mode after patching.
func PatchLokiOperatorOnGCPSTSforCCO(oc *exutil.CLI, namespace string, projectNumber string, poolID string, serviceAccount string) {
	patchConfig := `{
    	"spec": {
        	"config": {
            	"env": [
               		{
                    	"name": "PROJECT_NUMBER",
                    	"value": "%s"
                	},
                	{
                    	"name": "POOL_ID",
                    	"value": "%s"
                	},
                	{
                    	"name": "PROVIDER_ID",
                    	"value": "%s"
                	},
                	{
                    	"name": "SERVICE_ACCOUNT_EMAIL",
                    	"value": "%s"
                	}
            	]
        	}
    	}
	}`

	err := oc.NotShowInfo().AsAdmin().WithoutNamespace().Run("patch").Args("sub", "loki-operator", "-n", namespace, "-p", fmt.Sprintf(patchConfig, projectNumber, poolID, poolID, serviceAccount), "--type=merge").Execute()
	o.Expect(err).NotTo(o.HaveOccurred())
	WaitForPodReadyByLabel(oc, LoNS, "name=loki-operator-controller-manager")
}

// CompareExpectedTLSConfigWithCurrent compare the expected TLS config with the current TLS config on apiserver/cluster. Compares the spec.tlsSecurityProfile attribute.
func CompareExpectedTLSConfigWithCurrent(oc *exutil.CLI, expectedTLSConfig string) bool {
	currentTLSConfig, err := oc.AsAdmin().WithoutNamespace().Run("get").Args("apiserver/cluster", "-o", "jsonpath={.spec.tlsSecurityProfile}").Output()
	o.Expect(err).NotTo(o.HaveOccurred())

	return currentTLSConfig == expectedTLSConfig
}

// Function to check if logs are present under the MinIO bucket.
// Returns success if any one of the tenants under tenants[] are found.
func ValidatesIfLogsArePushedToMinIOBucket(oc *exutil.CLI, bucketName string, tenants []string) {
	// Build an S3 client pointing to the in-cluster MinIO endpoint
	cred := getMinIOCreds(oc, minioNS)
	factory, err := NewAWSClientFactory(context.TODO(), cred)
	if err != nil {
		e2e.Failf("error loading aws config: %v", err)
	}
	s3Client := factory.S3()
	o.Expect(err).NotTo(o.HaveOccurred())

	// Poll to check contents of the MinIO bucket
	err = wait.PollUntilContextTimeout(context.Background(), 30*time.Second, 300*time.Second, true, func(context.Context) (done bool, err error) {
		listObjectsOutput, err := s3Client.ListObjectsV2(context.TODO(), &s3.ListObjectsV2Input{
			Bucket: aws.String(bucketName),
		})
		if err != nil {
			return false, err
		}

		for _, object := range listObjectsOutput.Contents {
			for _, tenantName := range tenants {
				if strings.Contains(*object.Key, tenantName) {
					e2e.Logf("Logs %s found under the minio bucket: %s", *object.Key, bucketName)
					return true, nil
				}
			}
		}
		e2e.Logf("Waiting for data to be available under bucket: %s", bucketName)
		return false, nil
	})
	compat_otp.AssertWaitPollNoErr(err, "Timed out...No data is available under the bucket: "+bucketName)
}

// Creates Loki object storage secret on AWS STS cluster
func CreateObjectStorageSecretOnAWSSTSCluster(oc *exutil.CLI, region, storageSecret, bucketName, namespace string) {
	err := oc.NotShowInfo().AsAdmin().WithoutNamespace().Run("create").Args("secret", "generic", storageSecret, "--from-literal=region="+region, "--from-literal=bucketnames="+bucketName, "--from-literal=audience=openshift", "-n", namespace).Execute()
	o.Expect(err).NotTo(o.HaveOccurred())
}
