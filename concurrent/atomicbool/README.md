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

// compare b to old and set b to new if they match
// swapped is true iff b was set to the second arg
swapped := b.CompareandSwap(old, new)

// swap values
old := b.Swap(false)
```
