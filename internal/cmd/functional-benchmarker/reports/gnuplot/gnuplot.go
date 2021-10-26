package gnuplot

import (
	"fmt"
	"github.com/ViaQ/logerr/log"
	"github.com/openshift/cluster-logging-operator/internal/cmd/functional-benchmarker/stats"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"strings"
	"text/tabwriter"
	"time"
)

type GNUPlotReporter struct {
	Metrics     stats.ResourceMetrics
	Stats       stats.Statistics
	ArtifactDir string
}

func exportResourceMetricTo(outDir string, rm stats.ResourceMetrics) {
	exportCPU(outDir, rm.Samples)
	exportMemory(outDir, rm.Samples)
}

func exportMemory(outDir string, samples []stats.Sample) {
	log.V(3).Info("Exporting memory", "samples", samples)
	buffer := []string{}
	for _, sample := range samples {
		value := sample.MemoryBytesAsFloat()
		buffer = append(buffer, fmt.Sprintf("%d %.3f", sample.Time, value/1024/1024))
	}
	log.V(3).Info("Writing resource metric data", "outDir", outDir, "memoryInMb", buffer)
	err := ioutil.WriteFile(path.Join(outDir, "mem.data"), []byte(strings.Join(buffer, "\n")), 0600)
	if err != nil {
		log.Error(err, "Error writing resource Memory Metrics")
	}
}
func exportCPU(outDir string, samples []stats.Sample) {
	log.V(3).Info("Exporting cpu", "samples", samples)
	buffer := []string{}
	for _, sample := range samples {
		cpu := sample.CPUCoresAsFloat()
		buffer = append(buffer, fmt.Sprintf("%d %.3f", sample.Time, cpu))
	}
	log.V(3).Info("Writing resource metric data", "outDir", outDir, "cpu", buffer)
	err := ioutil.WriteFile(path.Join(outDir, "cpu.data"), []byte(strings.Join(buffer, "\n")), 0600)
	if err != nil {
		log.Error(err, "Error writing resource CPU Metrics")
	}
}

func (r *GNUPlotReporter) Generate() {
	exportResourceMetricTo(r.ArtifactDir, r.Metrics)
	for _, plot := range []string{memPlotPNG, cpuPlotPNG} {
		plotData(plot, r.ArtifactDir, nil)
	}
	for _, plot := range []string{memPlotDumb, cpuPlotDumb} {
		plotData(plot, r.ArtifactDir, os.Stdout)
	}
	r.generateStats()
}

func plotData(plot, dir string, writer io.Writer) {
	plot = strings.Join(strings.Split(plot, "\n"), ";")
	log.V(3).Info("running gnuplot", "cmd", plot)
	cmd := exec.Command("gnuplot", "-e", plot)
	cmd.Dir = dir
	if writer != nil {
		cmd.Stdout = writer
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		log.Error(err, "Unable to create StderrPipe")
	}
	if err := cmd.Run(); err != nil {
		log.Error(err, "Error running command", "cmd", plot)
		raw, err := ioutil.ReadAll(stderr)
		log.Info("Reading stderr", "stderr", raw, "err", err)
	}
}

func (r *GNUPlotReporter) generateStats() {
	s := r.Stats
	out := fmt.Sprintf(html,
		s.TotMessages(),
		s.MsgSize,
		s.Elapsed.Round(time.Second),
		s.Mean(),
		s.Min(),
		s.Max(),
		s.Median())
	err := ioutil.WriteFile(path.Join(r.ArtifactDir, "result.html"), []byte(out), 0600)
	if err != nil {
		log.Error(err, "Error writing html file")
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
