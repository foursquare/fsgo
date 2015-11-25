package thriftrpc

import (
	"net/http"
	"sync"
	"time"

	"github.com/apache/thrift/lib/go/thrift"
	"github.com/foursquare/fsgo/report"
)

type ThriftOverHTTPHandler struct {
	thrift.TProcessor
	stats   *report.Recorder
	buffers sync.Pool
}

func NewThriftOverHTTPHandler(p thrift.TProcessor, stats *report.Recorder) *ThriftOverHTTPHandler {
	return &ThriftOverHTTPHandler{p, stats, sync.Pool{}}
}

func (h *ThriftOverHTTPHandler) getBuf() *thrift.TMemoryBuffer {
	res := h.buffers.Get()
	if res == nil {
		return thrift.NewTMemoryBuffer()
	} else {
		out := res.(*thrift.TMemoryBuffer)
		out.Reset()
		return out
	}
}

func (h *ThriftOverHTTPHandler) ServeHTTP(out http.ResponseWriter, req *http.Request) {
	start := time.Now()
	if req.Method == "POST" {
		inbuf := h.getBuf()
		defer h.buffers.Put(inbuf)
		outbuf := h.getBuf()
		defer h.buffers.Put(outbuf)

		inbuf.ReadFrom(req.Body)
		defer req.Body.Close()

		compact := false

		if inbuf.Len() > 0 && inbuf.Bytes()[0] == thrift.COMPACT_PROTOCOL_ID {
			compact = true
		}

		var iprot thrift.TProtocol
		var oprot thrift.TProtocol

		if compact {
			iprot = thrift.NewTCompactProtocol(inbuf)
			oprot = thrift.NewTCompactProtocol(outbuf)
		} else {
			iprot = thrift.NewTBinaryProtocol(inbuf, true, true)
			oprot = thrift.NewTBinaryProtocol(outbuf, true, true)
		}

		ok, err := h.Process(iprot, oprot)

		if ok {
			outbuf.WriteTo(out)
		} else {
			http.Error(out, err.Error(), 500)
		}
	} else {
		http.Error(out, "Must POST TBinary encoded thrift RPC", 401)
	}
	if h.stats != nil {
		h.stats.TimeSince("servehttp", start)
	}
}
