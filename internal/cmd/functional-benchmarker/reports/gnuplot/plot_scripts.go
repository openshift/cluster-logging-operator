package gnuplot

const (
	cpuPlotPNG = `
set term png size 1024,768
set output 'cpu.png'
set timefmt '%s'
set xdata time
set title 'CPU(Cores)'
set xlabel 'Time'
plot 'cpu.data' using 1:2 with lines
`
	memPlotPNG = `
set term png size 1024,768
set output 'mem.png'
set timefmt '%s'
set xdata time
set title 'Mem(Mb)'
set xlabel 'Time'
plot 'mem.data' using 1:2 with lines
`
	cpuPlotDumb = `
set term dumb
set timefmt '%s'
set xdata time
set title 'CPU(Cores)'
set xlabel 'Time'
plot 'cpu.data' using 1:2 with lines
`
	memPlotDumb = `
set term dumb
set timefmt '%s'
set xdata time
set title 'Mem(Mb)'
set xlabel 'Time'
plot 'mem.data' using 1:2 with lines
`
	html = `
<html>
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
</html>
`
)
