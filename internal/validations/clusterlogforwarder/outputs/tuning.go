package outputs

import (
	"context"
	"fmt"
	loggingv1 "github.com/openshift/cluster-logging-operator/api/logging/v1"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/output/common"
	"github.com/openshift/cluster-logging-operator/internal/utils/sets"
	"k8s.io/apimachinery/pkg/api/resource"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sync"
)

const (
	compressionNotSupportedForType = "compression is not supported for the output type"

	maxRetryDurationNotSupportedForType = "maxRetryDuration is not supported for the output type"
	minRetryDurationNotSupportedForType = "minRetryDuration is not supported for the output type"
	maxWriteNotSupportedForType         = "maxWrite is not supported for the output type"

	nodeDiskLimitExceeded = "The amount of allocated buffer exceeds the allowed node limit for all ClusterLogForwarders. The delivery mode for one or more outputs for any ClusterLogForwarder must be adjusted."

	//nodeDiskLimitPercent is the maximum disk usage by all forwarder deployments
	nodeDiskLimitPercent = 15

	//nodeDiskSize is the assumed size of the disk of an OCP cluster
	nodeDiskSize = "120G"
)

var (
	compressionOutputMap = map[string]sets.String{
		"gzip": *sets.NewString(
			loggingv1.OutputTypeCloudwatch,
			loggingv1.OutputTypeElasticsearch,
			loggingv1.OutputTypeHttp,
			loggingv1.OutputTypeLoki,
			loggingv1.OutputTypeSplunk),
		"snappy": *sets.NewString(
			loggingv1.OutputTypeCloudwatch,
			loggingv1.OutputTypeHttp,
			loggingv1.OutputTypeKafka,
			loggingv1.OutputTypeLoki),
		"zlib": *sets.NewString(
			loggingv1.OutputTypeCloudwatch,
			loggingv1.OutputTypeElasticsearch,
			loggingv1.OutputTypeHttp),
		"zstd": *sets.NewString(
			loggingv1.OutputTypeCloudwatch,
			loggingv1.OutputTypeHttp,
			loggingv1.OutputTypeKafka,
		),
		"lz4": *sets.NewString(loggingv1.OutputTypeKafka),
	}
	unsupportedRequest = sets.NewString(loggingv1.OutputTypeSyslog, loggingv1.OutputTypeKafka)
	unsupportedBatch   = sets.NewString(loggingv1.OutputTypeSyslog)

	nodeDiskLimitBytes int64

	lock sync.Mutex
)

func init() {
	diskSizeBytes := resource.MustParse(nodeDiskSize)
	nodeDiskLimitBytes = diskSizeBytes.Value() * nodeDiskLimitPercent / 100
}

func VerifyTuning(spec loggingv1.OutputSpec) (bool, string) {
	if spec.Tuning == nil {
		return true, ""
	}

	// compression
	if msg := verifyCompression(spec.Type, spec.Tuning.Compression); msg != "" {
		return false, msg
	}

	// batch
	if unsupportedBatch.Has(spec.Type) && spec.Tuning.MaxWrite != nil && !spec.Tuning.MaxWrite.IsZero() {
		return false, maxWriteNotSupportedForType
	}

	//MaxRetryDuration
	if unsupportedRequest.Has(spec.Type) && spec.Tuning.MaxRetryDuration != nil && spec.Tuning.MaxRetryDuration.Seconds() > 0 {
		return false, maxRetryDurationNotSupportedForType
	}
	//MaxRetryDuration
	if unsupportedRequest.Has(spec.Type) && spec.Tuning.MinRetryDuration != nil && spec.Tuning.MinRetryDuration.Seconds() > 0 {
		return false, minRetryDurationNotSupportedForType
	}

	return true, ""
}

func verifyCompression(outputType string, compression string) string {
	if compression == "" || compression == "none" {
		return ""
	}

	if outputs, ok := compressionOutputMap[compression]; ok {
		if !outputs.Has(outputType) {
			return compressionNotSupportedForType
		}
		// Compression not in map
	} else {
		return compressionNotSupportedForType
	}

	return ""
}

// ValidateCumulativeDiskBuffer validates outputs using disk buffers for all ClusterLogForwarders do not exceed the allowed threshold
func ValidateCumulativeDiskBuffer(clf loggingv1.ClusterLogForwarder, k8sClient client.Client, extras map[string]bool) (error, *loggingv1.ClusterLogForwarderStatus) {
	lock.Lock()
	defer lock.Unlock()
	all := &loggingv1.ClusterLogForwarderList{}
	if err := k8sClient.List(context.TODO(), all); err != nil {
		return err, nil
	}
	allocatedBufferBytes := gatherAllocatedBuffer(clf)
	for _, anotherCLF := range all.Items {
		if anotherCLF.Namespace != clf.Namespace && anotherCLF.Name != clf.Name {
			allocatedBufferBytes += gatherAllocatedBuffer(anotherCLF)
		}
	}
	if allocatedBufferBytes > nodeDiskLimitBytes {
		return fmt.Errorf(nodeDiskLimitExceeded), nil
	}
	return nil, nil
}

func gatherAllocatedBuffer(clf loggingv1.ClusterLogForwarder) int64 {
	allocatedBufferBytes := int64(0)
	for _, o := range clf.Spec.Outputs {
		if o.Tuning != nil && o.Tuning.Delivery == loggingv1.OutputDeliveryModeAtLeastOnce {
			allocatedBufferBytes += common.BufferMinSizeBytes
		}
	}
	return allocatedBufferBytes
}
