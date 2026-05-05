package tls

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	clolog "github.com/ViaQ/logerr/v2/log/static"
	configv1 "github.com/openshift/api/config/v1"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	"github.com/openshift/cluster-logging-operator/internal/runtime"
	internaltls "github.com/openshift/cluster-logging-operator/internal/tls"
	"github.com/openshift/cluster-logging-operator/internal/utils"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
)

const (
	// JobTimeout is the timeout for the TLS Scanner Job to complete
	JobTimeout = 10 * time.Minute
	// ImageEnvVar is the environment variable for the TLS Scanner image
	ImageEnvVar = "IMAGE_TLS_SCANNER"

	Name = "tls-scanner"
)

var (
	Components = strings.Join([]string{constants.ClusterLoggingOperator, "eventrouter", constants.LogfilesmetricexporterName, constants.VectorName}, ",")
)

// Scanner manages TLS Scanner deployment and result retrieval
type Scanner struct {
	KubeClient *kubernetes.Clientset
	CleanupFns *[]func() error
}

// NewScanner creates a new TLS Scanner instance
func NewScanner(kubeClient *kubernetes.Clientset, cleanupFns *[]func() error) *Scanner {
	return &Scanner{
		KubeClient: kubeClient,
		CleanupFns: cleanupFns,
	}
}

// GetImage returns the TLS Scanner image to use
func GetImage() (image string) {
	if image = os.Getenv(ImageEnvVar); image == "" {
		clolog.Info("No tls-scanner image provided", "variable", ImageEnvVar)
		os.Exit(1)
	}
	return image
}

// Deploy deploys the TLS Scanner as a Job to scan the target namespace
func (s *Scanner) Deploy(scannerNamespace, targetNamespace string) (*batchv1.Job, error) {
	clolog.Info("Deploying TLS Scanner", "scannerNamespace", scannerNamespace, "targetNamespace", targetNamespace)

	// Create ServiceAccount for the scanner
	sa := runtime.NewServiceAccount(scannerNamespace, "tls-scanner")
	if _, err := s.KubeClient.CoreV1().ServiceAccounts(scannerNamespace).Create(context.TODO(), sa, metav1.CreateOptions{}); err != nil {
		if !apierrors.IsAlreadyExists(err) {
			return nil, fmt.Errorf("failed to create ServiceAccount: %w", err)
		}
	}
	s.addCleanup(func() error {
		return s.KubeClient.CoreV1().ServiceAccounts(scannerNamespace).Delete(context.TODO(), sa.Name, metav1.DeleteOptions{})
	})

	// Create ClusterRoleBindings
	subject := rbacv1.Subject{
		Kind:      "ServiceAccount",
		Name:      sa.Name,
		Namespace: scannerNamespace,
	}
	for _, roleName := range []string{"cluster-reader", "system:openshift:scc:privileged", "dedicated-admins-project"} {
		crb := runtime.NewClusterRoleBinding(fmt.Sprintf("tls-scanner-%s-%s", scannerNamespace, roleName),
			rbacv1.RoleRef{
				APIGroup: "rbac.authorization.k8s.io",
				Kind:     "ClusterRole",
				Name:     roleName,
			},
			subject)
		if _, err := s.KubeClient.RbacV1().ClusterRoleBindings().Create(context.TODO(), crb, metav1.CreateOptions{}); err != nil {
			if !apierrors.IsAlreadyExists(err) {
				return nil, fmt.Errorf("failed to create ClusterRoleBinding: %w", err)
			}
		}
		s.addCleanup(func() error {
			return s.KubeClient.RbacV1().ClusterRoleBindings().Delete(context.TODO(), crb.Name, metav1.DeleteOptions{})
		})
	}

	// Create the Job
	backoffLimit := int32(0)

	job := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      Name,
			Namespace: scannerNamespace,
		},
		Spec: batchv1.JobSpec{
			BackoffLimit: &backoffLimit,
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					ServiceAccountName: sa.Name,
					RestartPolicy:      corev1.RestartPolicyNever,
					InitContainers: []corev1.Container{
						{
							Name:  "scanner",
							Image: GetImage(),
							Args: []string{
								"--all-pods",
								"-component-filter", Components,
								"-namespace-filter", targetNamespace,
								"-json-file", "/tmp/scan-results.json",
								"-j", "4", // Use 4 concurrent threads
							},
							SecurityContext: &corev1.SecurityContext{
								Privileged: utils.GetPtr(true),
							},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "results",
									MountPath: "/tmp",
									ReadOnly:  false,
								},
							},
						},
					},
					Containers: []corev1.Container{
						{
							Name:    "results",
							Image:   GetImage(),
							Command: []string{"cat"},
							Args: []string{
								"/tmp/scan-results.json",
							},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "results",
									MountPath: "/tmp",
									ReadOnly:  true,
								},
							},
						},
					},
					Volumes: []corev1.Volume{
						{
							Name: "results",
							VolumeSource: corev1.VolumeSource{
								EmptyDir: &corev1.EmptyDirVolumeSource{},
							},
						},
					},
				},
			},
		},
	}

	// Create the Job with bounded retry logic for AlreadyExists errors
	const (
		maxRetries           = 5
		retryInterval        = 2 * time.Second
		deletionTimeout      = 30 * time.Second
		deletionPollInterval = 1 * time.Second
	)

	deleteFn := func() error {
		deletePolicy := metav1.DeletePropagationForeground
		return s.KubeClient.BatchV1().Jobs(scannerNamespace).Delete(context.TODO(), job.Name, metav1.DeleteOptions{
			PropagationPolicy: &deletePolicy,
		})
	}

	var lastErr error
	for attempt := 0; attempt < maxRetries; attempt++ {
		createdJob, err := s.KubeClient.BatchV1().Jobs(scannerNamespace).Create(context.TODO(), job, metav1.CreateOptions{})
		if err == nil {
			// Success - add cleanup and return
			s.addCleanup(deleteFn)
			return createdJob, nil
		}

		if !apierrors.IsAlreadyExists(err) {
			// Non-retryable error
			return nil, fmt.Errorf("failed to create TLS Scanner Job: %w", err)
		}

		// Job already exists - delete it and wait for it to disappear
		lastErr = err
		clolog.V(2).Info("Job already exists, deleting and retrying", "job", job.Name, "attempt", attempt+1, "maxRetries", maxRetries)

		if err := deleteFn(); err != nil && !apierrors.IsNotFound(err) {
			clolog.V(2).Error(err, "failed to delete existing job", "job", job.Name)
		}

		// Wait for the job to be fully deleted before retrying
		ctx, cancel := context.WithTimeout(context.TODO(), deletionTimeout)
		pollErr := wait.PollUntilContextTimeout(ctx, deletionPollInterval, deletionTimeout, true, func(pollCtx context.Context) (bool, error) {
			_, getErr := s.KubeClient.BatchV1().Jobs(scannerNamespace).Get(pollCtx, job.Name, metav1.GetOptions{})
			if apierrors.IsNotFound(getErr) {
				return true, nil
			}
			return false, nil
		})
		cancel()

		if pollErr != nil {
			clolog.V(2).Info("Job deletion did not complete within timeout", "job", job.Name, "timeout", deletionTimeout)
		}

		// Backoff before next retry (except on last attempt)
		if attempt < maxRetries-1 {
			time.Sleep(retryInterval)
		}
	}

	return nil, fmt.Errorf("failed to create TLS Scanner Job after %d attempts: %w", maxRetries, lastErr)

}

// WaitForCompletion waits for the TLS Scanner Job to complete
func (s *Scanner) WaitForCompletion(job *batchv1.Job, timeout time.Duration) error {
	clolog.Info("Waiting for TLS Scanner Job to complete", "job", job.Name, "namespace", job.Namespace)

	return wait.PollUntilContextTimeout(context.TODO(), 10*time.Second, timeout, true, func(ctx context.Context) (bool, error) {
		currentJob, err := s.KubeClient.BatchV1().Jobs(job.Namespace).Get(ctx, job.Name, metav1.GetOptions{})
		if err != nil {
			clolog.V(0).Error(err, "error polling TLS Scanner Job", "job", job.Name)
			return false, nil
		}

		// Check if job has completed successfully
		if currentJob.Status.Succeeded > 0 {
			clolog.Info("TLS Scanner Job completed successfully", "job", job.Name)
			return true, nil
		}

		// Check if job has failed
		if currentJob.Status.Failed > 0 {
			clolog.Error(nil, "TLS Scanner Job failed", "job", job.Name)
			return false, fmt.Errorf("TLS Scanner Job failed")
		}

		clolog.V(3).Info("TLS Scanner Job still running", "job", job.Name, "active", currentJob.Status.Active)
		return false, nil
	})
}

// GetResults retrieves and parses the scan results from the TLS Scanner Job logs
func (s *Scanner) GetResults(job *batchv1.Job) ([]ScanResult, error) {
	clolog.Info("Retrieving TLS Scanner results", "job", job.Name, "namespace", job.Namespace)

	// Get the pod for the job
	pods, err := s.KubeClient.CoreV1().Pods(job.Namespace).List(context.TODO(), metav1.ListOptions{
		LabelSelector: fmt.Sprintf("job-name=%s", job.Name),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list pods for job: %w", err)
	}
	if len(pods.Items) == 0 {
		return nil, fmt.Errorf("no pods found for job %s", job.Name)
	}

	pod := pods.Items[0]

	// Get logs from the pod
	logs, err := s.KubeClient.CoreV1().Pods(pod.Namespace).GetLogs(pod.Name, &corev1.PodLogOptions{
		Container: "results",
	}).Stream(context.TODO())
	if err != nil {
		return nil, fmt.Errorf("failed to get logs from pod: %w", err)
	}
	defer func() { _ = logs.Close() }()
	clolog.V(3).Info("Results", "logs", logs)

	// Parse from logs
	results, err := parseResults(logs)
	if err != nil {
		return nil, fmt.Errorf("failed to parse TLS scan results: %w", err)
	}

	clolog.Info("Retrieved TLS Scanner results", "count", len(results))
	return results, nil
}

// parseResults parses the JSON output from TLS Scanner
func parseResults(r io.Reader) ([]ScanResult, error) {
	// Read all content from the reader
	content, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("failed to read scanner output: %w", err)
	}
	clolog.V(3).Info("Scanner output", "content", string(content))

	// Parse JSON into ScanOutput struct
	var scanOutput ScanOutput
	if err := json.Unmarshal(content, &scanOutput); err != nil {
		return nil, fmt.Errorf("failed to unmarshal scanner output: %w", err)
	}

	// Convert ScanOutput to []ScanResult format
	var results []ScanResult
	for _, ipResult := range scanOutput.IPResults {
		for _, portResult := range ipResult.PortResults {
			result := ScanResult{
				Component:    ipResult.OpenShiftComponent.Component,
				IP:           ipResult.IP,
				Port:         fmt.Sprintf("%d", portResult.Port),
				Protocol:     portResult.Protocol,
				Service:      portResult.Service,
				Status:       portResult.Status,
				TLSReadiness: portResult.TLSReadiness,
			}

			// Extract TLS versions from TLSReadiness if available
			if portResult.TLSReadiness != nil {
				if portResult.TLSReadiness.TLS13Offered {
					result.TLSVersions = append(result.TLSVersions, "TLS 1.3")
				}
				if portResult.TLSReadiness.TLS12Only || (!portResult.TLSReadiness.TLS13Offered) {
					result.TLSVersions = append(result.TLSVersions, "TLS 1.2")
				}
			}

			results = append(results, result)
		}
	}

	clolog.V(3).Info("Parsed TLS scanner results", "ipResults", len(scanOutput.IPResults), "portResults", len(results))
	return results, nil
}

// ValidateCompliance validates that TLS scan results comply with the cluster TLS profile
func ValidateCompliance(results []ScanResult, profileSpec configv1.TLSProfileSpec) error {
	clolog.Info("Validating TLS compliance", "results", len(results), "profile", profileSpec.MinTLSVersion)

	var failures []string
	expectedMinTLS := internaltls.MinTLSVersion(profileSpec)

	for _, result := range results {
		// Skip non-TLS endpoints
		if result.Status == "NO_TLS" || result.Status == "LOCALHOST_ONLY" {
			clolog.V(2).Info("Skipping non-TLS endpoint", "pod", result.Pod, "port", result.Port, "status", result.Status)
			continue
		}

		// Check for error statuses
		if result.Status == "ERROR" || result.Status == "TIMEOUT" {
			failures = append(failures, fmt.Sprintf("Pod %s/%s port %s: scan failed with status %s",
				result.Namespace, result.Pod, result.Port, result.Status))
			continue
		}

		// For MTLS_REQUIRED, we can't validate the TLS version, but it's not necessarily a failure
		if result.Status == "MTLS_REQUIRED" {
			clolog.V(2).Info("Endpoint requires mutual TLS", "pod", result.Pod, "port", result.Port)
			continue
		}

		// For successful scans, validate TLS version
		if result.Status == "OK" {
			if !containsTLSVersion(result.TLSVersions, expectedMinTLS) {
				failures = append(failures, fmt.Sprintf("Pod %s/%s port %s: expected min TLS version %s, got versions: %s",
					result.Namespace, result.Pod, result.Port, expectedMinTLS, result.TLSVersions))
			}

			if result.TLSReadiness == nil {
				failures = append(failures, fmt.Sprintf("missing TLSReadiness for %s", result.Component))
			} else if !result.TLSReadiness.PQCCapable {
				failures = append(failures, fmt.Sprintf("%s is not 'pqc_capable': %s", result.Component, result.TLSReadiness.Notes))
			}

			// Log successful validation
			clolog.V(2).Info("TLS endpoint validated", "pod", result.Pod, "port", result.Port,
				"tlsVersions", result.TLSVersions, "status", result.Status)
		}
	}

	if len(failures) > 0 {
		return fmt.Errorf("TLS compliance validation failed:\n%s", strings.Join(failures, "\n"))
	}

	clolog.Info("TLS compliance validation passed", "results", len(results))
	return nil
}

// containsTLSVersion checks if all TLS versions meet the minimum required version
func containsTLSVersion(tlsVersions []string, minVersion string) bool {
	// If no TLS versions detected, fail validation
	if len(tlsVersions) == 0 {
		return false
	}

	// Parse the minimum version
	minVersionNum := parseTLSVersion(minVersion)
	if minVersionNum == 0 {
		// If we can't parse the min version, assume it's valid
		return true
	}

	// Check that ALL supported versions meet the minimum
	foundVersion := false
	for _, v := range tlsVersions {
		v = strings.TrimSpace(v)
		vNum := parseTLSVersion(v)
		if vNum == 0 {
			// Skip unparseable versions
			continue
		}
		foundVersion = true
		if vNum < minVersionNum {
			// Found a version below the minimum - fail validation
			return false
		}
	}

	// Return true only if we found at least one version and none were below minimum
	return foundVersion
}

// parseTLSVersion converts TLS version string to a comparable number
func parseTLSVersion(version string) int {
	version = strings.TrimSpace(strings.ToLower(version))
	switch {
	case strings.Contains(version, "1.0"):
		return 10
	case strings.Contains(version, "1.1"):
		return 11
	case strings.Contains(version, "1.2"):
		return 12
	case strings.Contains(version, "1.3"):
		return 13
	default:
		return 0
	}
}

// addCleanup adds a cleanup function to the list
func (s *Scanner) addCleanup(fn func() error) {
	if s.CleanupFns != nil {
		*s.CleanupFns = append(*s.CleanupFns, fn)
	}
}
