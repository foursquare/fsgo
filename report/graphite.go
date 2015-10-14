package report

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/foursquare/go-metrics"
)

type GraphiteConfig struct {
	Addr          *net.TCPAddr // Network address to connect to
	Format        ExportFormatStrings
	FlushInterval time.Duration // Flush interval
	DurationUnit  time.Duration // Time conversion unit for durations
	Prefix        string        // Prefix to be prepended to metric names
	Percentiles   []float64     // Percentiles to export from timers and histograms
}

func exporter(r metrics.Registry, c *GraphiteConfig) {
	for _ = range time.Tick(c.FlushInterval) {
		if err := sendToGraphite(r, c); nil != err {
			log.Println(err)
		}
	}
}

func sendToGraphite(r metrics.Registry, c *GraphiteConfig) error {
	now := time.Now().Unix()
	du := float64(c.DurationUnit)
	conn, err := net.DialTCP("tcp", nil, c.Addr)
	if nil != err {
		return err
	}
	defer conn.Close()
	w := bufio.NewWriter(conn)
	r.Each(func(name string, i interface{}) {
		switch metric := i.(type) {
		case metrics.Counter:
			fmt.Fprintf(w, c.Format.Counter, c.Prefix, name, metric.Count(), now)
		case metrics.Gauge:
			fmt.Fprintf(w, c.Format.Gauge, c.Prefix, name, metric.Value(), now)
		case metrics.GaugeFloat64:
			fmt.Fprintf(w, c.Format.GaugeFloat64, c.Prefix, name, metric.Value(), now)
		case metrics.Histogram:
			h := metric.Snapshot()
			ps := h.Percentiles(c.Percentiles)
			fmt.Fprintf(w, c.Format.HistogramCount, c.Prefix, name, h.Count(), now)
			fmt.Fprintf(w, c.Format.Min, c.Prefix, name, h.Min(), now)
			fmt.Fprintf(w, c.Format.Max, c.Prefix, name, h.Max(), now)
			fmt.Fprintf(w, c.Format.Mean, c.Prefix, name, h.Mean(), now)
			fmt.Fprintf(w, c.Format.Stddev, c.Prefix, name, h.StdDev(), now)
			for psIdx, psKey := range c.Percentiles {
				key := strings.Replace(strconv.FormatFloat(psKey*100.0, 'f', -1, 64), ".", "", 1)
				fmt.Fprintf(w, c.Format.Percentile, c.Prefix, name, key, ps[psIdx], now)
			}
		case metrics.Meter:
			m := metric.Snapshot()
			fmt.Fprintf(w, c.Format.HistogramCount, c.Prefix, name, m.Count(), now)
			fmt.Fprintf(w, c.Format.Rate1, c.Prefix, name, m.Rate1(), now)
			fmt.Fprintf(w, c.Format.Rate5, c.Prefix, name, m.Rate5(), now)
			fmt.Fprintf(w, c.Format.Rate15, c.Prefix, name, m.Rate15(), now)
			fmt.Fprintf(w, c.Format.Mean, c.Prefix, name, m.RateMean(), now)
		case metrics.Timer:
			t := metric.Snapshot()
			ps := t.Percentiles(c.Percentiles)
			fmt.Fprintf(w, c.Format.HistogramCount, c.Prefix, name, t.Count(), now)
			fmt.Fprintf(w, c.Format.Min, c.Prefix, name, t.Min()/int64(du), now)
			fmt.Fprintf(w, c.Format.Max, c.Prefix, name, t.Max()/int64(du), now)
			fmt.Fprintf(w, c.Format.Mean, c.Prefix, name, t.Mean()/du, now)
			fmt.Fprintf(w, c.Format.Stddev, c.Prefix, name, t.StdDev()/du, now)
			for psIdx, psKey := range c.Percentiles {
				key := strings.Replace(strconv.FormatFloat(psKey*100.0, 'f', -1, 64), ".", "", 1)
				fmt.Fprintf(w, c.Format.Percentile, c.Prefix, name, key, ps[psIdx]/du, now)
			}
			fmt.Fprintf(w, c.Format.Rate1, c.Prefix, name, t.Rate1(), now)
			fmt.Fprintf(w, c.Format.Rate5, c.Prefix, name, t.Rate5(), now)
			fmt.Fprintf(w, c.Format.Rate15, c.Prefix, name, t.Rate15(), now)
			fmt.Fprintf(w, c.Format.Mean, c.Prefix, name, t.RateMean(), now)
		default:
			log.Printf("Cannot export unknown metric type %T for '%s'\n", i, name)
		}
		w.Flush()
	})
	return nil
}
