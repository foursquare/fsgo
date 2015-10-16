package atomicbool

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestbToI(t *testing.T) {
	assert.Equal(t, bToI(true), 1)
	assert.Equal(t, bToI(false), 0)
}

func TestiToB(t *testing.T) {
	assert.Equal(t, iToB(1), true)
	assert.Equal(t, iToB(0), false)
}

func ExampleAtomicBool_Swap() {
	// default is false just like bool
	b := New()

	// set to true
	b.Set(true)

	// swap values
	fmt.Println(b.Swap(false))
	// Output:
	// true
}
func ExampleAtomicBool_CompareAndSwap() {
	b := New()

	// set to true
	b.Set(true)

	// get current value
	fmt.Println(b.Get())

	// compare b to old and set b to new if they match
	// swapped is true iff b was set to the second arg
	fmt.Println(b.CompareAndSwap(true, false))

	// CompareAndSwap does nothing here as b's value is false
	fmt.Println(b.CompareAndSwap(true, false))
	// Output:
	// true
	// true
	// false
}

func TestGet(t *testing.T) {
	b := New()
	assert.Equal(t, b.Get(), false)
}

func TestSet(t *testing.T) {
	b := New()
	b.Set(true)
	assert.Equal(t, b.Get(), true)
	b.Set(false)
	assert.Equal(t, b.Get(), false)
}

func TestCompareAndSwap(t *testing.T) {
	b := New()
	var ret bool
	ret = b.CompareAndSwap(true, false)
	// b starts false, so the cas should not change the value
	assert.Equal(t, b.Get(), false)
	assert.Equal(t, ret, false)

	ret = b.CompareAndSwap(false, true)
	// b should have been set to true and so ret should be true
	assert.Equal(t, b.Get(), true)
	assert.Equal(t, ret, true)
}

func TestSwap(t *testing.T) {
	b := New()
	var ret bool
	ret = b.Swap(true)
	// b starts false so it should now be true and ret false
	assert.Equal(t, b.Get(), true)
	assert.Equal(t, ret, false)

	// we set b to true above, so it should stay true
	ret = b.Swap(true)
	assert.Equal(t, b.Get(), true)
	assert.Equal(t, ret, true)
}
