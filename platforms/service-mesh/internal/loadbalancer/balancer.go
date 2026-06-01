package loadbalancer

import (
	"context"
	"crypto/md5"
	"encoding/binary"
	"errors"
	"math/rand"
	"sort"
	"sync"
	"sync/atomic"

	"github.com/tikiclone/tiki/platforms/service-mesh/internal/discovery"
)

type Algorithm string

const (
	RoundRobin       Algorithm = "round_robin"
	LeastConnections Algorithm = "least_connections"
	Random           Algorithm = "random"
	ConsistentHash   Algorithm = "consistent_hash"
)

var ErrNoInstances = errors.New("no available instances")

type LoadBalancer struct {
	algorithm Algorithm
	instances []*discovery.ServiceInstance
	rrCounter atomic.Uint64
	lcCounts  sync.Map
	mu        sync.RWMutex
}

func NewLoadBalancer(alg Algorithm) *LoadBalancer {
	return &LoadBalancer{
		algorithm: alg,
	}
}

func (lb *LoadBalancer) UpdateInstances(instances []*discovery.ServiceInstance) {
	lb.mu.Lock()
	defer lb.mu.Unlock()
	lb.instances = instances
}

func (lb *LoadBalancer) NextInstance(ctx context.Context, sourceIP string) (*discovery.ServiceInstance, error) {
	lb.mu.RLock()
	defer lb.mu.RUnlock()

	if len(lb.instances) == 0 {
		return nil, ErrNoInstances
	}

	switch lb.algorithm {
	case RoundRobin:
		return lb.roundRobin(), nil
	case LeastConnections:
		return lb.leastConnections(), nil
	case Random:
		return lb.random(), nil
	case ConsistentHash:
		return lb.consistentHash(sourceIP), nil
	default:
		return lb.roundRobin(), nil
	}
}

func (lb *LoadBalancer) roundRobin() *discovery.ServiceInstance {
	n := lb.rrCounter.Add(1) - 1
	return lb.instances[n%uint64(len(lb.instances))]
}

func (lb *LoadBalancer) leastConnections() *discovery.ServiceInstance {
	best := lb.instances[0]
	bestCount := int64(1 << 62)
	for _, inst := range lb.instances {
		val, _ := lb.lcCounts.LoadOrStore(inst.ID, int64(0))
		count, ok := val.(int64)
		if !ok {
			count = 0
		}
		if count < bestCount {
			bestCount = count
			best = inst
		}
	}
	lb.lcCounts.Store(best.ID, bestCount+1)
	return best
}

func (lb *LoadBalancer) random() *discovery.ServiceInstance {
	return lb.instances[rand.Intn(len(lb.instances))]
}

func (lb *LoadBalancer) consistentHash(sourceIP string) *discovery.ServiceInstance {
	if sourceIP == "" {
		return lb.instances[0]
	}

	sorted := make([]*discovery.ServiceInstance, len(lb.instances))
	copy(sorted, lb.instances)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].ID < sorted[j].ID
	})

	hash := md5.Sum([]byte(sourceIP))
	hashVal := binary.BigEndian.Uint64(hash[:8])
	return sorted[hashVal%uint64(len(sorted))]
}

func (lb *LoadBalancer) ReleaseConnection(instID string) {
	val, ok := lb.lcCounts.Load(instID)
	if ok {
		count := val.(int64)
		if count > 0 {
			lb.lcCounts.Store(instID, count-1)
		}
	}
}
