package main

import (
	"fmt"
	"os"
	"text/tabwriter"
)

type Report interface {
	Print()
	Add(name string, stats Statistics, metrics Metrics)
}

func NewReporter(outputType string) Report {
	base := BaseReporter{
		stats:   map[string]Statistics{},
		metrics: map[string]Metrics{},
	}
	if outputType == "csv" {
		return &CSVReporter{
			base,
		}
	}
	return &TableReporter{
		base,
	}
}

type BaseReporter struct {
	stats   map[string]Statistics
	metrics map[string]Metrics
}

type TableReporter struct {
	BaseReporter
}

func (r *TableReporter) Add(name string, stats Statistics, metrics Metrics) {
	r.stats[name] = stats
	r.metrics[name] = metrics
}
func (r *TableReporter) Print() {
	w := new(tabwriter.Writer)
	// minwidth, tabwidth, padding, padchar, flags
	w.Init(os.Stdout, 3, 2, 2, ' ', tabwriter.AlignRight)

	defer w.Flush()
	div := "--------"

	fmt.Fprintf(w, "\n %s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t", "Run", "Total", "Size", "cpu.user", "cpu.kernel", "mem.virtual.peak", "Elapsed", "Mean", "Min", "Max", "Median", "Mean")
	fmt.Fprintf(w, "\n %s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t", "", "Msg", "(bytes)", "(ticks)", "(ticks)", "(KB)", "(s)", "(s)", "(s)", "(s)", "(s)", "Bloat")
	fmt.Fprintf(w, "\n %s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t", div, div, div, div, div, div, div, div, div, div, div, div)

	for name, s := range r.stats {
		m := r.metrics[name]
		fmt.Fprintf(w, "\n %s\t%d\t%d\t%s\t%s\t%s\t%.3f\t%.3f\t%.3f\t%.3f\t%.3f\t%.3f\t",
			name,
			s.TotMessages(),
			s.msgSize,
			m.cpuUserTicks,
			m.cpuKernelTicks,
			m.memVirtualPeakKB,
			s.elapsed,
			s.mean(),
			s.min(),
			s.max(),
			s.median(),
			s.meanBloat())
	}

	fmt.Fprintf(w, "\n")
}

type CSVReporter struct {
	BaseReporter
}

func (r *CSVReporter) Add(name string, stats Statistics, metrics Metrics) {
	r.stats[name] = stats
	r.metrics[name] = metrics
}
func (r *CSVReporter) Print() {
	w := new(tabwriter.Writer)
	// minwidth, tabwidth, padding, padchar, flags
	w.Init(os.Stdout, 3, 2, 2, ' ', tabwriter.AlignRight)

	defer w.Flush()

	fmt.Fprintf(w, "\n %s,%s,%s,%s,%s,%s,%s,%s,%s,%s,%s,%s", "Run", "Total", "Size", "cpu.user", "cpu.kernel", "mem.virtual.peak", "Elapsed", "Mean", "Min", "Max", "Median", "Mean")
	fmt.Fprintf(w, "\n %s,%s,%s,%s,%s,%s,%s,%s,%s,%s,%s,%s", "", "Msg", "(bytes)", "(ticks)", "(ticks)", "(KB)", "(s)", "(s)", "(s)", "(s)", "(s)", "Bloat")

	for name, s := range r.stats {
		m := r.metrics[name]
		fmt.Fprintf(w, "\n %s,%d,%d,%s,%s,%s,%f,%f,%f,%f,%f,%f",
			name,
			s.TotMessages(),
			s.msgSize,
			m.cpuUserTicks,
			m.cpuKernelTicks,
			m.memVirtualPeakKB,
			s.elapsed,
			s.mean(),
			s.min(),
			s.max(),
			s.median(),
			s.meanBloat())
	}

	fmt.Fprintf(w, "\n")
}
