package atomicbool

import (
	"sync/atomic"
)

type AtomicBool struct {
	value *int32
}

func New() *AtomicBool {
	ab := new(AtomicBool)
	ab.value = new(int32)
	*ab.value = 0
	return ab
}

func bToI(b bool) int32 {
	if b {
		return 1
	}
	return 0
}

func iToB(i int32) bool {
	if i == 1 {
		return true
	}
	return false
}

// Get atomically loads from b
func (b *AtomicBool) Get() bool {
	return iToB(atomic.LoadInt32(b.value))
}

// Set atomically stores v to b
func (b *AtomicBool) Set(v bool) {
	atomic.StoreInt32(b.value, bToI(v))
}

// CompareAndSwap compares b to old and sets b to new if they match
//
// Return is true iff b was set to new
func (b *AtomicBool) CompareAndSwap(old, new bool) bool {
	return atomic.CompareAndSwapInt32(b.value, bToI(old), bToI(new))
}

// Swap atomically swaps b for new
//
// Returns the old value.
func (b *AtomicBool) Swap(new bool) bool {
	return iToB(atomic.SwapInt32(b.value, bToI(new)))
}
