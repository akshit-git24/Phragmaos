package phragmaos


import (
	_"sync"
	"time"
)

func (s *Store) GetOrCreate(inp string,  cfg *EndpointConfig) *token_bucket {

	newBucket := &token_bucket{
		capacity:  cfg.Limit,
		tokens:  cfg.Limit,
		refillRate: cfg.RefillRate,
		lastRefillTime: time.Now(),
	}

	actual, _ := s.buckets.LoadOrStore(inp, newBucket)
	return actual.(*token_bucket)

}