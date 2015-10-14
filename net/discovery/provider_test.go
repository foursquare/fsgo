package discovery

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func dummyInstances(count int) InstanceProvider {
	acc := make([]*ServiceInstance, count)
	for i := 0; i < count; i++ {
		acc[i] = NewSimpleServiceInstance(fmt.Sprintf("fake-%d", i), "", i)
	}
	return &FixedSetInstanceProvider{acc}
}

func TestRoundRobinProvider(t *testing.T) {
	p := &RoundRobinProvider{0}
	count := 10
	ip := dummyInstances(count)

	first, err := p.GetInstance(ip)
	assert.Nil(t, err)

	second, err := p.GetInstance(ip)
	assert.Equal(t, *first.Port+1, *second.Port)

	thrid, err := p.GetInstance(ip)
	assert.Equal(t, *first.Port+2, *thrid.Port)

	// make 100 calls per instance, then check how many times we got each instance.
	seen := make(map[int]int)
	for i := 0; i < count*100; i++ {
		x, err := p.GetInstance(ip)
		assert.Nil(t, err)
		seen[*x.Port] = seen[*x.Port] + 1
	}

	for i := 0; i < count; i++ {
		assert.Equal(t, 100, seen[i])
	}
}

func TestRandomProvider(t *testing.T) {
	p := NewRandomProvider()
	count := 10
	ip := dummyInstances(count)

	seen := make(map[int]int)
	for i := 0; i < count*10000; i++ {
		x, err := p.GetInstance(ip)
		assert.Nil(t, err)
		seen[*x.Port] = seen[*x.Port] + 1
	}

	min := int(^uint(0) >> 1)
	max := 0
	for i := 0; i < count; i++ {
		if seen[i] < min {
			min = seen[i]
		}
		if seen[i] > max {
			max = seen[i]
		}
	}
	diff := float64(min) / float64(max)
	if diff < .95 {
		t.Fatal(t, "bad distribution? min, max, min/max:", min, max, diff)
	}
}
