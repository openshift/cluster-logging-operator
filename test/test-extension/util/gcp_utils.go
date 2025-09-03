package util

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"cloud.google.com/go/logging"
	"cloud.google.com/go/logging/logadmin"
	"cloud.google.com/go/storage"
	g "github.com/onsi/ginkgo/v2"
	o "github.com/onsi/gomega"
	exutil "github.com/openshift/origin/test/extended/util"
	"github.com/openshift/origin/test/extended/util/compat_otp"
	"google.golang.org/api/cloudresourcemanager/v1"
	"google.golang.org/api/iam/v1"
	"google.golang.org/api/iterator"
	"google.golang.org/protobuf/types/known/structpb"
	"k8s.io/apimachinery/pkg/util/wait"
	e2e "k8s.io/kubernetes/test/e2e/framework"
)

type GoogleApplicationCredentials struct {
	CredentialType string `json:"type"`
	ProjectID      string `json:"project_id"`
	ClientID       string `json:"client_id"`
}

type GoogleCloudLogging struct {
	ProjectID string
	LogName   string
}

func GetGCPProjectID(oc *exutil.CLI) (string, error) {
	platform := CheckPlatform(oc)
	if platform == "gcp" {
		return GetGcpProjectID(oc)
	}

	credentialFile, present := os.LookupEnv("GOOGLE_APPLICATION_CREDENTIALS")
	if !present {
		g.Skip("Skip for the platform is not GCP and there is no GCP credentials")
	}
	file, err := os.ReadFile(credentialFile)
	if err != nil {
		g.Skip("Skip for the platform is not GCP and can't read google application credentials: " + err.Error())
	}
	var gac GoogleApplicationCredentials
	err = json.Unmarshal(file, &gac)
	return gac.ProjectID, err
}

// GetGcpProjectID returns the gcp project id
func GetGcpProjectID(oc *exutil.CLI) (string, error) {
	projectID, err := oc.AsAdmin().WithoutNamespace().Run("get").Args("infrastructure", "cluster", "-o", "jsonpath='{.status.platformStatus.gcp.projectID}'").Output()
	if err != nil {
		return "", err
	}
	return strings.Trim(projectID, "'"), err
}

// listLogEntries gets the most recent 5 entries
// example: https://cloud.google.com/logging/docs/reference/libraries#list_log_entries
// https://github.com/GoogleCloudPlatform/golang-samples/blob/HEAD/logging/simplelog/simplelog.go
func (gcl GoogleCloudLogging) ListLogEntries(queryString string) ([]*logging.Entry, error) {
	ctx := context.Background()

	adminClient, err := logadmin.NewClient(ctx, gcl.ProjectID)
	if err != nil {
		e2e.Logf("Failed to create logadmin client: %v", err)
	}
	defer adminClient.Close()

	var entries []*logging.Entry
	lastHour := time.Now().Add(-1 * time.Hour).Format(time.RFC3339)

	filter := fmt.Sprintf(`logName = "projects/%s/logs/%s" AND timestamp > "%s"`, gcl.ProjectID, gcl.LogName, lastHour)
	if len(queryString) > 0 {
		filter += queryString
	}

	iter := adminClient.Entries(ctx,
		logadmin.Filter(filter),
		// Get most recent entries first.
		logadmin.NewestFirst(),
	)

	// Fetch the most recent 5 entries.
	for len(entries) < 5 {
		entry, err := iter.Next()
		if err == iterator.Done {
			return entries, nil
		}
		if err != nil {
			return nil, err
		}
		entries = append(entries, entry)
	}
	return entries, nil
}

func (gcl GoogleCloudLogging) GetLogByType(logType string) ([]*logging.Entry, error) {
	searchString := " AND jsonPayload.log_type = \"" + logType + "\""
	return gcl.ListLogEntries(searchString)
}

func (gcl GoogleCloudLogging) GetLogByNamespace(namespace string) ([]*logging.Entry, error) {
	searchString := " AND jsonPayload.kubernetes.namespace_name = \"" + namespace + "\""
	return gcl.ListLogEntries(searchString)
}

func ExtractGoogleCloudLoggingLogs(gclLogs []*logging.Entry) ([]LogEntity, error) {
	var (
		logs []LogEntity
		log  LogEntity
	)
	for _, item := range gclLogs {
		if value, ok := item.Payload.(*structpb.Struct); ok {
			v, err := value.MarshalJSON()
			if err != nil {
				return nil, err
			}
			//e2e.Logf("\noriginal log:\n%s\n\n", string(v))
			err = json.Unmarshal(v, &log)
			if err != nil {
				return nil, err
			}
			logs = append(logs, log)
		}
	}
	return logs, nil
}

func (gcl GoogleCloudLogging) RemoveLogs() error {
	ctx := context.Background()

	adminClient, err := logadmin.NewClient(ctx, gcl.ProjectID)
	if err != nil {
		e2e.Logf("Failed to create logadmin client: %v", err)
	}
	defer adminClient.Close()

	return adminClient.DeleteLog(ctx, gcl.LogName)
}

func (gcl GoogleCloudLogging) WaitForLogsAppearByType(logType string) error {
	return wait.PollUntilContextTimeout(context.Background(), 30*time.Second, 180*time.Second, true, func(context.Context) (done bool, err error) {
		logs, err := gcl.GetLogByType(logType)
		if err != nil {
			return false, err
		}
		return len(logs) > 0, nil
	})
}

func (gcl GoogleCloudLogging) WaitForLogsAppearByNamespace(namespace string) error {
	return wait.PollUntilContextTimeout(context.Background(), 30*time.Second, 180*time.Second, true, func(context.Context) (done bool, err error) {
		logs, err := gcl.GetLogByNamespace(namespace)
		if err != nil {
			return false, err
		}
		return len(logs) > 0, nil
	})
}

// createSecretForGCL creates a secret for collector pods to forward logs to Google Cloud Logging
func CreateSecretForGCL(oc *exutil.CLI, name, namespace string) error {
	// get gcp-credentials from env var GOOGLE_APPLICATION_CREDENTIALS
	gcsCred := os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")
	return oc.AsAdmin().WithoutNamespace().Run("create").Args("secret", "generic", name, "-n", namespace, "--from-file=google-application-credentials.json="+gcsCred).Execute()
}

// CreateGCSBucket creates a GCS bucket in a project
func CreateGCSBucket(projectID, bucketName string) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	// initialize the GCS client, the credentials are got from the env var GOOGLE_APPLICATION_CREDENTIALS
	client, err := storage.NewClient(ctx)
	if err != nil {
		return fmt.Errorf("storage.NewClient: %v", err)
	}
	defer client.Close()

	// check if the bucket exists or not
	// if exists, clear all the objects in the bucket
	// if not, create the bucket
	exist := false
	buckets, err := ListGCSBuckets(*client, projectID)
	if err != nil {
		return err
	}
	for _, bu := range buckets {
		if bu == bucketName {
			exist = true
			break
		}
	}
	if exist {
		return EmptyGCSBucket(*client, bucketName)
	}

	bucket := client.Bucket(bucketName)
	if err := bucket.Create(ctx, projectID, &storage.BucketAttrs{}); err != nil {
		return fmt.Errorf("Bucket(%q).Create: %v", bucketName, err)
	}
	fmt.Printf("Created bucket %v\n", bucketName)
	return nil
}

// ListGCSBuckets gets all the bucket names under the projectID
func ListGCSBuckets(client storage.Client, projectID string) ([]string, error) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	var buckets []string
	it := client.Buckets(ctx, projectID)
	for {
		battrs, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}
		buckets = append(buckets, battrs.Name)
	}
	return buckets, nil
}

// EmptyGCSBucket removes all the objects in the bucket
func EmptyGCSBucket(client storage.Client, bucketName string) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	bucket := client.Bucket(bucketName)
	it := bucket.Objects(ctx, nil)
	for {
		objAttrs, err := it.Next()
		if err != nil && err != iterator.Done {
			return fmt.Errorf("can't get objects in bucket %s: %v", bucketName, err)
		}
		if err == iterator.Done {
			break
		}
		if err := bucket.Object(objAttrs.Name).Delete(ctx); err != nil {
			return fmt.Errorf("Object(%q).Delete: %v", objAttrs.Name, err)
		}
	}
	e2e.Logf("deleted all object items in the bucket %s.", bucketName)
	return nil
}

// DeleteGCSBucket deletes the GCS bucket
func DeleteGCSBucket(bucketName string) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	client, err := storage.NewClient(ctx)
	if err != nil {
		return fmt.Errorf("storage.NewClient: %v", err)
	}
	defer client.Close()

	// remove objects
	err = EmptyGCSBucket(*client, bucketName)
	if err != nil {
		return err
	}
	bucket := client.Bucket(bucketName)
	if err := bucket.Delete(ctx); err != nil {
		return fmt.Errorf("Bucket(%q).Delete: %v", bucketName, err)
	}
	e2e.Logf("Bucket %v is deleted\n", bucketName)
	return nil
}

// ValidatesIfLogsArePushedToGCSBucket check if tenant logs are present under the Google Cloud Storage bucket.
// Returns success if any one of the tenants under tenants[] are found.
func ValidatesIfLogsArePushedToGCSBucket(bucketName string, tenants []string) {
	// Create a new GCS client
	client, err := storage.NewClient(context.Background())
	o.Expect(err).NotTo(o.HaveOccurred(), "Failed to create GCS client")

	// Get a reference to the bucket
	bucket := client.Bucket(bucketName)

	// Create a query to list objects in the bucket
	query := &storage.Query{}

	// List objects in the bucket and check for tenant object
	err = wait.PollUntilContextTimeout(context.Background(), 30*time.Second, 300*time.Second, true, func(context.Context) (done bool, err error) {
		itr := bucket.Objects(context.Background(), query)
		for {
			objAttrs, err := itr.Next()
			if err == iterator.Done {
				break
			}
			if err != nil {
				return false, err
			}
			for _, tenantName := range tenants {
				if strings.Contains(objAttrs.Name, tenantName) {
					e2e.Logf("Logs %s found under the bucket: %s", objAttrs.Name, bucketName)
					return true, nil
				}
			}
		}
		e2e.Logf("Waiting for data to be available under bucket: %s", bucketName)
		return false, nil
	})
	compat_otp.AssertWaitPollNoErr(err, "Timed out...No data is available under the bucket: "+bucketName)
}

func GetGCPProjectNumber(projectID string) (string, error) {
	crmService, err := cloudresourcemanager.NewService(context.Background())
	if err != nil {
		return "", err
	}

	project, err := crmService.Projects.Get(projectID).Do()
	if err != nil {
		return "", err
	}
	return strconv.FormatInt(project.ProjectNumber, 10), nil
}

func GetGCPAudience(providerName string) (string, error) {
	ctx := context.Background()
	service, err := iam.NewService(ctx)

	if err != nil {
		return "", fmt.Errorf("iam.NewService: %w", err)
	}
	audience, err := service.Projects.Locations.WorkloadIdentityPools.Providers.Get(providerName).Do()
	if err != nil {
		return "", fmt.Errorf("can't get audience: %v", err)
	}
	return audience.Oidc.AllowedAudiences[0], nil

}

func GenerateServiceAccountNameForGCS(clusterName string) string {
	// Service Account should be between 6-30 characters long
	name := clusterName + GetRandomString()
	if len(name) > 30 {
		return (name[0:30])
	}
	return name
}

func CreateServiceAccountOnGCP(projectID, name string) (*iam.ServiceAccount, error) {
	ctx := context.Background()
	service, err := iam.NewService(ctx)
	if err != nil {
		return nil, fmt.Errorf("iam.NewService: %w", err)
	}

	request := &iam.CreateServiceAccountRequest{
		AccountId: name,
		ServiceAccount: &iam.ServiceAccount{
			DisplayName: "Service Account for " + name,
		},
	}
	account, err := service.Projects.ServiceAccounts.Create("projects/"+projectID, request).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to create serviceaccount: %w", err)
	}
	e2e.Logf("create serviceaccount: %s successfully", account.Name)
	return account, nil
}

// ref: https://github.com/GoogleCloudPlatform/golang-samples/blob/main/iam/quickstart/quickstart.go
func AddBinding(projectID, member, role string) error {
	crmService, err := cloudresourcemanager.NewService(context.Background())
	if err != nil {
		return fmt.Errorf("cloudresourcemanager.NewService: %v", err)
	}

	err = wait.ExponentialBackoffWithContext(context.Background(), wait.Backoff{Steps: 5, Factor: 2, Duration: 5 * time.Second}, func(context.Context) (done bool, err error) {
		policy, err := GetPolicy(crmService, projectID)
		if err != nil {
			return false, fmt.Errorf("error getting policy: %v", err)
		}
		// Find the policy binding for role. Only one binding can have the role.
		var binding *cloudresourcemanager.Binding
		for _, b := range policy.Bindings {
			if b.Role == role {
				binding = b
				break
			}
		}
		if binding != nil {
			// If the binding exists, adds the member to the binding
			binding.Members = append(binding.Members, member)
		} else {
			// If the binding does not exist, adds a new binding to the policy
			binding = &cloudresourcemanager.Binding{
				Role:    role,
				Members: []string{member},
			}
			policy.Bindings = append(policy.Bindings, binding)
		}
		err = SetPolicy(crmService, projectID, policy)
		if err == nil {
			return true, nil
		}
		/*
			According to https://github.com/hashicorp/terraform-provider-google/issues/8280, deleting another serviceaccount can make 400 error happen, so retry this step when 400 error happens
		*/
		if strings.Contains(err.Error(), `googleapi: Error 409: There were concurrent policy changes. Please retry the whole read-modify-write with exponential backoff.`) ||
			(strings.Contains(err.Error(), "googleapi: Error 400: Service account") && strings.Contains(err.Error(), "does not exist., badRequest")) {
			e2e.Logf("Hit error: %v, retry the request", err)
			return false, nil
		}
		e2e.Logf("Failed to update polilcy: %v", err)
		return false, err
	})
	if err != nil {
		return fmt.Errorf("failed to add role %s to %s", role, member)
	}
	return nil
}

// RemoveMember removes the member from the project's IAM policy
func RemoveMember(projectID, member, role string) error {
	crmService, err := cloudresourcemanager.NewService(context.Background())
	if err != nil {
		return fmt.Errorf("cloudresourcemanager.NewService: %v", err)
	}
	err = wait.ExponentialBackoffWithContext(context.Background(), wait.Backoff{Steps: 5, Factor: 2, Duration: 5 * time.Second}, func(context.Context) (done bool, err error) {
		policy, err := GetPolicy(crmService, projectID)
		if err != nil {
			return false, fmt.Errorf("error getting policy: %v", err)
		}
		// Find the policy binding for role. Only one binding can have the role.
		var binding *cloudresourcemanager.Binding
		var bindingIndex int
		for i, b := range policy.Bindings {
			if b.Role == role {
				binding = b
				bindingIndex = i
				break
			}
		}

		if len(binding.Members) == 1 && binding.Members[0] == member {
			// If the member is the only member in the binding, removes the binding
			last := len(policy.Bindings) - 1
			policy.Bindings[bindingIndex] = policy.Bindings[last]
			policy.Bindings = policy.Bindings[:last]
		} else {
			// If there is more than one member in the binding, removes the member
			var memberIndex int
			var exist bool
			for i, mm := range binding.Members {
				if mm == member {
					memberIndex = i
					exist = true
					break
				}
			}
			if exist {
				last := len(policy.Bindings[bindingIndex].Members) - 1
				binding.Members[memberIndex] = binding.Members[last]
				binding.Members = binding.Members[:last]
			}
		}

		err = SetPolicy(crmService, projectID, policy)
		if err == nil {
			return true, nil
		}
		if strings.Contains(err.Error(), `googleapi: Error 409: There were concurrent policy changes. Please retry the whole read-modify-write with exponential backoff.`) ||
			(strings.Contains(err.Error(), "googleapi: Error 400: Service account") && strings.Contains(err.Error(), "does not exist., badRequest")) {
			e2e.Logf("Hit error: %v, retry the request", err)
			return false, nil
		}
		e2e.Logf("Failed to update polilcy: %v", err)
		return false, err
	})
	if err != nil {
		return fmt.Errorf("failed to remove %s", member)
	}
	return nil
}

// GetPolicy gets the project's IAM policy
func GetPolicy(crmService *cloudresourcemanager.Service, projectID string) (*cloudresourcemanager.Policy, error) {
	request := new(cloudresourcemanager.GetIamPolicyRequest)
	policy, err := crmService.Projects.GetIamPolicy(projectID, request).Do()
	if err != nil {
		return nil, err
	}
	return policy, nil
}

// SetPolicy sets the project's IAM policy
func SetPolicy(crmService *cloudresourcemanager.Service, projectID string, policy *cloudresourcemanager.Policy) error {
	request := new(cloudresourcemanager.SetIamPolicyRequest)
	request.Policy = policy
	_, err := crmService.Projects.SetIamPolicy(projectID, request).Do()
	return err
}

func GrantPermissionsToGCPServiceAccount(poolID, projectID, projectNumber, lokiNS, lokiStackName, serviceAccountEmail string) error {
	gcsRoles := []string{
		"roles/iam.workloadIdentityUser",
		"roles/storage.objectAdmin",
	}
	subjects := []string{
		"system:serviceaccount:" + lokiNS + ":" + lokiStackName,
		"system:serviceaccount:" + lokiNS + ":" + lokiStackName + "-ruler",
	}

	for _, role := range gcsRoles {
		err := AddBinding(projectID, "serviceAccount:"+serviceAccountEmail, role)
		if err != nil {
			return fmt.Errorf("error adding role %s to %s: %v", role, serviceAccountEmail, err)
		}
		for _, sub := range subjects {
			err := AddBinding(projectID, "principal://iam.googleapis.com/projects/"+projectNumber+"/locations/global/workloadIdentityPools/"+poolID+"/subject/"+sub, role)
			if err != nil {
				return fmt.Errorf("error adding role %s to %s: %v", role, sub, err)
			}
		}
	}
	return nil
}

func RemovePermissionsFromGCPServiceAccount(poolID, projectID, projectNumber, lokiNS, lokiStackName, serviceAccountEmail string) error {
	gcsRoles := []string{
		"roles/iam.workloadIdentityUser",
		"roles/storage.objectAdmin",
	}
	subjects := []string{
		"system:serviceaccount:" + lokiNS + ":" + lokiStackName,
		"system:serviceaccount:" + lokiNS + ":" + lokiStackName + "-ruler",
	}

	for _, role := range gcsRoles {
		err := RemoveMember(projectID, "serviceAccount:"+serviceAccountEmail, role)
		if err != nil {
			return fmt.Errorf("error removing role %s from %s: %v", role, serviceAccountEmail, err)
		}
		for _, sub := range subjects {
			err := RemoveMember(projectID, "principal://iam.googleapis.com/projects/"+projectNumber+"/locations/global/workloadIdentityPools/"+poolID+"/subject/"+sub, role)
			if err != nil {
				return fmt.Errorf("error removing role %s from %s: %v", role, sub, err)
			}
		}
	}
	return nil
}

func RemoveServiceAccountFromGCP(name string) error {
	ctx := context.Background()
	service, err := iam.NewService(ctx)
	if err != nil {
		return fmt.Errorf("iam.NewService: %w", err)
	}
	_, err = service.Projects.ServiceAccounts.Delete(name).Do()
	if err != nil {
		return fmt.Errorf("can't remove service account: %v", err)
	}
	return nil
}
