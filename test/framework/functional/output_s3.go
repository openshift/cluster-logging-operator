package functional

import (
	"bytes"
	"compress/gzip"
	"compress/zlib"
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	log "github.com/ViaQ/logerr/v2/log/static"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/klauspost/compress/snappy"
	"github.com/klauspost/compress/zstd"
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	"github.com/openshift/cluster-logging-operator/internal/runtime"
	"github.com/openshift/cluster-logging-operator/internal/utils/toml"
	"github.com/openshift/cluster-logging-operator/test/helpers/oc"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/util/wait"
)

type s3State struct {
	client           *s3.Client
	portForwarder    *oc.PortForwarder
	minioService     *corev1.Service
	minioInitialized bool
}

var s3States = make(map[*CollectorFunctionalFramework]*s3State)

const (
	S3Secret      = "s3-secret"
	awsRegion     = "us-east-2"
	minioImage    = "quay.io/minio/minio:RELEASE.2025-07-23T15-54-02Z"
	MinioPort     = 9000
	minioCmd      = "server"
	minioDataPath = "/data"
)

func (f *CollectorFunctionalFramework) AddS3Output(b *runtime.PodBuilder, spec obs.OutputSpec) error {
	if _, ok := s3States[f]; !ok {
		s3States[f] = &s3State{}
	}

	state := s3States[f]

	if !state.minioInitialized {
		if err := f.createMinIOService(state); err != nil {
			return err
		}

		b.AddContainer("minio", minioImage).
			WithCmdArgs([]string{minioCmd, minioDataPath}).
			AddEnvVar("MINIO_ROOT_USER", AwsAccessKeyID).
			AddEnvVar("MINIO_ROOT_PASSWORD", AwsSecretAccessKey).
			AddContainerPort("minio-port", MinioPort).
			End()

		state.minioInitialized = true
	}
	return nil
}

func (f *CollectorFunctionalFramework) createMinIOService(state *s3State) error {
	state.minioService = runtime.NewService(f.Namespace, "minio-server")
	runtime.NewServiceBuilder(state.minioService).
		AddServicePort(MinioPort, MinioPort).
		WithSelector(map[string]string{"testname": "functional"})

	if err := f.Test.Create(state.minioService); err != nil {
		if !apierrors.IsAlreadyExists(err) {
			return fmt.Errorf("unable to create minio service: %v", err)
		}
		if err := f.Test.Get(state.minioService); err != nil {
			return fmt.Errorf("unable to retrieve existing minio service: %v", err)
		}
	}
	return nil
}

// SetS3BatchTimeout parses Vector TOML config, sets batch timeout, and returns the modified config.
func SetS3BatchTimeout(config string, timeoutSecs int) (string, error) {
	return toml.SetValue(config, []string{"sinks", "output_s3", "batch", "timeout_secs"}, int64(timeoutSecs))
}

// SetupS3Bucket sets up port-forward to minIO, then waits for it to be ready and creates the bucket
func (f *CollectorFunctionalFramework) SetupS3Bucket(bucketName string) error {
	log.V(2).Info("Setting up port-forward to minIO service")

	state, ok := s3States[f]
	if !ok || state == nil || !state.minioInitialized {
		return fmt.Errorf("AddS3Output must be called before SetupS3Bucket")
	}

	// Set up port-forwarding to the minIO service
	portForwarder, err := oc.SetupServicePortForwarder(f.Namespace, "minio-server", MinioPort)
	if err != nil {
		return fmt.Errorf("unable to setup port-forward to minIO: %v", err)
	}
	state.portForwarder = portForwarder
	log.V(2).Info("Port-forward to minIO established", "localPort", portForwarder.LocalPort())

	// Initialize S3 client with the forwarded port
	state.client = s3.NewFromConfig(
		aws.Config{
			Region: awsRegion,
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
		},
		func(o *s3.Options) {
			o.BaseEndpoint = aws.String("http://127.0.0.1:" + portForwarder.LocalPort())
			o.UsePathStyle = true
		},
	)

	log.V(2).Info("waiting for minIO mock to be ready")
	err = wait.PollUntilContextTimeout(context.TODO(), defaultRetryInterval, f.GetMaxReadDuration(), true, func(ctx context.Context) (done bool, err error) {
		_, err = state.client.CreateBucket(ctx, &s3.CreateBucketInput{
			Bucket: aws.String(bucketName),
			CreateBucketConfiguration: &types.CreateBucketConfiguration{
				LocationConstraint: types.BucketLocationConstraintUsEast2,
			},
		})
		if err != nil {
			// Bucket may already exist from a previous call — treat as success
			var baoby *types.BucketAlreadyOwnedByYou
			var bae *types.BucketAlreadyExists
			if errors.As(err, &baoby) || errors.As(err, &bae) {
				log.V(2).Info("S3 bucket already exists", "bucket", bucketName)
				return true, nil
			}
			log.V(3).Error(err, "Bucket creation failed, retrying...")
			return false, nil
		}
		log.V(2).Info("S3 bucket created successfully", "bucket", bucketName)
		return true, nil
	})
	return err
}

// ReadLogsFromS3 reads bucket files under the key prefix and returns individual log entries
func (f *CollectorFunctionalFramework) ReadLogsFromS3(bucketName, keyPrefix string) ([]string, error) {
	state, ok := s3States[f]
	if !ok || state == nil || state.client == nil {
		return nil, fmt.Errorf("S3 not initialized; call SetupS3Bucket before reading logs")
	}

	var allMessages []string
	var listOutput *s3.ListObjectsV2Output

	log.V(2).Info("waiting for logs in s3 bucket", "bucket", bucketName, "keyPrefix", keyPrefix)
	err := wait.PollUntilContextTimeout(context.TODO(), defaultRetryInterval, f.GetMaxReadDuration(), true, func(ctx context.Context) (done bool, err error) {
		listOutput, err = state.client.ListObjectsV2(ctx, &s3.ListObjectsV2Input{
			Bucket: aws.String(bucketName),
			Prefix: aws.String(keyPrefix),
		})
		if err != nil {
			log.V(3).Error(err, "Failed to list objects in s3, retrying...")
			return false, nil
		}
		if len(listOutput.Contents) == 0 {
			log.V(3).Info("still no s3 objects found yet...")
			return false, nil
		}
		return true, nil
	})
	if err != nil {
		return nil, fmt.Errorf("timed out waiting for logs to appear in s3 bucket: %w", err)
	}

	for _, object := range listOutput.Contents {
		log.V(3).Info("getting s3 file content", "key", *object.Key)

		// Use a bounded context for GetObject and body read to prevent stalling
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		getFileOutput, err := state.client.GetObject(ctx, &s3.GetObjectInput{
			Bucket: aws.String(bucketName),
			Key:    object.Key,
		})
		if err != nil {
			cancel()
			return nil, fmt.Errorf("failed to get object %s: %w", *object.Key, err)
		}

		body, err := io.ReadAll(getFileOutput.Body)
		_ = getFileOutput.Body.Close()
		if err != nil {
			cancel()
			return nil, fmt.Errorf("failed to read object %s: %w", *object.Key, err)
		}
		cancel()

		// Try to decompress if the content is not valid JSON (compressed data)
		content := strings.TrimSpace(string(body))
		if content != "" && !strings.HasPrefix(content, "[") && !strings.HasPrefix(content, "{") {
			if decompressed, err := tryDecompress(body); err == nil {
				content = strings.TrimSpace(string(decompressed))
			}
		}

		if content == "" {
			continue
		}

		// Vector writes S3 objects as JSON arrays: [{log1},{log2},...]
		if strings.HasPrefix(content, "[") {
			var entries []json.RawMessage
			if err := json.Unmarshal([]byte(content), &entries); err != nil {
				allMessages = append(allMessages, content)
				continue
			}
			for _, entry := range entries {
				allMessages = append(allMessages, string(entry))
			}
		} else {
			for _, line := range strings.Split(content, "\n") {
				if trimmed := strings.TrimSpace(line); trimmed != "" {
					allMessages = append(allMessages, trimmed)
				}
			}
		}
	}

	return allMessages, nil
}

// CleanupS3PortForward stops the port-forward if it's running
func (f *CollectorFunctionalFramework) CleanupS3PortForward() {
	state, ok := s3States[f]
	if !ok || state == nil {
		return
	}
	defer delete(s3States, f)

	if state.portForwarder != nil {
		state.portForwarder.Stop()
	}
}

// tryDecompress attempts to decompress data using common compression formats (gzip, zlib, snappy, zstd)
func tryDecompress(data []byte) ([]byte, error) {
	if gz, err := gzip.NewReader(bytes.NewReader(data)); err == nil {
		if decompressed, err := io.ReadAll(gz); err == nil {
			_ = gz.Close()
			return decompressed, nil
		}
		_ = gz.Close()
	}

	if zr, err := zlib.NewReader(bytes.NewReader(data)); err == nil {
		if decompressed, err := io.ReadAll(zr); err == nil {
			_ = zr.Close()
			return decompressed, nil
		}
		_ = zr.Close()
	}

	if decompressed, err := snappy.Decode(nil, data); err == nil {
		return decompressed, nil
	}

	if decoder, err := zstd.NewReader(bytes.NewReader(data)); err == nil {
		if decompressed, err := io.ReadAll(decoder); err == nil {
			decoder.Close()
			return decompressed, nil
		}
		decoder.Close()
	}

	return nil, fmt.Errorf("unable to decompress data")
}
