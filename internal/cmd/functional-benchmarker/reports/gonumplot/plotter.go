package gonumplot

import (
	"fmt"
	"github.com/openshift/cluster-logging-operator/internal/cmd/functional-benchmarker/stats"
	"github.com/vspaz/wls-go/pkg/models"
	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/plotutil"
	"gonum.org/v1/plot/vg"
	"path"
)

var (
	defaultWidth  = vg.Points(768)
	defaultHeight = vg.Points(576)

	megaBytesPerByte = 1.0 / float64(1024) / float64(1024)
)

type GoNumPlot struct {
	Metrics     stats.ResourceMetrics
	Stats       stats.Statistics
	ArtifactDir string
}

func (p GoNumPlot) Plot() {
	cpu, memory := extractPlotData(p.Metrics.Samples)
	p.exportCpuPlot(cpu)
	p.exportMemoryPlot(memory)
	exportLatencyPlot(p.ArtifactDir, p.Stats.Losses)
	exportLossStatsPlot(p.ArtifactDir, p.Stats.Losses)
}

func exportLatencyPlot(outDir string, stats stats.LossStats) {
	aPlot := plot.New()

	aPlot.Title.Text = "Latency"
	aPlot.Legend.Top = true
	aPlot.X.Label.Text = "Message"
	aPlot.Y.Label.Text = "Seconds"

	var data []interface{}
	for i, stream := range stats.Streams() {
		if streamState, err := stats.LossStatsFor(stream); err == nil {
			var xs []float64
			var ys []float64
			var xys plotter.XYs
			for i, entry := range streamState.Entries {
				xs = append(xs, float64(i))
				ys = append(ys, entry.ElapsedEpoc())
				xys = append(xys, plotter.XY{
					X: float64(i),
					Y: entry.ElapsedEpoc(),
				})
			}
			data = append(data, xys)

			wls := models.NewWlsWithoutWeights(xs, ys)
			point := wls.FitLinearRegression()
			slope := point.GetSlope()
			yIntercept := point.GetIntercept()

			fnLinear := plotter.NewFunction(func(x float64) float64 {
				return slope*x + yIntercept
			})
			fnLinear.Color = plotutil.Color(i)

			aPlot.Add(fnLinear)
			aPlot.Legend.Add(fmt.Sprintf("%s trend", stream), fnLinear)
		} else {
			log.V(0).Error(err, "Unable to calculate latency", "stream", stream)
		}
	}
	if err := plotutil.AddLinePoints(aPlot, data...); err != nil {
		panic(err)
	}

	if err := aPlot.Save(defaultWidth, defaultHeight, path.Join(outDir, "latency.png")); err != nil {
		panic(err)
	}
}

func (p GoNumPlot) exportMemoryPlot(data plotter.XYs) {
	log.V(3).Info("Generating Memory Plot", "data", data)
	aPlot := plot.New()

	aPlot.Title.Text = "Mem"
	aPlot.X.Label.Text = "Time(m)"
	aPlot.Y.Label.Text = "Megabytes"

	if err := plotutil.AddLinePoints(
		aPlot,
		"Mem",
		data,
	); err != nil {
		panic(err)
	}

	if err := aPlot.Save(defaultWidth, defaultHeight, path.Join(p.ArtifactDir, "mem.png")); err != nil {
		panic(err)
	}
}

func (p GoNumPlot) exportCpuPlot(data plotter.XYs) {
	log.V(3).Info("Generating CPU Plot", "data", data)
	aPlot := plot.New()

	aPlot.Title.Text = "CPU"
	aPlot.X.Label.Text = "Time(m)"
	aPlot.Y.Label.Text = "Cores"

	if err := plotutil.AddLinePoints(
		aPlot,
		"CPU",
		data,
	); err != nil {
		panic(err)
	}

	if err := aPlot.Save(defaultWidth, defaultHeight, path.Join(p.ArtifactDir, "cpu.png")); err != nil {
		panic(err)
	}
}

func extractPlotData(samples []stats.Sample) (cpu, memory plotter.XYs) {
	var first int64
	for i, sample := range samples {
		if i == 0 {
			first = sample.Time
		}
		timeSample := float64(sample.Time-first) / 60.0
		cpu = append(cpu, plotter.XY{
			X: timeSample,
			Y: sample.CPUCoresAsFloat(),
		})

		memory = append(memory, plotter.XY{
			X: timeSample,
			Y: sample.MemoryBytesAsFloat() * megaBytesPerByte,
		})
	}
	return cpu, memory
}

func exportLossStatsPlot(outDir string, samples stats.LossStats) {
	log.Info("Exporting message losses")
	aPlot := plot.New()
	aPlot.Title.Text = "Percent Collected Per Stream"
	aPlot.Legend.Top = true
	aPlot.X.Label.Text = "SeqId"
	aPlot.Y.Label.Text = "Per. Collected"
	aPlot.Y.Min = 0

	var data []interface{}
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
		var xys plotter.XYs

		minID := streamStats.MinSeqId

		for expSeqId := streamStats.MinSeqId; expSeqId <= streamStats.MaxSeqId; expSeqId++ {
			seqId := streamStats.Entries[i].SequenceId
			if seqId != expSeqId {
				lostLogs += 1
			} else {
				i += 1 //found entry
			}
			per := float64(i) / float64(expSeqId-minID+1) * 100.0
			xys = append(xys, plotter.XY{X: float64(expSeqId), Y: per})
		}
		log.V(3).Info("Adding loss stats to plot", "xys", xys, "stream", "stream")
		data = append(data, stream, xys)
	}
	if err := plotutil.AddLinePoints(aPlot, data...); err != nil {
		panic(err)
	}
	if err := aPlot.Save(defaultWidth, defaultHeight, path.Join(outDir, "loss.png")); err != nil {
		panic(err)
	}
}
