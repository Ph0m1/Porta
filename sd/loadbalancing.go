package sd

import (
	"errors"
	"math/rand"
	"sync/atomic"
)

type Balancer interface {
	Host() (string, error)
}

var ErrNoHosts = errors.New("no hosts available")

func NewRoundRobinLB(subscriber Subscriber) Balancer {
	return &roundRobinLB{
		subscriber: subscriber,
		counter:    0,
	}
}

type roundRobinLB struct {
	subscriber Subscriber
	counter    uint64
}

func (rr *roundRobinLB) Host() (string, error) {
	hosts, err := rr.subscriber.Hosts()
	if err != nil {
		return "", err
	}
	if len(hosts) <= 0 {
		return "", ErrNoHosts
	}
	offset := (atomic.AddUint64(&rr.counter, 1) - 1) % uint64(len(hosts))
	return hosts[offset], nil
}

func NewRandomLB(subscriber Subscriber, seed int64) Balancer {
	return &randomLB{
		subscriber: subscriber,
		rnd:        rand.New(rand.NewSource(seed)),
	}
}

type randomLB struct {
	subscriber Subscriber
	rnd        *rand.Rand
}

func (r *randomLB) Host() (string, error) {
	hosts, err := r.subscriber.Hosts()
	if err != nil {
		return "", err
	}
	if len(hosts) <= 0 {
		return "", ErrNoHosts
	}
	return hosts[r.rnd.Intn(len(hosts))], nil
}
