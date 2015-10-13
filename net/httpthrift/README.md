# Thrift-RPC-over-HTTP in Go
[![Build Status](https://api.travis-ci.org/dt/httpthrift.svg)](https://travis-ci.org/dt/httpthrift)

Tiny helper libraries for running Thrift RPC over HTTP -- putting the binary encoded messages in HTTP request/response bodies.

Standard Thrift RPC calls are encoded into byte buffers, which are sent as HTTP request/response bodies. This allows any off-the-shelf http tools (eg HAProxy) to interact with this thrift-RPC traffic.

# Authors
  * [David Taylor](http://github.com/dt) (maintainer)
