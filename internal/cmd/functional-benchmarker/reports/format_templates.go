package reports

const (
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
<table border="1">
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
  <br/>
  <br/>
  <div>
    <img src="mem.png">
  </div>
  <br/>
  <br/>
  <div>
    <img src="latency.png">
  </div>
  <br/>
  <br/>
  <div>
    <img src="loss.png">
  </div>
  <br/>
  <br/>
  <div>
	<table border="1">
	  <tr>
		<th>Stream</th>
		<th>Min Seq</th>
		<th>Max Seq</th>
		<th>Purged</th>
		<th>Collected</th>
		<th>Percent Collected</th>
	  </tr>
	  <tr>
      %s
    </table>
  </div>
  <div>
    <code style="display:block;white-space:pre-wrap">
    %s
	</code>
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
%d|%d|%s|%.3f|%.3f|%.3f|%.3f

![](cpu.png)

![](mem.png)

![](latency.png)

![](loss.png)

## Percent logs lost between first and last collected sequence ids
Stream |  Min Seq | Max Seq | Purged | Collected | Percent Collected |
-------| ---------| --------| -------|-----------|--------------|
%s

## Config

%s

`
)
