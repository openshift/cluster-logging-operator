package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"regexp"
	yaml "sigs.k8s.io/yaml"
	"strconv"
	"strings"
	"time"

	"github.com/ViaQ/logerr/log"
	"github.com/openshift/cluster-logging-operator/internal/pkg/generator/forwarder"
	logging "github.com/openshift/cluster-logging-operator/pkg/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/pkg/constants"
	"github.com/openshift/cluster-logging-operator/test/functional"
	"github.com/openshift/cluster-logging-operator/test/helpers/podman"
	"os"
)

const (
	logStressorImage = "quay.io/openshift-logging/log-stressor:latest"
	collectorConf    = "collector.conf"
	containerDir     = "/var/log/containers"
	collectorIdKey   = "collector"
	stressorIdKey    = "stressors"
	receiversIdKey   = "receiver"
)

type LocalRunner struct {
	Options
	collector       podman.Runner
	logStressors    []podman.Runner
	receiver        podman.Runner
	ids             map[string][]string
	cleanup         []func() error
	tmpDir          string
	collectorConfig string
}

func NewLocalRunner(collectorConfig string, options Options) *LocalRunner {
	var loglevel string
	switch options.verbosity {
	case 2:
		loglevel = "debug"
	case 3:
		loglevel = "trace"
	default:
		loglevel = "info"
	}
	runner := &LocalRunner{
		Options:         options,
		collectorConfig: collectorConfig,
		ids:             map[string][]string{},
		cleanup:         []func() error{},
		collector: podman.Run().
			Named(collectorIdKey).
			AsPrivileged(true).
			Detached(true).
			WithImage(os.Getenv(constants.FluentdImageEnvVar)).
			WithEnvVar("LOG_LEVEL", loglevel),
	}
	runner.receiver = podman.Run().
		Named("receiver").
		AsPrivileged(true).
		Detached(true).
		WithImage(os.Getenv(constants.FluentdImageEnvVar))

	for i := 0; i < options.totalLogStressors; i++ {
		stressor := podman.Run().
			Named(fmt.Sprintf("log-stressor-%d", i)).
			AsPrivileged(true).
			Detached(true).
			WithImage(logStressorImage).
			WithEnvVar("OUTPUT_FORMAT", "crio").
			WithEnvVar("PAYLOAD_GEN", "fixed").
			WithEnvVar("PAYLOAD_SIZE", strconv.Itoa(options.msgSize)).
			WithEnvVar("TOT_MSG", strconv.Itoa(options.totalMessages))
		runner.logStressors = append(runner.logStressors, stressor)
	}

	return runner
}

func (r *LocalRunner) Deploy() {
	tmpDir, err := ioutil.TempDir(".", "benchmark")
	if err != nil {
		log.Error(err, "Error creating temp director")
		os.Exit(1)
	}
	if err := os.Chmod(tmpDir, 0766); err != nil {
		log.Error(err, "Error modifying temp director permissions")
		os.Exit(1)
	}
	if r.tmpDir, err = filepath.Abs(tmpDir); err != nil {
		log.Error(err, "Unable to determine the absolute file path of the tmpDir")
		os.Exit(1)
	}
	log.V(2).Info("Created directory", "tempdir", tmpDir)
	r.cleanup = append(r.cleanup, func() error {
		return os.RemoveAll(r.tmpDir)
	})

	r.deployReceivers(r.tmpDir)
	r.deployCollector(r.tmpDir)

}
func (r *LocalRunner) deployCollector(tmpDir string) {
	r.writeCollectorConfig(tmpDir)
	r.collector.WithVolume(tmpDir, "/tmp/config").
		WithVolume(tmpDir, containerDir).
		WithNetwork("host").
		WithCmd("fluentd", "-c", filepath.Join("/tmp/config", collectorConf), "--no-supervisor")

	out, err := r.collector.Run()
	if err != nil {
		log.Error(err, "Error deploying collector")
		os.Exit(1)
	}
	r.ids[collectorIdKey] = []string{out}
}
func (r *LocalRunner) deployStressors(tmpDir string) {
	for i, stressor := range r.logStressors {
		stressor.WithVolume(tmpDir, containerDir).
			WithEnvVar("OUT_FILE", filepath.Join(containerDir, fmt.Sprintf("log-stressor-%d_fakenamepace_hash123.log", i)))
	}
	ids := []string{}
	for _, stressor := range r.logStressors {
		out, err := stressor.Run()
		if err != nil {
			log.Error(err, "Error deploying LogStressor")
			os.Exit(1)
		}
		ids = append(ids, out)
	}
	r.ids[stressorIdKey] = ids
}
func (r *LocalRunner) deployReceivers(tmpDir string) {
	err := ioutil.WriteFile(filepath.Join(tmpDir, "fluent.conf"), []byte(functional.UnsecureFluentConfBenchmark), 0600)
	if err != nil {
		log.Error(err, "Error writing benchmark receiver config")
		os.Exit(1)
	}
	r.receiver.WithVolume(tmpDir, "/tmp/config").
		WithNetwork("host").
		WithCmd("fluentd", "-c", "/tmp/config/fluent.conf")
	out, err := r.receiver.Run()
	if err != nil {
		log.Error(err, "Error deploying receiver")
		os.Exit(1)
	}
	r.ids[receiversIdKey] = []string{out}
}

func (r *LocalRunner) writeCollectorConfig(dir string) {
	conf := r.collectorConfig
	var err error
	if conf == "" {
		f := &logging.ClusterLogForwarder{}
		functional.NewClusterLogForwarderBuilder(f).
			FromInput(logging.InputNameApplication).
			ToFluentForwardOutput()

		log.V(2).Info("Generating config", "forwarder", f)
		clfYaml, _ := yaml.Marshal(f)

		if conf, err = forwarder.Generate(string(clfYaml), false, false); err != nil {
			log.Error(err, "Error generating config")
			os.Exit(1)
		}
	}
	log.V(1).Info("Preparing to modifying config for benchmarking", "config", conf)
	m := regexp.MustCompile("cache_size.*")
	conf = m.ReplaceAllString(conf, "test_api_adapter KubernetesMetadata::TestApiAdapter")

	m = regexp.MustCompile("(?s)<ssl>(.*)</ssl>")
	conf = m.ReplaceAllString(conf, "")

	if err = ioutil.WriteFile(filepath.Join(dir, collectorConf), []byte(conf), 0600); err != nil {
		log.Error(err, "Error writing config", "tempdir", dir)
		os.Exit(1)
	}
}

func (r *LocalRunner) WritesApplicationLogsOfSize(msgSize int) error {
	r.deployStressors(r.tmpDir)
	return nil
}

func (r *LocalRunner) ReadApplicationLogs() ([]string, error) {
	timeout, err := time.ParseDuration(r.readTimeout)
	if err != nil {
		return nil, err
	}
	n := r.totalMessages * len(r.logStressors)
	reader, err := podman.Exec().
		ToContainer(r.ids[receiversIdKey][0]).
		WithCmd("tail", "-F", "/tmp/app-logs").
		Reader()
	lines := []string{}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	if err != nil {
		log.V(3).Error(err, "Error creating tail reader")
		return nil, err
	}
	for {
		line, err := reader.ReadLineContext(ctx)
		if err != nil {
			log.V(3).Error(err, "Error readlinecontext")
			return nil, err
		}
		lines = append(lines, line)
		n--
		if n == 0 {
			break
		}
	}
	return lines, err

}

func (r *LocalRunner) Cleanup() {
	if out, err := podman.Stop().
		WithContainer(r.ids[collectorIdKey]...).
		WithContainer(r.ids[stressorIdKey]...).
		WithContainer(r.ids[receiversIdKey]...).
		Run(); err != nil {
		log.Error(err, out)
	}
	if out, err := podman.RM().
		WithContainer(r.ids[collectorIdKey]...).
		WithContainer(r.ids[stressorIdKey]...).
		WithContainer(r.ids[receiversIdKey]...).
		Run(); err != nil {
		log.Error(err, out)
	}
	for _, cleanup := range r.cleanup {
		err := cleanup()
		if err != nil {
			log.Error(err, "Error running cleanup function")
		}
	}
}

func (r *LocalRunner) Metrics() Metrics {
	if out, err := podman.Exec().
		ToContainer(collectorIdKey).
		WithCmd("bash", "-c", "cat /proc/1/stat | cut -d ' ' -f 14-15; grep VmPeak /proc/1/status|cut -d ' ' -f3").
		Run(); err == nil {
		segments := strings.Fields(out)
		if len(segments) > 3 {
			log.Error(nil, "Unexpected output", "out", out)
			return Metrics{}
		}
		return Metrics{
			segments[0],
			segments[1],
			segments[2],
		}
	} else {
		log.Error(err, "Unable to retrieve collector metrics")
	}
	return Metrics{}
}
