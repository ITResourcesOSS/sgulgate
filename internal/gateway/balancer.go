package gateway

import (
	"math/rand"
)

// RoundRobinStrategy is the round robin balancing strategy name.
const RoundRobinStrategy = "round-robin"

// RandomStrategy is the random balancing strategy name.
const RandomStrategy = "random"

type (
	// Balancer defines the load balancing interface for a concrete balancer.
	Balancer interface {
		Balance(endpoints []string) (int, string)
	}

	roundRobinBalancer struct {
		c int
	}

	randomBalancer struct {
	}
)

// Balancers hold the load balancing strategies.
var balancers = map[string]Balancer{
	RandomStrategy:     &randomBalancer{},
	RoundRobinStrategy: &roundRobinBalancer{c: 0},
}

func (rrb *roundRobinBalancer) Balance(endpoints []string) (int, string) {
	if rrb.c >= len(endpoints) {
		rrb.c = 0
	}
	idx := rrb.c
	rrb.c = rrb.c + 1
	logger.Debugf("RooundRobin balancing to idx: %d, endpoint: %s", idx, endpoints[idx])
	return idx, endpoints[idx]
}

func (rb *randomBalancer) Balance(endpoints []string) (int, string) {
	idx := rand.Intn(len(endpoints))
	logger.Debugf("Random balancing to idx: %d, endpoint: %s", idx, endpoints[idx])
	return idx, endpoints[idx]
}

// RandomBalander returns the random balancer instance.
func RandomBalander() Balancer {
	return balancers[RandomStrategy]
}

// RoundRobinBalancer returns the round-robin balancer instance.
func RoundRobinBalancer() Balancer {
	return balancers[RoundRobinStrategy]
}

// BalancerFor returns the load balancer instance for the requested strategy.
func BalancerFor(strategy string) Balancer {
	return balancers[strategy]
}
