package thriftrpc

import (
	"log"
	"time"

	"github.com/apache/thrift/lib/go/thrift"
	"github.com/foursquare/fsgo/report"
)

// Thrift's generated Processors have `GetProcessorFunction` and satisfy this interface.
type HasProcessFunc interface {
	GetProcessorFunction(key string) (processor thrift.TProcessorFunction, ok bool)
}

// Wraps a generated thrift Processor, providing a ServeHTTP method to serve thrift-over-http.
type LoggedProcessor struct {
	HasProcessFunc
	stats *report.Recorder
	debug bool
}

func AddLogging(p HasProcessFunc, stats *report.Recorder, debug bool) thrift.TProcessor {
	return LoggedProcessor{p, stats, debug}
}

// Mostly borrowed from generated thrift code `Process` method, but with timing added.
func (p LoggedProcessor) Process(iprot, oprot thrift.TProtocol) (success bool, err thrift.TException) {
	name, _, seqId, err := iprot.ReadMessageBegin()
	if err != nil {
		return false, err
	}

	if processor, ok := p.GetProcessorFunction(name); ok {
		if p.debug {
			log.Println("[rpc]", name)
		}
		start := time.Now()
		success, err = processor.Process(seqId, iprot, oprot)
		if p.stats != nil {
			if err != nil {
				if p.debug {
					log.Println("[rpc]", name, err)
				}

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

	if p.debug {
		log.Println("[rpc] unknown function:", name)
	}

	if p.stats != nil {
		p.stats.Inc("rpc.error.unknown_function." + name)
	}

	oprot.WriteMessageBegin(name, thrift.EXCEPTION, seqId)
	e.Write(oprot)
	oprot.WriteMessageEnd()
	oprot.Flush()

	return true, e
}
