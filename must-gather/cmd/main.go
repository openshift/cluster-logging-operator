package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/openshift/cluster-logging-operator/must-gather"
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
	gather, err := mustgather.NewGather(destDir, loggingNamespace, logFileName, multiWriter)
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
