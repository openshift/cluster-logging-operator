package main

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"os"
	"runtime"
	"strings"
	"time"

	stub "github.com/openshift/elasticsearch-operator/pkg/stub"
	"github.com/openshift/elasticsearch-operator/pkg/utils"
	sdk "github.com/operator-framework/operator-sdk/pkg/sdk"
	k8sutil "github.com/operator-framework/operator-sdk/pkg/util/k8sutil"
	sdkVersion "github.com/operator-framework/operator-sdk/version"

	"github.com/sirupsen/logrus"
)

const (
	// supported log formats
	logFormatLogfmt = "logfmt"
	logFormatJSON   = "json"

	// env vars
	logLevelEnv  = "LOG_LEVEL"
	logFormatEnv = "LOG_FORMAT"
)

var (
	logLevel            string
	logFormat           string
	availableLogLevels  string
	availableLogFormats = []string{
		logFormatLogfmt,
		logFormatJSON,
	}
	// this is a constant, but can't be in the `const` section because
	// the value is a runtime function return value
	defaultLogLevel = logrus.InfoLevel.String()
)

func printVersion() {
	logrus.Infof("Go Version: %s", runtime.Version())
	logrus.Infof("Go OS/Arch: %s/%s", runtime.GOOS, runtime.GOARCH)
	logrus.Infof("operator-sdk Version: %v", sdkVersion.Version)
}

func init() {
	// create availableLogLevels
	buf := &bytes.Buffer{}
	comma := ""
	for _, logrusLevel := range logrus.AllLevels {
		buf.WriteString(comma)
		buf.WriteString(logrusLevel.String())
		comma = ", "
	}
	availableLogLevels = buf.String()
	// default values are ""
	// that means that if no arguments are provided env variables take precedence
	// otherwise command-line arguments take precendence
	flagset := flag.CommandLine
	flagset.StringVar(&logLevel, "log-level", "", fmt.Sprintf("Log level to use. Possible values: %s", availableLogLevels))
	flagset.StringVar(&logFormat, "log-format", "", fmt.Sprintf("Log format to use. Possible values: %s", strings.Join(availableLogFormats, ", ")))
	flagset.Parse(os.Args[1:])
}

func initLogger() error {
	// first check cmd arguments, then environment variables
	if logLevel == "" {
		logLevel = utils.LookupEnvWithDefault(logLevelEnv, defaultLogLevel)
	}
	if logFormat == "" {
		logFormat = utils.LookupEnvWithDefault(logFormatEnv, logFormatLogfmt)
	}

	// set log level, default to info level
	level, err := logrus.ParseLevel(logLevel)
	if err != nil {
		return fmt.Errorf("log level '%s' unknown.  Possible values: %v", logLevel, availableLogLevels)
	}
	logrus.SetLevel(level)

	// set log format, default to text formatter
	switch logFormat {
	case logFormatLogfmt:
		logrus.SetFormatter(&logrus.TextFormatter{})
		break
	case logFormatJSON:
		logrus.SetFormatter(&logrus.JSONFormatter{})
		break
	default:
		return fmt.Errorf("log format '%s' unknown, %v are possible values", logFormat, availableLogFormats)
	}
	// log to stdout; logrus defaults to stderr
	logrus.SetOutput(os.Stdout)

	return nil
}

func Main() int {
	if err := initLogger(); err != nil {
		fmt.Fprintf(os.Stderr, "instantiating elasticsearch controller failed: %v\n", err)
		return 1
	}
	printVersion()

	sdk.ExposeMetricsPort()

	resource := "logging.openshift.io/v1alpha1"
	kind := "Elasticsearch"
	namespace, err := k8sutil.GetWatchNamespace()
	if err != nil {
		logrus.Fatalf("Failed to get watch namespace: %v", err)
	}
	resyncPeriod := time.Duration(5) * time.Second
	logrus.Infof("Watching %s, %s, %s, %d", resource, kind, namespace, resyncPeriod)
	sdk.Watch(resource, kind, namespace, resyncPeriod)
	sdk.Handle(stub.NewHandler())
	sdk.Run(context.TODO())
	return 0
}

func main() {
	os.Exit(Main())
}
