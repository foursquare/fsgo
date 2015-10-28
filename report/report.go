package report

import (
	"flag"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/foursquare/go-metrics"
)

type Meter interface {
	metrics.Meter
}

type Guage interface {
	metrics.Gauge
}

type Timer interface {
	metrics.Timer
}

func Flag() *string {
	return flag.String("graphite", "", "graphite server/prefix for reporting collected metrics")
}

type Recorder struct {
	metrics.Registry
	isExporting  bool
	Format       ExportFormatStrings
	DurationUnit time.Duration // Time conversion unit for durations
	Prefix       string        // Prefix to be prepended to metric names
	Percentiles  []float64     // Percentiles to export from timers and histograms
}

func NewRecorder() *Recorder {
	return &Recorder{
		metrics.NewRegistry(),
		false,
		OstrichFormats,
		time.Millisecond,
		"",
		[]float64{0.5, 0.9, 0.95, 0.99, 0.999},
	}
}

func (r *Recorder) GetGuage(name string) Guage {
	return metrics.GetOrRegisterGauge(name, r)
}

type ClearableTimer struct {
	metrics.Timer
	h metrics.Histogram
}

func (c *ClearableTimer) Clear() {
	c.h.Clear()
}

func (r *Recorder) makeTimer() metrics.Timer {
	if r.isExporting {
		h := metrics.NewHistogram(metrics.NewUniformSample(1000 * 30))
		t := metrics.NewCustomTimer(h, metrics.NewMeter())
		return &ClearableTimer{t, h}
	} else {
		return metrics.NewTimer()
	}
}

func (r *Recorder) GetTimer(name string) Timer {
	return r.GetOrRegister(name, r.makeTimer).(Timer)
}

func (r *Recorder) GetMeter(name string) Meter {
	return metrics.GetOrRegisterMeter(name, r)
}

func (r *Recorder) Inc(name string) {
	r.GetMeter(name).Mark(1)
}

func (r *Recorder) Time(name string, du time.Duration) {
	r.GetTimer(name).Update(du)
}

func (r *Recorder) TimeSince(name string, t time.Time) {
	r.GetTimer(name).UpdateSince(t)
}

func (r *Recorder) LogToConsole(freq time.Duration) *Recorder {
	log.Println("Stats reporting enabled...")
	go metrics.LogScaled(r, freq, time.Millisecond, log.New(os.Stderr, "metrics: ", log.Lmicroseconds))
	return r
}

func (r *Recorder) EnableGCInfoCollection() *Recorder {
	metrics.RegisterDebugGCStats(r)
	go metrics.CaptureRuntimeMemStats(r, 15*time.Second)
	metrics.RegisterRuntimeMemStats(r)
	go metrics.CaptureDebugGCStats(r, 15*time.Second)
	return r
}

func (r *Recorder) MaybeReportTo(serverSlashPrefix *string) *Recorder {
	if serverSlashPrefix == nil || len(*serverSlashPrefix) < 1 {
		return r
	}
	return r.ReportTo(*serverSlashPrefix)
}

func (r *Recorder) ReportTo(serverSlashPrefix string) *Recorder {
	parts := strings.Split(serverSlashPrefix, "/")
	if len(parts) != 2 || len(parts[0]) < 1 || len(parts[1]) < 1 {
		panic("bad graphite server and prefix format. must be server/prefix.path (both non-empty).")
	}
	return r.ReportToServer(parts[0], parts[1])
}

func (r *Recorder) ReportToServer(graphiteServer, graphitePrefix string) *Recorder {
	log.Printf("Stats reporting to graphite server '%s' under '%s'...\n", graphiteServer, graphitePrefix)
	addr, err := net.ResolveTCPAddr("tcp", graphiteServer)
	if err != nil {
		panic(err)
	}
	r.Prefix = graphitePrefix

	cfg := &GraphiteConfig{
		Addr:          addr,
		FlushInterval: 1 * time.Minute,
	}
	r.isExporting = true

	go exporter(r, cfg)
	return r
}

func (r *Recorder) RegisterHttp() *Recorder {
	http.Handle("/statz", r)
	return r
}

var defaultRecorder *Recorder

func (r *Recorder) SetAsDefault() *Recorder {
	defaultRecorder = r
	return r
}

func GetDefault() *Recorder {
	if defaultRecorder == nil {
		panic("Must call SetAsDefault() on a Recorder instace before using the default stats collector.")
	}
	return defaultRecorder
}

func Inc(name string) {
	metrics.GetOrRegisterMeter(name, GetDefault()).Mark(1)
}

func Time(name string, du time.Duration) {
	metrics.GetOrRegisterTimer(name, GetDefault()).Update(du)
}

func TimeSince(name string, t time.Time) {
	metrics.GetOrRegisterTimer(name, GetDefault()).UpdateSince(t)
}

func (r *Recorder) ServeHTTP(out http.ResponseWriter, req *http.Request) {
	out.Header().Add("Content-Type", "text/plain")
	writeStats(r, out, true)
}
