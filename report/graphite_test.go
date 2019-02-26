package report

import (
	"bufio"
	"net"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/rcrowley/go-metrics"
)

func floatEquals(a, b float64) bool {
	return (a-b) < 0.000001 && (b-a) < 0.000001
}

func NewTestServer(t *testing.T, prefix string) (map[string]float64, net.Listener, *Recorder, *sync.WaitGroup) {
	res := make(map[string]float64)

	ln, err := net.Listen("tcp", ":0")
	if err != nil {
		t.Fatal("could not start dummy server:", err)
	}

	var wg sync.WaitGroup
	t.Log("test")
	go func() {
		t.Log("test")
		for {
			conn, err := ln.Accept()
			if err != nil {
				t.Log("listener failed: ", err)
				return
			}
			r := bufio.NewReader(conn)
			line, err := r.ReadString('\n')
			for err == nil {
				parts := strings.Split(line, " ")
				i, _ := strconv.ParseFloat(parts[1], 0)
				if testing.Verbose() {
					t.Log("recv", parts[0], i)
				}
				res[parts[0]] = res[parts[0]] + i
				line, err = r.ReadString('\n')
			}
			wg.Done()
			conn.Close()
		}
	}()

	r := NewRecorder()
	r.Prefix = prefix
	r.graphite = ln.Addr().(*net.TCPAddr)

	return res, ln, r, &wg
}

type DummyMeter struct {
	count int64
	rate1 float64
	metrics.Meter
}

func (m DummyMeter) Count() int64            { return m.count }
func (m DummyMeter) Rate1() float64          { return m.rate1 }
func (m DummyMeter) Snapshot() metrics.Meter { return m }

var _ (metrics.Meter) = (*DummyMeter)(nil)

func fillMetrics(r *Recorder) {
	r.Register("bar", DummyMeter{40, 4.0, metrics.NilMeter{}})
	r.Time("baz", time.Second*5)
	r.Time("baz", time.Second*4)
	r.Time("baz", time.Second*3)
	r.Time("baz", time.Second*2)
	r.Time("baz", time.Second*1)
}

// TODO: test is producing deadlock
func IgnoreTestGoMetricsWrites(t *testing.T) {
	res, l, r, wg := NewTestServer(t, "foobar")
	defer l.Close()

	metrics.GetOrRegisterCounter("foo", r).Inc(2)

	fillMetrics(r)

	wg.Add(1)
	r.Format = GoMetricsFormats
	if testing.Verbose() {
		t.Log("Sening go-metrics format to graphite..")
	}
	r.sendToGraphite()
	wg.Wait()

	if expected, found := 2.0, res["foobar.foo.count"]; !floatEquals(found, expected) {
		t.Fatal("bad value:", expected, found)
	}

	if expected, found := 40.0, res["foobar.bar.count"]; !floatEquals(found, expected) {
		t.Fatal("bad value:", expected, found)
	}

	if expected, found := 4.0, res["foobar.bar.one-minute"]; !floatEquals(found, expected) {
		t.Fatal("bad value:", expected, found)
	}

	if expected, found := 5.0, res["foobar.baz.count"]; !floatEquals(found, expected) {
		t.Fatal("bad value:", expected, found)
	}

	if expected, found := 5000.0, res["foobar.baz.99-percentile"]; !floatEquals(found, expected) {
		t.Fatal("bad value:", expected, found)
	}

	if expected, found := 3000.0, res["foobar.baz.50-percentile"]; !floatEquals(found, expected) {
		t.Fatal("bad value:", expected, found)
	}
}

// TODO: test is producing deadlock
func IgnoreTestOstrichWrites(t *testing.T) {
	res, l, r, wg := NewTestServer(t, "foobar")
	defer l.Close()

	fillMetrics(r)

	wg.Add(1)
	r.Format = OstrichFormats
	if testing.Verbose() {
		t.Log("Sending ostrich format to graphite..")
	}
	r.sendToGraphite()
	wg.Wait()

	if expected, found := 0.0, res["foobar.baz.99-percentile"]; !floatEquals(found, expected) {
		t.Fatal("bad value:", expected, found)
	}

	if expected, found := 0.0, res["foobar.baz.50-percentile"]; !floatEquals(found, expected) {
		t.Fatal("bad value:", expected, found)
	}

	if expected, found := 5000.0, res["foobar.baz.percentiles.p99"]; !floatEquals(found, expected) {
		t.Fatal("bad value:", expected, found)
	}

	if expected, found := 3000.0, res["foobar.baz.percentiles.p50"]; !floatEquals(found, expected) {
		t.Fatal("bad value:", expected, found)
	}

	for k, _ := range res {
		delete(res, k)
	}

	wg.Add(1)
	if testing.Verbose() {
		t.Log("Sending recently cleared metrics to graphite...")
	}
	r.sendToGraphite()
	wg.Wait()

	if expected, found := 0.0, res["foobar.baz.percentiles.p99"]; !floatEquals(found, expected) {
		t.Fatal("bad value:", expected, found)
	}

	if expected, found := 0.0, res["foobar.baz.percentiles.p50"]; !floatEquals(found, expected) {
		t.Fatal("bad value:", expected, found)
	}
}
