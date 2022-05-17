package gnuplot

import (
	"fmt"
	htmllib "html"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/ViaQ/logerr/v2/log"
	"github.com/openshift/cluster-logging-operator/internal/cmd/functional-benchmarker/config"
	"github.com/openshift/cluster-logging-operator/internal/cmd/functional-benchmarker/stats"
)

type GNUPlotReporter struct {
	Options     config.Options
	Metrics     stats.ResourceMetrics
	Stats       stats.Statistics
	ArtifactDir string
}

func exportResourceMetricTo(outDir string, rm stats.ResourceMetrics) {
	exportCPU(outDir, rm.Samples)
	exportMemory(outDir, rm.Samples)
}

/* #nosec G306*/
func exportLatency(outDir string, logs stats.PerfLogs) {
	logger := log.NewLogger("")
	logger.V(3).Info("Exporting latency")
	buffer := []string{}
	for i, log := range logs {
		value := log.ElapsedEpoc()
		buffer = append(buffer, fmt.Sprintf("%d %.3f", i, value))
	}
	logger.V(3).Info("Writing latency data", "outDir", outDir, "data", buffer)
	err := ioutil.WriteFile(path.Join(outDir, "latency.data"), []byte(strings.Join(buffer, "\n")), 0755)
	if err != nil {
		logger.Error(err, "Error writing latency data")
	}
}

/* #nosec G306*/
func exportMemory(outDir string, samples []stats.Sample) {
	logger := log.NewLogger("")
	logger.V(3).Info("Exporting memory", "samples", samples)
	buffer := []string{}
	for _, sample := range samples {
		value := sample.MemoryBytesAsFloat()
		buffer = append(buffer, fmt.Sprintf("%d %.3f", sample.Time, value/1024/1024))
	}
	logger.V(3).Info("Writing resource metric data", "outDir", outDir, "memoryInMb", buffer)
	err := ioutil.WriteFile(path.Join(outDir, "mem.data"), []byte(strings.Join(buffer, "\n")), 0755)
	if err != nil {
		logger.Error(err, "Error writing resource Memory Metrics")
	}
}

/* #nosec G306*/
func exportCPU(outDir string, samples []stats.Sample) {
	logger := log.NewLogger("")
	logger.V(3).Info("Exporting cpu", "samples", samples)
	buffer := []string{}
	for _, sample := range samples {
		cpu := sample.CPUCoresAsFloat()
		buffer = append(buffer, fmt.Sprintf("%d %.3f", sample.Time, cpu))
	}
	logger.V(3).Info("Writing resource metric data", "outDir", outDir, "cpu", buffer)
	err := ioutil.WriteFile(path.Join(outDir, "cpu.data"), []byte(strings.Join(buffer, "\n")), 0755)
	if err != nil {
		logger.Error(err, "Error writing resource CPU Metrics")
	}
}

func (r *GNUPlotReporter) Generate() {
	exportResourceMetricTo(r.ArtifactDir, r.Metrics)
	exportLatency(r.ArtifactDir, r.Stats.Logs)
	for _, plot := range []string{memPlotPNG, cpuPlotPNG, latencyPlotPNG} {
		plotData(plot, r.ArtifactDir, nil)
	}
	for _, plot := range []string{memPlotDumb, cpuPlotDumb, latencyPlotDumb} {
		plotData(plot, r.ArtifactDir, os.Stdout)
	}
	r.generateStats()
}

func plotData(plot, dir string, writer io.Writer) {
	plot = strings.Join(strings.Split(plot, "\n"), ";")
	logger := log.NewLogger("")
	logger.V(3).Info("running gnuplot", "cmd", plot)
	cmd := exec.Command("gnuplot", "-e", plot)
	cmd.Dir = dir
	if writer != nil {
		cmd.Stdout = writer
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		logger.Error(err, "Unable to create StderrPipe")
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		logger.Error(err, "Unable to create StdOutPipe")
	}
	if err := cmd.Run(); err != nil {
		logger.Error(err, "Error running command", "dir", cmd.Dir, "cmd", cmd.String())

		if stderr != nil {
			raw, err := ioutil.ReadAll(stderr)
			logger.Info("Reading stderr", "stderr", raw, "err", err)
		}

		if stdout != nil {
			raw, err := ioutil.ReadAll(stdout)
			logger.Info("Reading stdout", "stdout", raw, "stdout", err)
		}
	}
}

/* #nosec G306*/
func (r *GNUPlotReporter) generateStats() {
	reports := map[string]string{
		"results.html": html,
		"readme.md":    markdown,
	}
	s := r.Stats
	o := r.Options
	escapeFn := func(content string) string {
		return content
	}
	for file, template := range reports {
		if template == html {
			escapeFn = htmllib.EscapeString
		}
		out := fmt.Sprintf(template,
			o.Image,
			o.TotalLogStressors,
			o.LinesPerSecond,
			o.RunDuration,
			o.PayloadSource,
			s.TotMessages(),
			s.MsgSize,
			s.Elapsed.Round(time.Second),
			s.Mean(),
			s.Min(),
			s.Max(),
			s.Median(),
			escapeFn(o.CollectorConfig),
		)
		err := ioutil.WriteFile(path.Join(r.ArtifactDir, file), []byte(out), 0755)
		if err != nil {
			log.NewLogger("").Error(err, "Error writing file")
		}
	}
	writeStatsToConsole(s)
}
func writeStatsToConsole(s stats.Statistics) {
	w := new(tabwriter.Writer)
	// minwidth, tabwidth, padding, padchar, flags
	w.Init(os.Stdout, 3, 2, 2, ' ', tabwriter.AlignRight)

	defer w.Flush()
	div := "--------"
	headerFmt := "\n %s\t%s\t%s\t%s\t%s\t%s\t%s\t"

	fmt.Fprintf(w, "Latency of logs collected based on the time the log was generated and ingested")
	fmt.Fprintf(w, headerFmt, "Total", "Size", "Elapsed", "Mean", "Min", "Max", "Median")
	fmt.Fprintf(w, headerFmt, "Msg", "(bytes)", "", "(s)", "(s)", "(s)", "(s)")
	fmt.Fprintf(w, headerFmt, div, div, div, div, div, div, div)

	fmt.Fprintf(w, "\n %d\t%d\t%s\t%.3f\t%.3f\t%.3f\t%.3f\t",
		s.TotMessages(),
		s.MsgSize,
		s.Elapsed.Round(time.Second),
		s.Mean(),
		s.Min(),
		s.Max(),
		s.Median())

	fmt.Fprintf(w, "\n")
}
