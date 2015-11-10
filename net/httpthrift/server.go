package httpthrift

import (
	"net/http"
	"sync"
	"time"

	"github.com/apache/thrift/lib/go/thrift"
	"github.com/foursquare/fsgo/report"
)

// Thrift's generated Processors have `GetProcessorFunction` and satisfy this interface.
type HasProcessFunc interface {
	GetProcessorFunction(key string) (processor thrift.TProcessorFunction, ok bool)
}

// Wraps a generated thrift Processor, providing a ServeHTTP method to serve thrift-over-http.
type ThriftOverHTTPHandler struct {
	stats   *report.Recorder
	buffers sync.Pool
	HasProcessFunc
}

func NewThriftOverHTTPHandler(p HasProcessFunc, stats *report.Recorder) *ThriftOverHTTPHandler {
	return &ThriftOverHTTPHandler{stats, sync.Pool{}, p}
}

// Mostly borrowed from generated thrift code `Process` method, but with timing added.
func (p ThriftOverHTTPHandler) handle(iprot, oprot thrift.TProtocol) (success bool, err thrift.TException) {
	name, _, seqId, err := iprot.ReadMessageBegin()
	if err != nil {
		return false, err
	}

	if processor, ok := p.GetProcessorFunction(name); ok {
		start := time.Now()
		success, err = processor.Process(seqId, iprot, oprot)
		if p.stats != nil {
			if err != nil {
				p.stats.Inc("rpc.error." + name)
			}
			dur := time.Now().Sub(start)
			p.stats.Time("rpc.timing._all_", dur)
			p.stats.Time("rpc.timing."+name, dur)
		}
		return success, err
	}

	iprot.Skip(thrift.STRUCT)
	iprot.ReadMessageEnd()
	e := thrift.NewTApplicationException(thrift.UNKNOWN_METHOD, "Unknown function "+name)

	if p.stats != nil {
		p.stats.Inc("rpc.error.unknown_function." + name)
	}

	oprot.WriteMessageBegin(name, thrift.EXCEPTION, seqId)
	e.Write(oprot)
	oprot.WriteMessageEnd()
	oprot.Flush()

	return false, e
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

		ok, err := h.handle(iprot, oprot)

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
