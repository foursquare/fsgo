package httpthrift

import (
	"io"
	"net/http"

	"github.com/apache/thrift/lib/go/thrift"
)

type sendProt struct {
	transport *http.Client
	url       func() string
	sendbuf   *thrift.TMemoryBuffer
	recvbuf   *thrift.TMemoryBuffer

	thrift.TProtocol
}

func (t *sendProt) Flush() error {
	req, err := http.NewRequest("POST", t.url(), t.sendbuf)
	req.Header.Set("Content-Length", string(t.sendbuf.Len()))
	req.Header.Set("Content-Type", "application/x-thrift")

	resp, err := t.transport.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	io.Copy(t.recvbuf, resp.Body)
	return nil
}

func getSendProt(url func() string, recvbuf *thrift.TMemoryBuffer, compact bool) thrift.TProtocol {
	sendbuf := thrift.NewTMemoryBuffer()
	var underlying thrift.TProtocol
	if compact {
		underlying = thrift.NewTCompactProtocol(sendbuf)
	} else {
		underlying = thrift.NewTBinaryProtocol(sendbuf, true, true)
	}
	return &sendProt{&http.Client{Transport: &http.Transport{}}, url, sendbuf, recvbuf, underlying}
}

func NewDynamicClientProts(url func() string, compact bool) (recv, send thrift.TProtocol) {
	recvbuf := thrift.NewTMemoryBuffer()
	send = getSendProt(url, recvbuf, compact)
	if compact {
		recv = thrift.NewTCompactProtocol(recvbuf)
	} else {
		recv = thrift.NewTBinaryProtocol(recvbuf, true, true)
	}
	return recv, send
}

// pass these to the generated `NewFooClientProtocol(nil, recv, send)` method.
func NewClientProts(url string, compact bool) (recv, send thrift.TProtocol) {
	return NewDynamicClientProts(func() string { return url }, compact)
}
