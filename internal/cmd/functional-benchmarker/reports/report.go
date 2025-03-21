package reports

import (
	"fmt"
	log "github.com/ViaQ/logerr/v2/log/static"
	"github.com/openshift/cluster-logging-operator/internal/cmd/functional-benchmarker/config"
	"github.com/openshift/cluster-logging-operator/internal/cmd/functional-benchmarker/reports/gonumplot"
	"github.com/openshift/cluster-logging-operator/internal/cmd/functional-benchmarker/stats"
	htmllib "html"
	"os"
	"path"
	"time"
)

type Report interface {
	Generate()
}

type Plotter interface {
	Plot()
}

func NewReporter(options config.Options, artifactDir string, metrics *stats.ResourceMetrics, statistics *stats.Statistics) Report {
	return &Reporter{
		Options:     options,
		Stats:       *statistics,
		ArtifactDir: artifactDir,
		Plotter:     NewPlotter(artifactDir, metrics, statistics),
	}
}

func NewPlotter(artifactDir string, metrics *stats.ResourceMetrics, statistics *stats.Statistics) Plotter {
	return &gonumplot.GoNumPlot{
		Metrics:     *metrics,
		Stats:       *statistics,
		ArtifactDir: artifactDir,
	}
}

type Reporter struct {
	Options     config.Options
	Stats       stats.Statistics
	Plotter     Plotter
	ArtifactDir string
}

func (r *Reporter) Generate() {
	r.Plotter.Plot()
	r.generateReports()
}

/* #nosec G306*/
func (r *Reporter) generateReports() {
	log.Info("Generating stats reports")
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
		err := os.WriteFile(path.Join(r.ArtifactDir, file), []byte(out), 0755)
		if err != nil {
			log.Error(err, "Error writing file")
		}
	}
}
