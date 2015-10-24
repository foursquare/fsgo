package discovery

import (
	"math/rand"
	"sync/atomic"
	"time"
)

type InstanceProvider interface {
	/**
	 * Return the current available set of instances <b>IMPORTANT: </b> users
	 * should not hold on to the instance returned. They should always get a fresh list.
	 */
	GetAllInstances() ([]*ServiceInstance, error)
}

type ProviderStrategy interface {
	// Given a source of instances, return one of them for a single use.
	GetInstance(instanceProvider InstanceProvider) (*ServiceInstance, error)
}

type FixedSetInstanceProvider struct {
	instances []*ServiceInstance
}

func (p *FixedSetInstanceProvider) GetAllInstances() ([]*ServiceInstance, error) {
	return p.instances, nil
}

// Ensure FixedSetInstanceProvider implements InstanceProvider
var _ InstanceProvider = (*FixedSetInstanceProvider)(nil)

type ServiceProvider interface {
	InstanceProvider
	/**
	 * Return an instance for a single use. <b>IMPORTANT: </b> users
	 * should not hold on to the instance returned. They should always get a fresh instance.
	 */
	GetInstance() (*ServiceInstance, error)
}

type RandomProvider struct {
	*rand.Rand
}

func NewRandomProvider() *RandomProvider {
	return &RandomProvider{rand.New(rand.NewSource(time.Now().UnixNano()))}
}

// Ensure RandomProvider implements ProviderStrategy
var _ ProviderStrategy = (*RandomProvider)(nil)

func (r RandomProvider) GetInstance(ip InstanceProvider) (*ServiceInstance, error) {
	instances, err := ip.GetAllInstances()
	if err != nil {
		return nil, err
	}
	if len(instances) < 1 {
		return nil, nil
	}

	return instances[r.Intn(len(instances))], nil
}

type RoundRobinProvider struct {
	index uint64
}

func NewRoundRobinProvider() *RoundRobinProvider {
	return &RoundRobinProvider{0}
}

// Ensure RoundRobinProvider implements ProviderStrategy
var _ ProviderStrategy = (*RoundRobinProvider)(nil)

func (r *RoundRobinProvider) GetInstance(ip InstanceProvider) (*ServiceInstance, error) {
	instances, err := ip.GetAllInstances()
	if err != nil {
		return nil, err
	}
	if len(instances) < 1 {
		return nil, nil
	}
	res := atomic.AddUint64(&r.index, 1)
	return instances[res%uint64(len(instances))], nil
}
