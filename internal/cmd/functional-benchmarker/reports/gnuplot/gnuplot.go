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

	log "github.com/ViaQ/logerr/v2/log/static"
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
	log.V(3).Info("Exporting latency")
	buffer := []string{}
	for i, log := range logs {
		value := log.ElapsedEpoc()
		buffer = append(buffer, fmt.Sprintf("%d %.3f", i, value))
	}
	log.V(3).Info("Writing latency data", "outDir", outDir, "data", buffer)
	err := ioutil.WriteFile(path.Join(outDir, "latency.data"), []byte(strings.Join(buffer, "\n")), 0755)
	if err != nil {
		log.Error(err, "Error writing latency data")
	}
}

/* #nosec G306*/
func exportMemory(outDir string, samples []stats.Sample) {
	log.V(3).Info("Exporting memory", "samples", samples)
	buffer := []string{}
	for _, sample := range samples {
		value := sample.MemoryBytesAsFloat()
		buffer = append(buffer, fmt.Sprintf("%d %.3f", sample.Time, value/1024/1024))
	}
	log.V(3).Info("Writing resource metric data", "outDir", outDir, "memoryInMb", buffer)
	err := ioutil.WriteFile(path.Join(outDir, "mem.data"), []byte(strings.Join(buffer, "\n")), 0755)
	if err != nil {
		log.Error(err, "Error writing resource Memory Metrics")
	}
}

/* #nosec G306*/
func exportCPU(outDir string, samples []stats.Sample) {
	log.V(3).Info("Exporting cpu", "samples", samples)
	buffer := []string{}
	for _, sample := range samples {
		cpu := sample.CPUCoresAsFloat()
		buffer = append(buffer, fmt.Sprintf("%d %.3f", sample.Time, cpu))
	}
	log.V(3).Info("Writing resource metric data", "outDir", outDir, "cpu", buffer)
	err := ioutil.WriteFile(path.Join(outDir, "cpu.data"), []byte(strings.Join(buffer, "\n")), 0755)
	if err != nil {
		log.Error(err, "Error writing resource CPU Metrics")
	}
}

/* #nosec G306*/
func exportLoss(outDir string, samples stats.LossStats) {
	log.V(3).Info("Exporting message losses", "samples", samples)
	buffer := []string{}
	for _, stream := range samples.Streams() {
		streamStats, err := samples.LossStatsFor(stream)
		if err != nil {
			log.Error(err, "Unable to generate stats", "stream", stream)
			return
		}
		if len(streamStats.Entries) == 0 {
			log.V(0).Info("No entries returned for stream", "stream", stream, "streamStats", streamStats)
			continue
		}
		lostLogs := 0
		i := 0
		for expSeqId := streamStats.MinSeqId; expSeqId <= streamStats.MaxSeqId; expSeqId++ {
			seqId := streamStats.Entries[i].SequenceId
			if seqId != expSeqId {
				lostLogs += 1
			} else {
				i += 1 //found entry
			}
			buffer = append(buffer, fmt.Sprintf("%d %d", expSeqId, lostLogs))
		}

		log.V(3).Info("Writing message losses data", "outDir", outDir, "losses", buffer)
		err = ioutil.WriteFile(path.Join(outDir, fmt.Sprintf("%s-loss.data", stream)), []byte(strings.Join(buffer, "\n")), 0755)
		if err != nil {
			log.Error(err, "Error writing message loss metrics", "stream", stream)
		}

	}
}

func formatLossPlot(base string, samples stats.LossStats) string {
	buffer := base
	for i, stream := range samples.Streams() {
		if i == 0 {
			buffer = fmt.Sprintf("%s;plot '%s-loss.data' using 1:2 title '%s' with lines", buffer, stream, stream)
		} else {
			buffer = fmt.Sprintf("%s,'%s-loss.data' using 1:2 title '%s' with lines", buffer, stream, stream)
		}
	}
	return buffer
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

	exportLoss(r.ArtifactDir, r.Stats.Losses)
	plotData(formatLossPlot(lossPlotPNG, r.Stats.Losses), r.ArtifactDir, nil)
	plotData(formatLossPlot(lossPlotDumb, r.Stats.Losses), r.ArtifactDir, os.Stdout)

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

	err := cmd.Run()
	if err != nil {
		log.Error(err, "Error starting command", "dir", cmd.Dir, "cmd", cmd.String())
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

	var lossFmtFun func() string
	var escapeFn func(content string) string
	for file, template := range reports {
		if template == html {
			escapeFn = htmllib.EscapeString
			lossFmtFun = func() string {
				losses := ""
				for _, name := range s.Losses.Streams() {
					streamLoss, _ := s.Losses.LossStatsFor(name)
					losses += fmt.Sprintf("<tr><td>%s</td><td>%d</td><td>%d</td><td>%d</td><td>%d</td><td>%.1f%%</td><tr>",
						name,
						streamLoss.MinSeqId,
						streamLoss.MaxSeqId,
						streamLoss.Purged,
						streamLoss.Collected,
						streamLoss.PercentCollected())
				}
				return losses
			}
		}
		if template == markdown {
			escapeFn = func(content string) string {
				return "```\n" + content + "\n```"
			}
			lossFmtFun = func() string {
				losses := ""
				for _, name := range s.Losses.Streams() {
					streamLoss, _ := s.Losses.LossStatsFor(name)
					losses += fmt.Sprintf("| %s|%d|%d|%d|%d|%.1f%%\n",
						name,
						streamLoss.MinSeqId,
						streamLoss.MaxSeqId,
						streamLoss.Purged,
						streamLoss.Collected,
						streamLoss.PercentCollected())
				}
				return losses
			}
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
			lossFmtFun(),
			escapeFn(o.CollectorConfig),
		)
		err := ioutil.WriteFile(path.Join(r.ArtifactDir, file), []byte(out), 0755)
		if err != nil {
			log.Error(err, "Error writing file")
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

	headerFmt = "\n %s\t%s\t%s\t%s\t%s\t%s\t"
	fmt.Fprintf(w, "\nPercent logs lost between first and last collected sequence ids")
	fmt.Fprintf(w, headerFmt, "Stream", "Min", "Max", "Purged", "Collected", "Per. Coll.")
	fmt.Fprintf(w, headerFmt, div, div, div, div, div, div)
	for _, name := range s.Losses.Streams() {
		streamLoss, _ := s.Losses.LossStatsFor(name)
		fmt.Fprintf(w, "\n%s\t%d\t%d\t%d\t%d\t\t%.1f%%\n",
			name,
			streamLoss.MinSeqId,
			streamLoss.MaxSeqId,
			streamLoss.Purged,
			streamLoss.Collected,
			streamLoss.PercentCollected())
	}
	fmt.Fprintf(w, "\n")
}
