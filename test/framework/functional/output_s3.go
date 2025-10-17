package functional

/*
import (
	"bufio"
	"context"
	"crypto/tls"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	"io"
	"net/http"
	"strings"

	log "github.com/ViaQ/logerr/v2/log/static"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	openshiftv1 "github.com/openshift/api/route/v1"
	"github.com/openshift/cluster-logging-operator/internal/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
)

const (
	S3Secret      = "s3-secret"
	awsRegion     = "us-east-2"
	minioImage    = "quay.io/minio/minio:RELEASE.2025-07-23T15-54-02Z"
	minioPort     = 9000
	minioCmd      = "server"
	minioDataPath = "/data"
)

var (
	s3Client *s3.Client
)

func (f *CollectorFunctionalFramework) AddS3Output(b *runtime.PodBuilder, spec obs.OutputSpec) error {
	if err := f.createMinIOService(); err != nil {
		return err
	}

	if err := f.createMinIORoute(); err != nil {
		return err
	}

	// initialize the client
	s3Client = s3.NewFromConfig(
		aws.Config{
			Region: "us-east-2",
			Credentials: aws.CredentialsProviderFunc(func(context.Context) (aws.Credentials, error) {
				return aws.Credentials{
					AccessKeyID:     AwsAccessKeyID,
					SecretAccessKey: AwsSecretAccessKey,
				}, nil
			}),
			HTTPClient: &http.Client{
				Transport: &http.Transport{
					TLSClientConfig: &tls.Config{
						InsecureSkipVerify: true, //nolint:gosec
					},
				},
			},
			EndpointResolver: aws.EndpointResolverFunc(func(service, region string) (aws.Endpoint, error) {
				return aws.Endpoint{
					PartitionID:   "aws",
					URL:           "https://" + route.Spec.Host,
					SigningRegion: "us-east-2",
					//Source:        aws.EndpointSourceCustom,
				}, nil
			}),
		},
	)

	// add the minIO container
	b.AddContainer("minio", minioImage).
		WithCmdArgs([]string{minioCmd, minioDataPath}).
		AddEnvVar("MINIO_ROOT_USER", AwsAccessKeyID).
		AddEnvVar("MINIO_ROOT_PASSWORD", AwsSecretAccessKey).
		AddContainerPort("minio-port", minioPort).
		End()

	return nil
}

func (f *CollectorFunctionalFramework) createMinIOService() error {
	service = runtime.NewService(f.Namespace, "minio-server")
	runtime.NewServiceBuilder(service).
		AddServicePort(minioPort, minioPort).
		WithSelector(map[string]string{"testname": "functional"})

	if err := f.Test.Client.Create(service); err != nil {
		return fmt.Errorf("unable to create minio service: %v", err)
	}
	return nil
}

func (f *CollectorFunctionalFramework) createMinIORoute() error {
	log.V(2).Info("Creating route for minio-server")
	// Re-using the global route variable defined in the functional package
	route = runtime.NewRoute(f.Namespace, "minio-server", "minio-server", fmt.Sprintf("%d", minioPort))
	route.Spec.TLS = &openshiftv1.TLSConfig{
		Termination:                   openshiftv1.TLSTerminationPassthrough,
		InsecureEdgeTerminationPolicy: openshiftv1.InsecureEdgeTerminationPolicyNone,
	}
	if err := f.Test.Client.Create(route); err != nil {
		return fmt.Errorf("unable to create minio route: %v", err)
	}
	return nil
}

// SetupS3Bucket wait for minIO to be ready, then create the bucket
func (f *CollectorFunctionalFramework) SetupS3Bucket(bucketName string) error {
	log.V(2).Info("waiting for minIO mock to be ready")
	err := wait.PollUntilContextTimeout(context.TODO(), defaultRetryInterval, f.GetMaxReadDuration(), true, func(ctx context.Context) (done bool, err error) {
		// keep trying until minIO responds
		_, err = s3Client.CreateBucket(ctx, &s3.CreateBucketInput{
			Bucket: aws.String(bucketName),
			CreateBucketConfiguration: &types.CreateBucketConfiguration{
				LocationConstraint: types.BucketLocationConstraintUsEast2,
			},
		})
		if err != nil {
			log.V(3).Error(err, "Bucket creation failed, retrying...")
			return false, nil
		}
		log.V(2).Info("S3 bucket created successfully", "bucket", bucketName)
		return true, nil
	})
	return err
}

// ReadLogsFromS3 read bucket files under the key prefix and concat all log lines into a single slice
func (f *CollectorFunctionalFramework) ReadLogsFromS3(bucketName, keyPrefix string) ([]string, error) {
	var allMessages []string
	var listOutput *s3.ListObjectsV2Output

	// wait until files are visible
	log.V(2).Info("waiting for logs in s3 bucket", "bucket", bucketName, "keyPrefix", keyPrefix)
	err := wait.PollUntilContextTimeout(context.TODO(), defaultRetryInterval, f.GetMaxReadDuration(), true, func(ctx context.Context) (done bool, err error) {
		listOutput, err = s3Client.ListObjectsV2(ctx, &s3.ListObjectsV2Input{
			Bucket: aws.String(bucketName),
			Prefix: aws.String(keyPrefix),
		})
		if err != nil {
			return false, fmt.Errorf("failed to list objects in s3: %w", err)
		}
		if listOutput.Contents == nil || len(listOutput.Contents) == 0 {
			log.V(3).Info("still no s3 objects found yet...")
			return false, nil
		}
		return true, nil
	})
	if err != nil {
		return nil, fmt.Errorf("timed out waiting for logs to appear in s3 bucket: %w", err)
	}

	// get all file contents
	for _, object := range listOutput.Contents {
		log.V(3).Info("getting s3 file content", "key", *object.Key)

		// getter
		getFileOutput, err := s3Client.GetObject(context.TODO(), &s3.GetObjectInput{
			Bucket: aws.String(bucketName),
			Key:    object.Key,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to get object %s: %w", *object.Key, err)
		}
		defer getFileOutput.Body.Close()

		// read and concat
		reader := bufio.NewReader(getFileOutput.Body)
		for {
			line, err := reader.ReadString('\n')
			if err != nil && err != io.EOF {
				return nil, fmt.Errorf("error reading file: %w", err)
			}
			// need to sanitize
			trimmedLine := strings.TrimSpace(line)
			if trimmedLine != "" {
				allMessages = append(allMessages, trimmedLine)
			}
			if err == io.EOF {
				break
			}
		}
	}

	return allMessages, nil
}
*/
