package vector

import (
	"strconv"

	corev1 "k8s.io/api/core/v1"
)

// vectorThreads derives the value for the VECTOR_THREADS environment variable from the
// collector's CPU limit. Vector sizes its worker-thread pool from the host's available
// parallelism, which is not aware of the container's CFS quota. On nodes with many more
// cores than the CPU limit this oversubscribes the quota and causes heavy CPU throttling.
// Pinning the worker count to the CPU limit avoids that.
//
// The count is floor(limit cores) with a minimum of 1, so the number of workers never
// exceeds the quota. When no CPU limit is set (unlimited CPU) no value is returned, since
// there is no quota to throttle against and capping the workers would only reduce throughput.
func vectorThreads(resources corev1.ResourceRequirements) (string, bool) {
	cpu, ok := resources.Limits[corev1.ResourceCPU]
	if !ok || cpu.IsZero() {
		return "", false
	}
	threads := cpu.MilliValue() / 1000
	if threads < 1 {
		threads = 1
	}
	return strconv.FormatInt(threads, 10), true
}
