package gnuplot

const (
	cpuPlotPNG = `set term png size 1024,768
set output 'cpu.png'
set timefmt '%s'
set xdata time
set title 'CPU(Cores)'
set xlabel 'Time'
plot 'cpu.data' using 1:2 with lines`

	memPlotPNG = `set term png size 1024,768
set output 'mem.png'
set timefmt '%s'
set xdata time
set title 'Mem(Mb)'
set xlabel 'Time'
plot 'mem.data' using 1:2 with lines`

	latencyPlotPNG = `set term png size 1024,768
set output 'latency.png'
set title 'Latency(s)'
set xlabel 'Message'
f(x)=m*x+b
fit f(x) 'latency.data' using 1:2 via m,b
plot 'latency.data' using 1:2 with lines title 'Data', f(x) title 'Trend'`

	lossPlotPNG = `set term png size 1024,768
set output 'loss.png';
set xlabel 'SeqId';
set ylabel 'Lost Count'`

	lossPlotDumb = `set term dumb
set xlabel 'SeqId';
set ylabel 'Lost Count'`

	cpuPlotDumb = `set term dumb
set timefmt '%s'
set xdata time
set title 'CPU(Cores)'
set xlabel 'Time'
plot 'cpu.data' using 1:2 with lines`

	memPlotDumb = `set term dumb
set timefmt '%s'
set xdata time
set title 'Mem(Mb)'
set xlabel 'Time'
plot 'mem.data' using 1:2 with lines`

	latencyPlotDumb = `set term dumb
set title 'Latency(s)'
set xlabel 'Message'
plot 'latency.data' using 1:2 with lines`

	html = `
<html>
<div>
  <div><b>Options</b><div>
  <div>Image: %s</div>
  <div>Total Log Stressors: %d</div>
  <div>Lines Per Second: %d</div>
  <div>Run Duration: %s</div>
  <div>Payload Source: %s</div>
</div>
<div>
  Latency of logs collected based on the time the log was generated and ingested
</div>
<table>
  <tr>
    <th>Total</th>
    <th>Size</th>
    <th>Elapsed</th>
    <th>Mean</th>
    <th>Min</th>
    <th>Max</th>
    <th>Median</th>
  </tr>
  <tr>
    <th>Msg</th>
    <th></th>
    <th>(s)</th>
    <th>(s)</th>
    <th>(s)</th>
    <th>(s)</th>
    <th>(s)</th>
  </tr>
  <tr>
   <td>%d</td>
   <td>%d</td>
   <td>%s</td>
   <td>%.3f</td>
   <td>%.3f</td>
   <td>%.3f</td>
   <td>%.3f</td>
  </tr>
</table>
  <div>
    <img src="cpu.png">
  </div>
  <div>
    <img src="mem.png">
  </div>
  <div>
    <img src="latency.png">
  </div>
  <div>
    <img src="loss.png">
  </div>
  <div>
	<table>
	  <tr>
		<th>Stream</th>
		<th>Percent Lost</th>
	  </tr>
	  <tr>
      %s
  </div>
  <div>
    %s
  </div>
</html>
`
	markdown = `
# collector Functionl Benchmark Results
## Options
* Image: %s
* Total Log Stressors: %d
* Lines Per Second: %d
* Run Duration: %s
* Payload Source: %s

## Latency of logs collected based on the time the log was generated and ingested

Total Msg| Size | Elapsed (s) | Mean (s)| Min(s) | Max (s)| Median (s)
---------|------|-------------|---------|--------|--------|---
%d|%d|%s|%.3f|%.3f|%.3f%.3f

![](cpu.png)

![](mem.png)

![](latency.png)

![](loss.png)

## Percent logs lost between first and last collected sequence ids
Stream | Percent Lost
-------| ------------
%s

## Config
<code style="white-space:pre;">
%s
</code>
`
)
