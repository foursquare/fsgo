## AtomicBool

A simple atomic boolean implemented in go ontop of sync/atomic

### Usage

```go
import "github.com/theevocater/go-atomicbool"

// default is false just like bool
b := New()

// set to true
b.Set(true)

// get current value
b.Get()
```

See the docs for examples of more complex usage.
