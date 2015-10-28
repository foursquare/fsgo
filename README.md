# Foursquare Common Go Libraries
[![Build Status](https://api.travis-ci.org/foursquare/fsgo.svg)](https://travis- ci.org/foursquare/fsgo) [![Coverage Status](https://coveralls.io/repos/foursquare/fsgo/badge.svg?branch=master&service=github)](https://coveralls.io/github/foursquare/fsgo?branch=master)

A collection of reusable libraries and tools for building webservices in Go.

- [atomicbool](./concurrent/atomicbool) atomic boolean
- [discovery](./net/discovery) curator-like service discovery
- [httpthrift](./net/httpthrift) thrift-over-http rpc
- [report](./report) instrumentation and reporting

## Contributing

### Go Version and PATH
fsgo libraries are tested and developed assuming Go 1.5 and `$GOPATH/bin` is on PATH.

_Foursquare engineers_: you can add [this](https://github.com/dt/shell/blob/master/lang.d/go.sh) to your `bashrc`.

### GoImports
fsgo uses `goimports` formatting (a superset of `go fmt` rules, including grouping imports):

* Install `goimports`: `go get golang.org/x/tools/cmd/goimports`
* Fix up files: `goimports -w *.go`
