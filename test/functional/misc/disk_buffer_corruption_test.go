package misc

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	"github.com/openshift/cluster-logging-operator/internal/collector/vector"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	"github.com/openshift/cluster-logging-operator/internal/runtime"
	"github.com/openshift/cluster-logging-operator/internal/utils"
	"github.com/openshift/cluster-logging-operator/test/framework/functional"
	"github.com/openshift/cluster-logging-operator/test/helpers/oc"
	"github.com/openshift/cluster-logging-operator/test/helpers/types"
	testruntime "github.com/openshift/cluster-logging-operator/test/runtime/observability"
	"k8s.io/apimachinery/pkg/util/wait"
)

// LOG-9386: Vector should not CrashLoop when disk buffer contains corrupted records.
// The fix in vectordotdev/vector#25691 treats decode errors as bad reads during
// initialization seek, allowing Vector to skip corrupted records and start normally.
var _ = Describe("[Functional][Misc] Vector disk buffer corruption recovery", func() {

	var (
		framework *functional.CollectorFunctionalFramework
	)

	BeforeEach(func() {
		framework = functional.NewCollectorFunctionalFramework()
		testruntime.NewClusterLogForwarderBuilder(framework.Forwarder).
			FromInput(obs.InputTypeApplication).
			ToHttpOutput(func(output *obs.OutputSpec) {
				output.HTTP.Tuning = &obs.HTTPTuningSpec{
					BaseOutputTuningSpec: obs.BaseOutputTuningSpec{
						DeliveryMode: obs.DeliveryModeAtLeastOnce,
					},
				}
			})
	})

	AfterEach(func() {
		framework.Cleanup()
	})

	It("should recover from corrupted disk buffer and continue collecting logs", func() {
		dataPath := vector.GetDataPath(framework.Namespace, framework.Name)

		Expect(framework.DeployWithVisitors(
			append(framework.AddOutputContainersVisitors(),
				framework.AddToolsContainerVisitor("datadir", dataPath),
				func(b *runtime.PodBuilder) error {
					b.Pod.Spec.ShareProcessNamespace = utils.GetPtr(true)
					return nil
				}),
		)).To(BeNil())

		By("writing application logs and verifying delivery (baseline)")
		timestamp := "2024-11-04T18:13:59.061892+00:00"
		message := "disk-buffer-test-before-corruption"
		applicationLogLine := functional.NewFullCRIOLogMessage(timestamp, message)
		Expect(framework.WriteMessagesToApplicationLog(applicationLogLine, 10)).To(BeNil())

		result, err := framework.ReadFileFrom(string(obs.OutputTypeHTTP), functional.ApplicationLogFile)
		Expect(err).To(BeNil(), "Expected no errors reading baseline logs")
		Expect(result).ToNot(BeEmpty())
		raw := strings.Split(strings.TrimSpace(result), "\n")
		logs, err := types.ParseLogs(utils.ToJsonLogs(raw))
		Expect(err).To(BeNil(), "Expected no errors parsing baseline logs")
		Expect(logs[0].Message).To(Equal(message))

		By("freezing the HTTP receiver")
		receiverPid := findVectorPid(framework, string(obs.OutputTypeHTTP))
		out, err := framework.RunCommand(functional.ToolsContainerName, "sh", "-c", fmt.Sprintf("kill -STOP %s", receiverPid))
		Expect(err).To(BeNil(), "Failed to freeze HTTP receiver: %s", out)
		time.Sleep(5 * time.Second)

		By("writing messages in waves to force multiple buffer records")
		const waves = 4
		const msgsPerWave = 50
		for w := 0; w < waves; w++ {
			burstWriteApplicationLogs(framework, timestamp, w*msgsPerWave+1, (w+1)*msgsPerWave)
			if w < waves-1 {
				time.Sleep(3 * time.Second)
			}
		}

		By("waiting for the disk buffer file to appear and stabilize")
		datFilePath := waitForBufferDataFile(framework, dataPath)

		By("verifying the HTTP receiver is not receiving messages while frozen")
		sinkOutput, err := framework.RunCommand(string(obs.OutputTypeHTTP), "cat", functional.ApplicationLogFile)
		Expect(err).To(BeNil())
		Expect(sinkOutput).ToNot(ContainSubstring("buffer-fill-msg-"),
			"HTTP receiver should not have received burst messages while frozen")

		By("freezing the collector so the buffer file is stable during corruption")
		collectorPid := findVectorPid(framework, constants.CollectorName)
		_, err = framework.RunCommand(functional.ToolsContainerName, "sh", "-c", fmt.Sprintf("kill -STOP %s", collectorPid))
		Expect(err).To(BeNil(), "Failed to freeze collector")

		By("corrupting the disk buffer via the tools sidecar")
		corruptBufferFile(framework, datFilePath)

		By("killing the frozen collector so it restarts and reads the corrupted buffer")
		out, err = framework.RunCommand(functional.ToolsContainerName, "sh", "-c", fmt.Sprintf("kill -9 %s", collectorPid))
		Expect(err).To(BeNil(), "Failed to kill frozen collector: %s", out)
		waitForCollectorRestart(framework)

		By("waiting for Vector to start and detect corrupted records")
		waitForCollectorStartAndValidateCorruptionDetected(framework)

		By("verifying the collector is not crash-looping")
		assertNoCollectorCrashLoop(framework)

		By("resuming the HTTP receiver")
		out, err = framework.RunCommand(functional.ToolsContainerName, "sh", "-c", fmt.Sprintf("kill -CONT %s", receiverPid))
		Expect(err).To(BeNil(), "Failed to resume HTTP receiver: %s", out)

		By("verifying post-restart delivery")
		const postRestartCount = 5
		for i := 1; i <= postRestartCount; i++ {
			line := functional.NewFullCRIOLogMessage(timestamp, fmt.Sprintf("post-corruption-msg-%d", i))
			Expect(framework.WriteMessagesToApplicationLog(line, 1)).To(BeNil())
		}

		var sinkResult string
		err = wait.PollUntilContextTimeout(context.TODO(), 2*time.Second, 90*time.Second, true, func(ctx context.Context) (bool, error) {
			sinkResult, err = framework.RunCommand(string(obs.OutputTypeHTTP), "cat", functional.ApplicationLogFile)
			if err != nil {
				return false, nil
			}
			return strings.Contains(sinkResult, fmt.Sprintf("post-corruption-msg-%d", postRestartCount)), nil
		})
		Expect(err).To(BeNil(), "Post-restart messages did not arrive within timeout")
		for i := 1; i <= postRestartCount; i++ {
			Expect(sinkResult).To(ContainSubstring(fmt.Sprintf("post-corruption-msg-%d", i)),
				"Missing post-restart message %d of %d", i, postRestartCount)
		}
	})
})

// burstWriteApplicationLogs writes log lines numbered from..to directly to the
// container log file without per-message sleeps.
func burstWriteApplicationLogs(framework *functional.CollectorFunctionalFramework, timestamp string, from, to int) {
	logFile := fmt.Sprintf("/var/log/pods/%s_%s_%s/%s/0.log",
		framework.Namespace, framework.Name, framework.Pod.UID, constants.CollectorName)
	burstCmd := fmt.Sprintf(
		`for n in $(seq %d %d); do echo "%s stdout F buffer-fill-msg-$n" >> %s; done`,
		from, to, timestamp, logFile)
	_, err := framework.RunCommand(constants.CollectorName, "bash", "-c", burstCmd)
	Expect(err).To(BeNil(), "Failed to write burst messages")
}

// assertNoCollectorCrashLoop waits briefly then verifies the collector has
// exactly 1 restart and is not in CrashLoopBackOff.
func assertNoCollectorCrashLoop(framework *functional.CollectorFunctionalFramework) {
	time.Sleep(10 * time.Second)

	countStr, err := oc.Get().
		WithNamespace(framework.Namespace).
		Resource("pod", framework.Name).
		OutputJsonpath("{.status.containerStatuses[?(@.name==\"" + constants.CollectorName + "\")].restartCount}").
		Run()
	Expect(err).To(BeNil())
	Expect(strings.TrimSpace(countStr)).To(Equal("1"), "Expected exactly 1 restart")

	stateJson, err := oc.Get().
		WithNamespace(framework.Namespace).
		Resource("pod", framework.Name).
		OutputJsonpath("{.status.containerStatuses[?(@.name==\"" + constants.CollectorName + "\")].state}").
		Run()
	Expect(err).To(BeNil())
	Expect(stateJson).To(ContainSubstring("running"), "Expected collector to be running, not in CrashLoopBackOff")
}

// waitForBufferDataFile polls until a .dat file with substantial data appears
// in the disk buffer directory and returns its path. The minimum size ensures
// the burst messages have actually been buffered (not just leftover baseline records).
func waitForBufferDataFile(framework *functional.CollectorFunctionalFramework, dataPath string) string {
	const minBufferSize = 10000
	findCmd := fmt.Sprintf(`ls -t %s/buffer/v2/*/buffer-data-*.dat 2>/dev/null | head -1`, dataPath)
	var datFilePath string
	err := wait.PollUntilContextTimeout(context.TODO(), 2*time.Second, 90*time.Second, true, func(ctx context.Context) (bool, error) {
		out, err := framework.RunCommand(functional.ToolsContainerName, "bash", "-c", findCmd)
		if err != nil || strings.TrimSpace(out) == "" {
			return false, nil
		}
		candidate := strings.TrimSpace(out)
		sizeStr, err := framework.RunCommand(functional.ToolsContainerName, "bash", "-c",
			fmt.Sprintf("wc -c < '%s'", candidate))
		if err != nil {
			return false, nil
		}
		size := 0
		_, _ = fmt.Sscanf(strings.TrimSpace(sizeStr), "%d", &size)
		if size < minBufferSize {
			return false, nil
		}
		datFilePath = candidate
		return true, nil
	})
	Expect(err).To(BeNil(), "Buffer data file did not appear within timeout")
	return datFilePath
}

// waitForCollectorStartAndValidateCorruptionDetected polls collector logs until
// Vector has started and asserts that corrupted records were detected during seek.
func waitForCollectorStartAndValidateCorruptionDetected(framework *functional.CollectorFunctionalFramework) {
	var collectorLogs string
	err := wait.PollUntilContextTimeout(context.TODO(), 2*time.Second, 90*time.Second, true, func(ctx context.Context) (bool, error) {
		var pollErr error
		collectorLogs, pollErr = oc.Literal().From("oc logs -n %s pod/%s -c %s", framework.Namespace, framework.Name, constants.CollectorName).Run()
		if pollErr != nil {
			return false, nil
		}
		return strings.Contains(collectorLogs, "Vector has started."), nil
	})
	Expect(err).To(BeNil(), "Vector did not start after restart with corrupted buffer")
	Expect(collectorLogs).To(ContainSubstring("Corrupted record found during buffer initialization seek"),
		"Expected collector logs to contain warning about corrupted records during buffer seek")
}

// findVectorPid locates the Vector process belonging to a given container by
// matching the container ID from pod status against /proc/<pid>/cgroup.
func findVectorPid(framework *functional.CollectorFunctionalFramework, containerName string) string {
	var containerID string
	for _, cs := range framework.Pod.Status.ContainerStatuses {
		if cs.Name == containerName {
			parts := strings.SplitN(cs.ContainerID, "://", 2)
			if len(parts) == 2 {
				containerID = parts[1]
			}
			break
		}
	}
	Expect(containerID).ToNot(BeEmpty(), "Container ID not found for %s", containerName)

	cmd := fmt.Sprintf(`
for d in /proc/[0-9]*; do
  pid="${d##*/}"
  if grep -q '%s' "$d/cgroup" 2>/dev/null; then
    read -r comm < "$d/comm" 2>/dev/null
    if [ "$comm" = "vector" ]; then
      echo "$pid"
      break
    fi
  fi
done`, containerID)
	pid, err := framework.RunCommand(functional.ToolsContainerName, "sh", "-c", cmd)
	pid = strings.TrimSpace(pid)
	Expect(err).To(BeNil(), "Failed to find Vector PID for container %s", containerName)
	Expect(pid).ToNot(BeEmpty(), "Vector PID not found for container %s", containerName)
	return pid
}

// waitForCollectorRestart polls until the collector container has restarted.
func waitForCollectorRestart(framework *functional.CollectorFunctionalFramework) {
	err := wait.PollUntilContextTimeout(context.TODO(), 2*time.Second, 90*time.Second, true, func(ctx context.Context) (bool, error) {
		countStr, err := oc.Get().
			WithNamespace(framework.Namespace).
			Resource("pod", framework.Name).
			OutputJsonpath("{.status.containerStatuses[?(@.name==\"" + constants.CollectorName + "\")].restartCount}").
			Run()
		if err != nil {
			return false, nil
		}
		count, err := strconv.Atoi(strings.TrimSpace(countStr))
		if err != nil {
			return false, nil
		}
		return count >= 1, nil
	})
	Expect(err).To(BeNil(), "Collector did not restart within timeout")
}

// corruptBufferFile reads a buffer data file from the tools sidecar,
// corrupts record payloads locally, and writes the corrupted file back in
// base64 chunks to avoid ARG_MAX limits. Returns the number of corrupted records.
func corruptBufferFile(framework *functional.CollectorFunctionalFramework, datFilePath string) int {
	readCmd := fmt.Sprintf("base64 -w0 '%s'", datFilePath)
	b64Data, err := framework.RunCommand(functional.ToolsContainerName, "bash", "-c", readCmd)
	Expect(err).To(BeNil(), "Failed to read buffer file")
	Expect(b64Data).ToNot(BeEmpty(), "Buffer file is empty")

	rawData, err := base64.StdEncoding.DecodeString(strings.TrimSpace(b64Data))
	Expect(err).To(BeNil(), "Failed to decode base64 data")

	tmpFile, err := os.CreateTemp("", "buffer-data-*.dat")
	Expect(err).To(BeNil())
	_ = tmpFile.Close()
	defer func() { _ = os.Remove(tmpFile.Name()) }()
	Expect(os.WriteFile(tmpFile.Name(), rawData, 0o640)).To(Succeed())

	corruptedCount, corruptErr := corruptDiskBufferPayload(tmpFile.Name())
	Expect(corruptErr).To(BeNil(), "Failed to corrupt buffer payload")
	Expect(corruptedCount).To(BeNumerically(">=", 1),
		"Expected at least 1 record to be corrupted")

	corruptedData, err := os.ReadFile(tmpFile.Name())
	Expect(err).To(BeNil())
	corruptedB64 := base64.StdEncoding.EncodeToString(corruptedData)

	stagingFile := datFilePath + ".b64"
	_, err = framework.RunCommand(functional.ToolsContainerName, "bash", "-c", fmt.Sprintf("true > '%s'", stagingFile))
	Expect(err).To(BeNil())

	const chunkSize = 65536
	for i := 0; i < len(corruptedB64); i += chunkSize {
		end := min(i+chunkSize, len(corruptedB64))
		chunk := corruptedB64[i:end]
		appendCmd := fmt.Sprintf("echo -n '%s' >> '%s'", chunk, stagingFile)
		_, err = framework.RunCommand(functional.ToolsContainerName, "bash", "-c", appendCmd)
		Expect(err).To(BeNil(), "Failed to write chunk to staging file")
	}

	decodeCmd := fmt.Sprintf("base64 -d '%s' > '%s' && rm -f '%s'", stagingFile, datFilePath, stagingFile)
	_, err = framework.RunCommand(functional.ToolsContainerName, "bash", "-c", decodeCmd)
	Expect(err).To(BeNil(), "Failed to write corrupted buffer file back")

	sizeCmd := fmt.Sprintf("wc -c < '%s'", datFilePath)
	sizeStr, err := framework.RunCommand(functional.ToolsContainerName, "bash", "-c", sizeCmd)
	Expect(err).To(BeNil())
	Expect(strings.TrimSpace(sizeStr)).To(Equal(fmt.Sprintf("%d", len(corruptedData))),
		"Corrupted buffer file size on disk does not match expected size")

	return corruptedCount
}
