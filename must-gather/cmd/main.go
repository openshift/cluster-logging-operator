package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	mustgather "github.com/openshift/cluster-logging-operator/must-gather"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func main() {
	var (
		destDir          string
		loggingNamespace string
		logFileName      string
	)

	flag.StringVar(&destDir, "dest-dir", "/must-gather", "Destination directory for collecting must-gather data")
	flag.StringVar(&loggingNamespace, "logging-namespace", "openshift-logging", "Namespace where cluster logging operator is deployed")
	flag.StringVar(&logFileName, "log-file", "gather-debug.log", "Name of the debug log file")
	flag.Parse()

	// oc adm must-gather passes --since/--since-time to the plugin via the
	// MUST_GATHER_SINCE/MUST_GATHER_SINCE_TIME env vars; the plugin itself
	// is run with a fixed command and receives no flags.
	var since time.Duration
	if v := os.Getenv("MUST_GATHER_SINCE"); v != "" {
		d, err := time.ParseDuration(v)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Invalid MUST_GATHER_SINCE value %q: %v\n", v, err)
			os.Exit(1)
		}
		since = d
	}
	sinceTimeStr := os.Getenv("MUST_GATHER_SINCE_TIME")

	if since < 0 {
		fmt.Fprintf(os.Stderr, "MUST_GATHER_SINCE must be a positive duration\n")
		os.Exit(1)
	}

	if since > 0 && sinceTimeStr != "" {
		fmt.Fprintf(os.Stderr, "MUST_GATHER_SINCE and MUST_GATHER_SINCE_TIME are mutually exclusive\n")
		os.Exit(1)
	}

	var sinceSeconds *int64
	if since > 0 {
		s := int64((since + time.Second - 1) / time.Second)
		sinceSeconds = &s
	}

	var sinceTime *metav1.Time
	if sinceTimeStr != "" {
		t, err := time.Parse(time.RFC3339, sinceTimeStr)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Invalid MUST_GATHER_SINCE_TIME value %q: must be RFC3339 (e.g. 2006-01-02T15:04:05Z)\n", sinceTimeStr)
			os.Exit(1)
		}
		sinceTime = &metav1.Time{Time: t}
	}

	// Ensure destination directory exists
	if err := os.MkdirAll(destDir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create destination directory: %v\n", err)
		os.Exit(1)
	}

	// Set up logging
	logFilePath := filepath.Join(destDir, logFileName)
	logFile, err := os.Create(logFilePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create log file: %v\n", err)
		os.Exit(1)
	}
	defer func() {
		if err := logFile.Close(); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to close log file: %v\n", err)
		}
	}()

	// Use multi-writer to write to both file and stdout
	multiWriter := io.MultiWriter(os.Stdout, logFile)

	// Create and run the gather
	gather, err := mustgather.NewGather(destDir, loggingNamespace, logFileName, multiWriter, sinceSeconds, sinceTime)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create gather: %v\n", err)
		os.Exit(1)
	}

	ctx := context.Background()
	if err := gather.Run(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "Must-gather failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Must-gather completed successfully")
}
