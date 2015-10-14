# Metrics Reporting Helper

A thin wrapper around [go-metrics](github.com/rcrowley/go-metrics) that streamlines setup and a few common uses (tracking event rates and timings) and exports to graphite.

## Usage

```
  r := report.NewRecorder().ReportTo("graphite-collector:2170", "foobar.baz")
  r.Inc("request")
  r.Time("handler", 5*time.Second)
```
### Default Recorder
Alternately, rather than pass around a Recorder instance, you can also set a configured `Recorder` as the default during startup, then use helpers elsewhere that just report to the default recorder.

```
  report.NewRecorder().ReportTo("graphite-collector:2170", "foobar.baz").SetAsDefault()
  ...
  report.Inc("request")
  report.Time("handler", 5*time.Second)
```
  
# Authors
- [David Taylor](http://github.com/dt)

