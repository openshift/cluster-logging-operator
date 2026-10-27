package lokistack

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"syscall"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// defaultMaxConcurrentSpecs caps how many lokistack specs may run at once. Each
// spec stands up a full LokiStack (microservices), a collector DaemonSet, minio
// and a log generator, so unbounded concurrency starves the shared CI cluster
// and logs fail to ingest within the query timeout. The cap is deliberately
// lower than ginkgo's --procs: `ginkgo -p` runs the workers as separate OS
// processes, so this is enforced with a cross-process flock semaphore rather
// than lowering --procs in hack/test-e2e-from-ci-bundle.sh.
const defaultMaxConcurrentSpecs = 2

// maxConcurrentSpecsEnv overrides defaultMaxConcurrentSpecs when set, letting CI
// tune throttling for a given cluster size without a recompile.
const maxConcurrentSpecsEnv = "LOKI_E2E_MAX_CONCURRENT_SPECS"

// slotPollInterval is how often a blocked spec retries for a free slot.
const slotPollInterval = 500 * time.Millisecond

// maxConcurrentSpecs returns the configured concurrency cap (>= 1).
func maxConcurrentSpecs() int {
	if v := os.Getenv(maxConcurrentSpecsEnv); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n >= 1 {
			return n
		}
	}
	return defaultMaxConcurrentSpecs
}

// acquireSpecSlot blocks until one of maxConcurrentSpecs() cross-process slots
// is free and registers the release so the slot is held until this spec has
// finished tearing down (DeferCleanup runs after the spec's AfterEach nodes).
//
// The ginkgo parallel workers share this host's filesystem, so an exclusive,
// non-blocking flock on a per-slot lock file is a reliable counting semaphore:
// the lock is tied to the open file description and the kernel drops it if a
// worker dies mid-spec, so a crashed spec can never leak a slot.
func acquireSpecSlot() {
	n := maxConcurrentSpecs()
	dir := filepath.Join(os.TempDir(), "loki-e2e-slots")
	Expect(os.MkdirAll(dir, 0o755)).To(Succeed(), "exp. to create lokistack concurrency slot dir")

	for {
		for i := 0; i < n; i++ {
			f, err := os.OpenFile(filepath.Join(dir, fmt.Sprintf("slot-%d.lock", i)), os.O_CREATE|os.O_RDWR, 0o644)
			if err != nil {
				continue
			}
			if err := syscall.Flock(int(f.Fd()), syscall.LOCK_EX|syscall.LOCK_NB); err == nil {
				DeferCleanup(releaseSpecSlot, f)
				return
			}
			_ = f.Close()
		}
		time.Sleep(slotPollInterval)
	}
}

// releaseSpecSlot drops the flock and closes the slot file, freeing it for a
// waiting spec.
func releaseSpecSlot(f *os.File) {
	if f == nil {
		return
	}
	_ = syscall.Flock(int(f.Fd()), syscall.LOCK_UN)
	_ = f.Close()
}
