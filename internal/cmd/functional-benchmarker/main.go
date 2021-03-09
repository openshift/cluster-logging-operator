package main

import (
	"flag"
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/ViaQ/logerr/log"

	logging "github.com/openshift/cluster-logging-operator/pkg/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/pkg/constants"
	"github.com/openshift/cluster-logging-operator/test"
	"github.com/openshift/cluster-logging-operator/test/client"
	"github.com/openshift/cluster-logging-operator/test/functional"
	"github.com/openshift/cluster-logging-operator/test/helpers/types"
)

// HACK - This command is for development use only
func main() {

	image := flag.String("image", "quay.io/openshift/origin-logging-fluentd:latest", "The image to use to run the benchmark")
	totalMessages := flag.Uint64("totMessages", 10000, "The number of messages to write")
	msgSize := flag.Uint64("size", 1024, "The message size in bytes")
	verbosity := flag.Int("verbosity", 0, "")
	doCleanup := flag.Bool("docleanup", true, "set to false to preserve the namespace")
	sample := flag.Bool("sample", false, "set to true to dump a sample message")

	flag.Parse()

	log.MustInit("functional-benchmark")
	log.SetLogLevel(*verbosity)
	log.V(1).Info("Starting functional benchmarker", "args", os.Args)

	if err := os.Setenv(constants.FluentdImageEnvVar, *image); err != nil {
		log.Error(err, "Error setting fluent image env var")
		os.Exit(1)
	}
	testclient := client.NewNamesapceClient()
	framework := functional.NewFluentdFunctionalFrameworkUsing(&testclient.Test, testclient.Close, *verbosity)
	if *doCleanup {
		defer framework.Cleanup()
	}

	functional.NewClusterLogForwarderBuilder(framework.Forwarder).
		FromInput(logging.InputNameApplication).
		ToFluentForwardOutput()
	if err := framework.Deploy(); err != nil {
		log.Error(err, "Error deploying test pod")
		os.Exit(1)
	}
	startTime := time.Now()
	var (
		logs    []string
		readErr error
	)
	done := make(chan bool)
	go func() {
		logs, readErr = framework.ReadNApplicationLogsFrom(*totalMessages, logging.OutputTypeFluentdForward)
		done <- true
	}()
	//defer reader to get logs
	if err := framework.WritesNApplicationLogsOfSize(*totalMessages, *msgSize); err != nil {
		log.Error(err, "Error writing logs to test pod")
		os.Exit(1)
	}
	<-done
	endTime := time.Now()
	if readErr != nil {
		log.Error(readErr, "Error reading logs")
		os.Exit(1)
	}
	log.V(4).Info("Read logs", "raw", logs)
	jsonlogs, err := types.ParseLogs(fmt.Sprintf("[%s]", strings.Join(logs, ",")))
	if err != nil {
		log.Error(err, "Error parsing logs")
		os.Exit(1)
	}
	log.V(4).Info("Read logs", "parsed", jsonlogs)
	if *sample {
		fmt.Printf("Sample:\n%s\n", test.JSONString(jsonlogs[0]))
	}
	timeDiffs := sortLogsByTimeDiff(jsonlogs)
	fmt.Printf("  Total Msg: %d\n", *totalMessages)
	fmt.Printf("Size(bytes): %d\n", *msgSize)
	fmt.Printf(" Elapsed(s): %s\n", endTime.Sub(startTime))
	fmt.Printf("    Mean(s): %f\n", mean(jsonlogs))
	fmt.Printf("     Min(s): %f\n", min(timeDiffs))
	fmt.Printf("     Max(s): %f\n", max(timeDiffs))
	fmt.Printf("  Median(s): %f\n", median(timeDiffs))
	fmt.Printf(" Mean Bloat: %f\n", meanBloat(jsonlogs))
}

func meanBloat(logs types.Logs) float64 {
	return genericMean(logs, (*types.AllLog).Bloat)
}

func mean(logs types.Logs) float64 {
	return genericMean(logs, (*types.AllLog).ElapsedEpoc)
}

func genericMean(logs types.Logs, f func(l *types.AllLog) float64) float64 {
	if len(logs) == 0 {
		return 0
	}
	var total float64
	for i := range logs {
		total += f(&logs[i])
	}
	return total / float64(len(logs))
}

func median(diffs []float64) float64 {
	if len(diffs) == 0 {
		return 0
	}
	return diffs[(len(diffs) / 2)]
}

func min(diffs []float64) float64 {
	if len(diffs) == 0 {
		return 0
	}
	return diffs[0]
}

func max(diffs []float64) float64 {
	if len(diffs) == 0 {
		return 0
	}
	return diffs[len(diffs)-1]
}

func sortLogsByTimeDiff(logs types.Logs) []float64 {
	diffs := make([]float64, len(logs))
	for i := range logs {
		diffs[i] = logs[i].ElapsedEpoc()
	}
	sort.Slice(diffs, func(i, j int) bool { return diffs[i] < diffs[j] })
	return diffs
}
